package main

import (
	"encoding/gob"
	"fmt"
	"go_notes/dbhandle"
	"go_notes/note"
	"go_notes/note/note_change"
	"go_notes/utils"
	"log"
	"net"
	"sort"
	"strconv"
)

type Message struct {
	Type    string
	Param   string
	NoteChg note_change.NoteChange
}

const SynchPort string = "8090"

func synchClient(host string, serverSecret string) {
	conn, err := net.Dial("tcp", host+":"+SynchPort)
	if err != nil {
		log.Fatal("Error connecting to server ", err)
	}
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			log.Println("Failed to close connection", err)
		}
	}(conn)
	msg := Message{} // init to empty struct
	enc := gob.NewEncoder(conn)
	dec := gob.NewDecoder(conn)
	defer sendMsg(enc, Message{Type: "Hangup", Param: "", NoteChg: note_change.NoteChange{}})

	// Send handshake - Client initiates
	sendMsg(enc, Message{
		Type: "WhoAreYou", Param: whoAmI(), NoteChg: note_change.NoteChange{Guid: serverSecret}, // borrow NoteChange.Guid
	})

	rcxMsg(dec, &msg) // Decode the response
	if msg.Type == "WhoIAm" {
		peerId := msg.Param // retrieve the server's guid
		utils.Pl("The server's id is", utils.ShortSHA(peerId))
		if len(peerId) != 40 {
			fmt.Println("The server's id is invalid. Run the server once with the -setup_db option")
			return
		}
		// Is there a token for us?
		if len(msg.NoteChg.Guid) == 40 {
			err := setPeerToken(peerId, msg.NoteChg.Guid)
			if err != nil {
				log.Println("Failed to set peer token:", err)
			} // make sure to save new auth token
		}
		// Obtain the peer object which represents the server
		peer, err := getPeerByGuid(peerId)
		if err != nil {
			fmt.Println("Error retrieving peer object")
			return
		}
		msg.NoteChg.Guid = "" // hide the evidence

		// Auth
		msg.Type = "AuthMe"
		msg.Param = peer.Token // This is set for the server(peer) by some access granting mechanism
		sendMsg(enc, msg)
		rcxMsg(dec, &msg)
		if msg.Param != "Authorized" {
			fmt.Println("The server declined the authorization request")
			return
		}

		// Do we need to Synch?
		// (SynchPos is the NoteChg.Guid of the last change applied in a synch operation)
		if peer.SynchPos != "" { // Do we have a point of last synch with this peer?
			lastChange := retrieveLatestChange() // Retrieve last local Note Change
			if lastChange.Id > 0 && lastChange.Guid == peer.SynchPos {
				// Get server last change
				msg.Type = "LatestChange"
				sendMsg(enc, msg)
				msg = Message{}
				rcxMsg(dec, &msg)
				if msg.NoteChg.Id > 0 && msg.NoteChg.Guid == peer.SynchPos {
					utils.Pf("We are already in synch with peer: %s at note_change: %s\n",
						utils.ShortSHA(peerId), utils.ShortSHA(lastChange.Guid))
					return
				}
			} // else we probably have never synched so carry on
		}
		utils.Pf("Last known Synch position is \"%s\"\n", utils.ShortSHA(peer.SynchPos))

		// Get Server Changes
		sendMsg(enc, Message{Type: "NumberOfChanges", Param: peer.SynchPos}) // heads up on number of changes
		rcxMsg(dec, &msg)                                                    // Decode the response
		numChanges, err := strconv.Atoi(msg.Param)
		if err != nil {
			fmt.Println("Could not decode the number of change messages")
			return
		}
		utils.Pl(numChanges, "changes")

		peerChanges := make([]note_change.NoteChange, numChanges) // preallocate slice (optimization)
		sendMsg(enc, Message{Type: "SendChanges"})                // please send the actual changes
		for i := 0; i < numChanges; i++ {
			msg = Message{}
			rcxMsg(dec, &msg)
			peerChanges[i] = msg.NoteChg
		}
		utils.Pf("\n%d server changes received:\n", numChanges)

		// Get Local Changes
		localChanges := retrieveLocalNoteChangesFromSynchPoint(peer.SynchPos)
		utils.Pf("%d local changes after synch point found\n", len(localChanges))
		// Push local changes to server
		if len(localChanges) > 0 {
			sendMsg(enc, Message{Type: "NumberOfClientChanges",
				Param: strconv.Itoa(len(localChanges))})
			rcxMsg(dec, &msg)
			if msg.Type == "SendChanges" {
				msg.Type = "NoteChange"
				msg.Param = ""
				var nte note.Note
				var noteFrag note_change.NoteFragment
				for _, change := range localChanges {
					nte = note.Note{}
					noteFrag = note_change.NoteFragment{}
					// We have the change but now we need the NoteFragment or Note depending on the operation type
					if change.Operation == 1 {
						dbhandle.DB.Where("id = ?", change.NoteId).First(&nte)
						nte.Id = 0
						change.Note = nte
					}
					if change.Operation == 2 {
						dbhandle.DB.Where("id = ?", change.NoteFragmentId).First(&noteFrag)
						change.NoteFragment = noteFrag
					}
					msg.NoteChg = change
					msg.NoteChg.Print()
					sendMsg(enc, msg)
				}
			}
		}

		// Process remote changes received
		if len(peerChanges) > 0 {
			processChanges(&peerChanges, &localChanges)
		}

		// Mark Synch Point with a special NoteChange (Operation: 9)
		// Save on client and server
		if len(peerChanges) > 0 || len(localChanges) > 0 {
			synch_nc := note_change.NoteChange{Guid: utils.GenerateSHA1(), Operation: 9}
			dbhandle.DB.Save(&synch_nc)
			peer.SynchPos = synch_nc.Guid
			dbhandle.DB.Save(&peer)
			// Mark the server with the same NoteChange
			msg.NoteChg = synch_nc
			msg.Type = "NewSynchPoint"
			sendMsg(enc, msg)
		}

	} else {
		fmt.Println("Peer does not respond to request for database id")
		fmt.Println("Make sure both server and client databases have been properly setup(migrated) with the -setup_db option")
		fmt.Println("or make sure peer version is >= 0.9")
		return
	}

	defer fmt.Println("Synch Operation complete")
}

