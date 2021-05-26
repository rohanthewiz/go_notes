package main

import (
	"encoding/gob"
	"fmt"
	"go_notes/note"
	"log"
	"net"
	"sort"
	"strconv"
	"time"
)

type LocalSig struct {
	Id           int64
	Guid         string `sql:"size:40"`
	ServerSecret string `sql:"size:40"`
	CreatedAt    time.Time
}

// A Peer represents a single client (db)
type Peer struct {
	Id        int64
	Guid      string `sql:"size:40"`
	Token     string `sql:"size:40"`
	User      string // (GUID) has one user //todo - Add Index
	Name      string `sql:"size:64"`
	SynchPos  string `sql:"size:40"` // Last changeset applied
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Message struct {
	Type    string
	Param   string
	NoteChg NoteChange
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
	defer sendMsg(enc, Message{Type: "Hangup", Param: "", NoteChg: NoteChange{}})

	// Send handshake - Client initiates
	sendMsg(enc, Message{
		Type: "WhoAreYou", Param: whoAmI(), NoteChg: NoteChange{Guid: serverSecret}, // borrow NoteChange.Guid
	})

	rcxMsg(dec, &msg) // Decode the response
	if msg.Type == "WhoIAm" {
		peerId := msg.Param // retrieve the server's guid
		pl("The server's id is", shortSHA(peerId))
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
					pf("We are already in synch with peer: %s at note_change: %s\n",
						shortSHA(peerId), shortSHA(lastChange.Guid))
					return
				}
			} // else we probably have never synched so carry on
		}
		pf("Last known Synch position is \"%s\"\n", shortSHA(peer.SynchPos))

		// Get Server Changes
		sendMsg(enc, Message{Type: "NumberOfChanges", Param: peer.SynchPos}) // heads up on number of changes
		rcxMsg(dec, &msg)                                                    // Decode the response
		numChanges, err := strconv.Atoi(msg.Param)
		if err != nil {
			fmt.Println("Could not decode the number of change messages")
			return
		}
		pl(numChanges, "changes")

		peerChanges := make([]NoteChange, numChanges) // preallocate slice (optimization)
		sendMsg(enc, Message{Type: "SendChanges"})    // please send the actual changes
		for i := 0; i < numChanges; i++ {
			msg = Message{}
			rcxMsg(dec, &msg)
			peerChanges[i] = msg.NoteChg
		}
		pf("\n%d server changes received:\n", numChanges)

		// Get Local Changes
		localChanges := retrieveLocalNoteChangesFromSynchPoint(peer.SynchPos)
		pf("%d local changes after synch point found\n", len(localChanges))
		// Push local changes to server
		if len(localChanges) > 0 {
			sendMsg(enc, Message{Type: "NumberOfClientChanges",
				Param: strconv.Itoa(len(localChanges))})
			rcxMsg(dec, &msg)
			if msg.Type == "SendChanges" {
				msg.Type = "NoteChange"
				msg.Param = ""
				var nte note.Note
				var noteFrag NoteFragment
				for _, change := range localChanges {
					nte = note.Note{}
					noteFrag = NoteFragment{}
					// We have the change but now we need the NoteFragment or Note depending on the operation type
					if change.Operation == 1 {
						db.Where("id = ?", change.NoteId).First(&nte)
						nte.Id = 0
						change.Note = nte
					}
					if change.Operation == 2 {
						db.Where("id = ?", change.NoteFragmentId).First(&noteFrag)
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
			synch_nc := NoteChange{Guid: generateSHA1(), Operation: 9}
			db.Save(&synch_nc)
			peer.SynchPos = synch_nc.Guid
			db.Save(&peer)
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

func processChanges(peer_changes *[]NoteChange, local_changes *[]NoteChange) {
	pl("Processing received changes...")
	sort.Sort(byCreatedAt(*peer_changes)) // we will apply in created order
	var localChange NoteChange
	var skip bool

	for _, peerChange := range *peer_changes {
		// If we already have this NoteChange locally then skip // same change
		localChange = NoteChange{} // make sure local_change is inited here
		// otherwise GORM uses its id in the query - weird!
		db.Where("guid = ?", peerChange.Guid).First(&localChange)
		if localChange.Id > 1 {
			pf("We already have NoteChange: %s -- skipping\n", shortSHA(localChange.Guid))
			continue // we already have that NC
		}
		// If there is a newer local change of the same note and field then skip
		for _, localChange = range *local_changes {
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
		pl("____________________APPLYING CHANGE_________________________________")
		performNoteChange(peerChange)
		verifyNoteChangeApplied(peerChange)
	}
}

func retrieveLastChangeForNote(note_guid string) NoteChange {
	var noteChange NoteChange
	db.Where("note_guid = ?", note_guid).Order("created_at desc").Limit(1).Find(&noteChange)
	return noteChange
}

func retrieveLatestChange() NoteChange {
	var noteChange NoteChange
	db.Order("created_at desc").First(&noteChange)
	return noteChange
}

// Create, Update or Delete a note, while tracking the change
func performNoteChange(nc NoteChange) bool {
	nc.Print()
	// Get The latest change for the current note in the local changeset
	lastNC := retrieveLastChangeForNote(nc.NoteGuid)

	switch nc.Operation {
	case op_create:
		if lastNC.Id > 0 {
			fmt.Println("Note - Title", lastNC.Note.Title, "Guid:", shortSHA(lastNC.NoteGuid), "already exists locally - cannot create")
			return false
		}
		nc.Note.Id = 0 // Make sure the embedded note object has a zero id for creation
	case op_update:
		nte, err := getNote(nc.NoteGuid)
		if err != nil {
			fmt.Println("Cannot update a non-existent note:", shortSHA(nc.NoteGuid))
			return false
		}
		updateNote(nte, nc)
		nc.NoteFragment.Id = 0 // Make sure the embedded note_fragment has a zero id for creation
	case op_delete:
		if lastNC.Id < 1 {
			fmt.Printf("Cannot delete a non-existent note (Guid:%s)", shortSHA(nc.NoteGuid))
			return false
		} else {
			db.Where("guid = ?", lastNC.NoteGuid).Delete(note.Note{})
		}
	default:
		return false
	}
	return saveNoteChange(nc)
}

func updateNote(n note.Note, nc NoteChange) {
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
	db.Save(&n)
}

// Save the change object which will create a Note on CreateOp or a NoteFragment on UpdateOp
func saveNoteChange(nc NoteChange) bool {
	pf("Saving change object...%s\n", shortSHA(nc.Guid))
	// Make sure all ids are zeroed - A non-zero Id will not be created
	nc.Id = 0

	db.Create(&nc)         // will auto create contained objects too and it's smart - 'nil' children will not be created :-)
	if !db.NewRecord(nc) { // was it saved?
		pl("Note change saved:", shortSHA(nc.Guid), ", Operation:", nc.Operation)
		return true
	}
	fmt.Println("Failed to record note changes.", nc.Note.Title, "Changed note Guid:",
		shortSHA(nc.NoteGuid), "NoteChange Guid:", shortSHA(nc.Guid))
	return false
}

// Get all local NCs later than the synchPoint
func retrieveLocalNoteChangesFromSynchPoint(synch_guid string) []NoteChange {
	var noteChange NoteChange
	var noteChanges []NoteChange

	db.Where("guid = ?", synch_guid).First(&noteChange) // There should be only one
	pf("Synch point note change is: %v\n", noteChange)
	if noteChange.Id < 1 {
		pl("Can't find synch point locally - retrieving all note_changes", shortSHA(synch_guid))
		db.Find(&noteChanges).Order("created_at, asc")
	} else {
		pl("Attempting to retrieve note_changes beyond synch_point")
		db.Find(&noteChanges, "created_at > ?", noteChange.CreatedAt).Order("created_at asc")
	}
	return noteChanges
}

// VERIFICATION

func verifyNoteChangeApplied(nc NoteChange) {
	pl("----------------------------------")
	retrievedChange, err := nc.Retrieve()
	if err != nil {
		fmt.Println("Error retrieving the note change")
	} else if nc.Operation == 1 {
		retrievedNote, err := retrievedChange.RetrieveNote()
		pf("retrievedNote: %s\n", retrievedNote)
		if err != nil {
			fmt.Println("Error retrieving the note changed")
		} else {
			pf("Note created:\n%v\n", retrievedNote)
		}
	} else if nc.Operation == 2 {
		retrievedFrag, err := retrievedChange.RetrieveNoteFrag()
		if err != nil {
			fmt.Println("Error retrieving the note fragment")
		} else {
			pf("Note Fragment created:\n%v\n", retrievedFrag)
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