func processChanges(peerChanges *[]note_change.NoteChange, localChanges *[]note_change.NoteChange) {
	utils.Pl("Processing received changes...")
	sort.Sort(note_change.ByCreatedAt(*peerChanges)) // we will apply in created order
	var localChange note_change.NoteChange
	var skip bool

	for _, peerChange := range *peerChanges {
		// If we already have this NoteChange locally then skip // same change
		localChange = note_change.NoteChange{} // make sure local_change is inited here
		// otherwise GORM uses its id in the query - weird!
		dbhandle.DB.Where("guid = ?", peerChange.Guid).First(&localChange)
		if localChange.Id > 1 {
			utils.Pf("We already have NoteChange: %s -- skipping\n", utils.ShortSHA(localChange.Guid))
			continue // we already have that NC
		}
		// If there is a newer local change of the same note and field then skip
		for _, localChange = range *localChanges {
			if localChange.NoteGuid == peerChange.NoteGuid && // same note
				localChange.CreatedAt.After(peerChange.CreatedAt) && ( // local newer
			// any field in peer_change matches a field in local_change
			localChange.NoteFragment.Bitmask&0x8 == peerChange.NoteFragment.Bitmask&0x8 ||
				localChange.NoteFragment.Bitmask&0x4 == peerChange.NoteFragment.Bitmask&0x4 ||
				localChange.NoteFragment.Bitmask&0x2 == peerChange.NoteFragment.Bitmask&0x2 ||
				localChange.NoteFragment.Bitmask&0x1 == peerChange.NoteFragment.Bitmask&0x1) {
				skip = true
			}
		}
		if skip {
			skip = false // reset
			continue
		}

		// Apply Changes
		utils.Pl("____________________APPLYING CHANGE_________________________________")
		performNoteChange(peerChange)
		verifyNoteChangeApplied(peerChange)
	}
}

func retrieveLastChangeForNote(note_guid string) note_change.NoteChange {
	var noteChange note_change.NoteChange
	dbhandle.DB.Where("note_guid = ?", note_guid).Order("created_at desc").Limit(1).Find(&noteChange)
	return noteChange
}

func retrieveLatestChange() note_change.NoteChange {
	var noteChange note_change.NoteChange
	dbhandle.DB.Order("created_at desc").First(&noteChange)
	return noteChange
}

// Create, Update or Delete a note, while tracking the change
func performNoteChange(nc note_change.NoteChange) bool {
	nc.Print()
	// Get The latest change for the current note in the local changeset
	lastNC := retrieveLastChangeForNote(nc.NoteGuid)

	switch nc.Operation {
	case note_change.OpCreate:
		if lastNC.Id > 0 {
			fmt.Println("Note - Title", lastNC.Note.Title, "Guid:", utils.ShortSHA(lastNC.NoteGuid), "already exists locally - cannot create")
			return false
		}
		nc.Note.Id = 0 // Make sure the embedded note object has a zero id for creation
	case note_change.OpUpdate:
		nte, err := getNote(nc.NoteGuid)
		if err != nil {
			fmt.Println("Cannot update a non-existent note:", utils.ShortSHA(nc.NoteGuid))
			return false
		}
		updateNote(nte, nc)
		nc.NoteFragment.Id = 0 // Make sure the embedded note_fragment has a zero id for creation
	case note_change.OpDelete:
		if lastNC.Id < 1 {
			fmt.Printf("Cannot delete a non-existent note (Guid:%s)", utils.ShortSHA(nc.NoteGuid))
			return false
		} else {
			dbhandle.DB.Where("guid = ?", lastNC.NoteGuid).Delete(note.Note{})
		}
	default:
		return false
	}
	return saveNoteChange(nc)
}

func updateNote(n note.Note, nc note_change.NoteChange) {
	// Update bitmask allowed fields - this allows us to set a field to ""  // Updates are stored as note fragments
	if nc.NoteFragment.Bitmask&0x8 == 8 {
		n.Title = nc.NoteFragment.Title
	}
	if nc.NoteFragment.Bitmask&0x4 == 4 {
		n.Description = nc.NoteFragment.Description
	}
	if nc.NoteFragment.Bitmask&0x2 == 2 {
		n.Body = nc.NoteFragment.Body
	}
	if nc.NoteFragment.Bitmask&0x1 == 1 {
		n.Tag = nc.NoteFragment.Tag
	}
	dbhandle.DB.Save(&n)
}

// Save the change object which will create a Note on CreateOp or a NoteFragment on UpdateOp
func saveNoteChange(nc note_change.NoteChange) bool {
	utils.Pf("Saving change object...%s\n", utils.ShortSHA(nc.Guid))
	// Make sure all ids are zeroed - A non-zero Id will not be created
	nc.Id = 0

	dbhandle.DB.Create(&nc)         // will auto create contained objects too and it's smart - 'nil' children will not be created :-)
	if !dbhandle.DB.NewRecord(nc) { // was it saved?
		utils.Pl("Note change saved:", utils.ShortSHA(nc.Guid), ", Operation:", nc.Operation)
		return true
	}
	fmt.Println("Failed to record note changes.", nc.Note.Title, "Changed note Guid:",
		utils.ShortSHA(nc.NoteGuid), "NoteChange Guid:", utils.ShortSHA(nc.Guid))
	return false
}

// Get all local NCs later than the synchPoint
func retrieveLocalNoteChangesFromSynchPoint(synch_guid string) []note_change.NoteChange {
	var noteChange note_change.NoteChange
	var noteChanges []note_change.NoteChange

	dbhandle.DB.Where("guid = ?", synch_guid).First(&noteChange) // There should be only one
	utils.Pf("Synch point note change is: %v\n", noteChange)
	if noteChange.Id < 1 {
		utils.Pl("Can't find synch point locally - retrieving all note_changes", utils.ShortSHA(synch_guid))
		dbhandle.DB.Find(&noteChanges).Order("created_at, asc")
	} else {
		utils.Pl("Attempting to retrieve note_changes beyond synch_point")
		dbhandle.DB.Find(&noteChanges, "created_at > ?", noteChange.CreatedAt).Order("created_at asc")
	}
	return noteChanges
}

// VERIFICATION

func verifyNoteChangeApplied(nc note_change.NoteChange) {
	utils.Pl("----------------------------------")
	retrievedChange, err := nc.Retrieve()
	if err != nil {
		fmt.Println("Error retrieving the note change")
	} else if nc.Operation == 1 {
		retrievedNote, err := retrievedChange.RetrieveNote()
		utils.Pf("retrievedNote: %s\n", retrievedNote)
		if err != nil {
			fmt.Println("Error retrieving the note changed")
		} else {
			utils.Pf("Note created:\n%v\n", utils.TruncString(retrievedNote.Guid, 12), retrievedNote.Title)
		}
	} else if nc.Operation == 2 {
		retrievedFrag, err := retrievedChange.RetrieveNoteFrag()
		if err != nil {
			fmt.Println("Error retrieving the note fragment")
		} else {
			utils.Pf("Note Fragment created:\n%v\n", retrievedFrag.Id, retrievedFrag.Bitmask, "-", retrievedFrag.Title)
		}
	}
}

/*	We did not follow some of the below
    Synch philosophy - From the Perspective of the Client

    	- Have we met before? - Do I have you as a peer stored in my peer DB?
    		- Yes: Pull up the last synch point - Changeset from our Peer db
    			(the server should have a matching synch point for us)
    		- Else: Synch point will be 0 index of changesets sorted by created at ASC
    	- Are we in synch? - Does your latest change = my latest change?
    		Else let's synch

			Actual synching
			- From the synch point
				- Get all changesets from both sides more recent than the synch point
				- mark each change with a boolean 'local'
				- store in arr_unsynched_changes --> arr_uc
				- Sort by note guid, then by date asc
					- apply by desired algorithm
						(Thoughts)
						- apply by created_at asc?
						- don't apply changes I don't own
						- maintain the guid of the change, but it is recreated so new created_at on applyee
						(More Thoughts)
						- Apply update changes in date order however
							- Delete - Ends applying of changesets for that note
							- Create cannot follow Create or Update
							- In the DB make sure GUIDs are unique - so shouldn't have to check for create, update

			- We need to save our current synch point
				- sort the unsynched changes array by created_at
				- save the latest changeset guid in our peer db and the same guid in server's peer db
*/
