package main
import(
	"fmt"
	"log"
	"strconv"
	"net"
	"sort"
	"encoding/gob"
	"time"
)

type LocalSig struct {
	Id 			int64
	Guid		string `sql: "size:40"`
	ServerSecret string `sql: "size:40"`
	CreatedAt	time.Time
}

type Peer struct {
	Id			int64
	Guid		string `sql: "size:40"`
	Token		string `sql: "size:40"`
	Name		string `sql: "size:64"`
	SynchPos	string `sql: "size:40"` // Last changeset applied
	CreatedAt 	time.Time
	UpdatedAt	time.Time
}

type Message struct {
	Type		string
	Param		string
	NoteChg		NoteChange
}

const SYNCH_PORT  string = "8090"

func synch_client(host string, server_secret string) {
	conn, err := net.Dial("tcp", host + ":" + SYNCH_PORT)
	if err != nil {log.Fatal("Error connecting to server ", err)}
	defer conn.Close()
	msg := Message{} // init to empty struct
	enc := gob.NewEncoder(conn)
	dec := gob.NewDecoder(conn)
	defer sendMsg(enc, Message{Type: "Hangup", Param: "", NoteChg: NoteChange{}})

	// Send handshake
	sendMsg(enc, Message{
		Type: "WhoAreYou", Param: whoAmI(), NoteChg: NoteChange{Guid: server_secret},
	})

	rcxMsg(dec, &msg) // Decode the response
	if msg.Type == "WhoIAm" {
		peer_id := msg.Param  // retrieve the server's guid
		println("The server's id is", short_sha(peer_id))
		if len(peer_id) != 40 {
			println("The server's id is invalid. Run the server once with the -setup_db option")
			return
		}
		// Is there a token for us?
		if len(msg.NoteChg.Guid) == 40 {
			setPeerToken(peer_id, msg.NoteChg.Guid)
		}
		peer, err := getPeerByGuid(peer_id)
		if err != nil { println("Error retrieving peer object"); return }
		msg.NoteChg.Guid = ""  // hide the evidence

		// Auth
		msg.Type = "AuthMe"
		msg.Param = peer.Token // This is set for the server(peer) by some access granting mechanism
		                       // which for right now is manual
		sendMsg(enc, msg)
		rcxMsg(dec, &msg)
		if msg.Param != "Authorized" {
			println("The server declined the authorization request")
			return
		}

		// Do we need to Synch?
		if peer.SynchPos != "" { // Do we have a point of last synch with this peer?
			last_change := retrieveLatestChange()
			if last_change.Id > 0 && last_change.Guid == peer.SynchPos {
				// Get server last change
				msg.Type = "LatestChange"
				sendMsg(enc, msg)
				msg = Message{}
				rcxMsg(dec, &msg)
				if msg.NoteChg.Id > 0 && msg.NoteChg.Guid == peer.SynchPos {
					pf("We are already in synch with peer: %s at note_change: %s\n",
						short_sha(peer_id), short_sha(last_change.Guid))
					return
				}
			}
		}
		pf("Last known Synch position is \"%s\"\n", short_sha(peer.SynchPos))

		// Get Server Changes
		sendMsg(enc, Message{Type: "NumberOfChanges", Param: peer.SynchPos}) // heads up on number of changes
		rcxMsg(dec, &msg) // Decode the response
		numChanges, err := strconv.Atoi(msg.Param)
		if err != nil { println("Could not decode the number of change messages"); return }
		println(numChanges, "changes")

		peer_changes := make([]NoteChange, numChanges)
		sendMsg(enc, Message{Type: "SendChanges"})  // send the actual changes
		for i := 0; i < numChanges; i++ {
			msg = Message{}
			rcxMsg(dec, &msg)
			peer_changes[i] = msg.NoteChg
		}
		pf("\n%d server changes received:\n", numChanges)
		
    // Get Local Changes
		local_changes := retrieveLocalNoteChangesFromSynchPoint(peer.SynchPos)
		pf("%d local changes after synch point found\n", len(local_changes))
		// Push local changes to server
		if len(local_changes) > 0 {
			sendMsg(enc, Message{Type: "NumberOfClientChanges",
				Param: strconv.Itoa(len(local_changes))})
			rcxMsg(dec, &msg)
			if msg.Type == "SendChanges" {
				msg.Type = "NoteChange"
				msg.Param = ""
				var note Note
				var note_frag NoteFragment
				for _, change := range (local_changes) {
					note = Note{}
					note_frag = NoteFragment{}
					// We have the change but now we need the NoteFragment or Note depending on the operation type
					if change.Operation == 1 {
						db.Where("id = ?", change.NoteId).First(&note)
						note.Id = 0
						change.Note = note
					}
					if change.Operation == 2 {
						db.Where("id = ?", change.NoteFragmentId).First(&note_frag)
						change.NoteFragment = note_frag
					}
					msg.NoteChg = change
					msg.NoteChg.Print()
					sendMsg(enc, msg)
				}
			}
		}

		// Process remote changes received
		if len(peer_changes) > 0 {
			processChanges(&peer_changes, &local_changes)
		}

		// Mark Synch Point with a special NoteChange; Save on client and server
		if len(peer_changes) > 0 || len(local_changes) > 0 {
			synch_point := generate_sha1()
			synch_nc := NoteChange{Guid: synch_point, Operation: 9}
			db.Save(&synch_nc)
			peer.SynchPos = synch_nc.Guid
			db.Save(&peer)
			// Mark the server with the same NoteChange
			msg.NoteChg = synch_nc
			msg.Type = "NewSynchPoint"
			sendMsg(enc, msg)
		}

	} else {
			println("Peer does not respond to request for database id")
			println("Make sure both server and client databases have been properly setup(migrated) with the -setup_db option")
			println("or make sure peer version is >= 0.9")
			return
    }

	defer println("Synch Operation complete")
}

func processChanges(peer_changes * []NoteChange, local_changes * []NoteChange) {
	sort.Sort(byCreatedAt(*peer_changes)) // we will apply in created order
	var local_change NoteChange
	var skip bool

	for _, peer_change := range(*peer_changes) {
		// If we already have this NoteChange locally then skip // same change
		db.Where("guid = ?", peer_change.Guid).First(&local_change)
		if local_change.Id > 1 { continue } // we already have that NC
		// If there is a newer local change of the same note and field then skip
		for _, local_change = range(*local_changes) {
			if local_change.NoteId == peer_change.NoteId && // same note
					local_change.CreatedAt.After(peer_change.CreatedAt) && ( // local newer
						// any field in peer_change matches a field in local_change
						local_change.NoteFragment.Bitmask&0x8 == peer_change.NoteFragment.Bitmask&0x8 ||
						local_change.NoteFragment.Bitmask&0x4 == peer_change.NoteFragment.Bitmask&0x4 ||
						local_change.NoteFragment.Bitmask&0x2 == peer_change.NoteFragment.Bitmask&0x2 ||
						local_change.NoteFragment.Bitmask&0x1 == peer_change.NoteFragment.Bitmask&0x1 ) {
				skip = true
			}
		}
		if skip {
			skip = false // reset
			continue
		}

		// Apply Changes
		println("____________________________________________________________________")
		performNoteChange(peer_change)
		verifyNoteChangeApplied(peer_change)
	}
}

func retrieveLastChangeForNote(note_guid string) (NoteChange) {
	var noteChange NoteChange
	db.Where("note_guid = ?", note_guid).Order("created_at desc").Limit(1).Find(&noteChange)
	return noteChange
}

func retrieveLatestChange() (NoteChange) {
	var noteChange NoteChange
	db.Order("created_at desc").First(&noteChange)
	return noteChange
}

func performNoteChange(nc NoteChange) bool {
	nc.Print()
	// Get The latest change for the current note in the local changeset
	last_nc := retrieveLastChangeForNote(nc.NoteGuid)

	switch nc.Operation {
	case op_create:
		if last_nc.Id > 0 {
			println("Note - Title", last_nc.Note.Title, "Guid:", short_sha(last_nc.NoteGuid), "already exists locally - cannot create")
			return false
		}
		nc.Note.Id = 0  // Make sure the embedded note object has a zero id for creation
	case op_update:
		note, err := getNote(nc.NoteGuid)
		if err != nil {
			println("Cannot update a non-existent note:", short_sha(nc.NoteGuid))
			return false
		}
		updateNote(note, nc)
		nc.NoteFragment.Id = 0 // Make sure the embedded note_fragment has a zero id for creation
	case op_delete:
		if last_nc.Id < 1 {
			fmt.Printf("Cannot delete a non-existent note (Guid:%s)", short_sha(nc.NoteGuid))
			return false
		} else {
			db.Where("guid = ?", last_nc.NoteGuid).Delete(Note{})
		}
	default:
		return false
	}
	return saveNoteChange(nc)
}

func updateNote(note Note, nc NoteChange) {
	// Update bitmask allowed fields - this allows us to set a field to ""  // Updates are stored as note fragments
	if nc.NoteFragment.Bitmask & 0x8 == 8 {
		note.Title = nc.NoteFragment.Title
	}
	if nc.NoteFragment.Bitmask & 0x4 == 4 {
		note.Description = nc.NoteFragment.Description
	}
	if nc.NoteFragment.Bitmask & 0x2 == 2 {
		note.Body = nc.NoteFragment.Body
	}
	if nc.NoteFragment.Bitmask & 0x1 == 1 {
		note.Tag = nc.NoteFragment.Tag
	}
	db.Save(&note)
}

// Save the change object which will create a Note on CreateOp or a NoteFragment on UpdateOp
func saveNoteChange(nc NoteChange) bool {
	pf("Saving change object...%s\n", short_sha(nc.Guid))
	// Make sure all ids are zeroed - A non-zero Id will not be created
	nc.Id = 0

	db.Create(&nc) // will auto create contained objects too and it's smart - 'nil' children will not be created :-)
	if !db.NewRecord(nc) { // was it saved?
		println("Note change saved:", short_sha(nc.Guid), ", Operation:", nc.Operation)
		return true
	}
	println("Failed to record note changes.", nc.Note.Title, "Changed note Guid:",
			short_sha(nc.NoteGuid), "NoteChange Guid:", short_sha(nc.Guid))
	return false
}

// Get all local NCs later than the synchPoint
func retrieveLocalNoteChangesFromSynchPoint(synch_guid string) ([]NoteChange) {
	var noteChange NoteChange
	var noteChanges []NoteChange

	db.Where("guid = ?", synch_guid).First(&noteChange) // There should be only one
	pf("Synch point note change is: %v\n", noteChange)
	if noteChange.Id < 1 {
		println("Can't find synch point locally - retrieving all note_changes", short_sha(synch_guid))
		db.Find(&noteChanges).Order("created_at, asc")
	} else {
		println("Attempting to retrieve note_changes beyond synch_point")
		db.Find(&noteChanges, "created_at > ?", noteChange.CreatedAt).Order("created_at asc")
	}
	return noteChanges
}

// VERIFICATION

func verifyNoteChangeApplied(nc NoteChange) {
	println("----------------------------------")
	retrievedChange, err := nc.Retrieve()
	if err != nil {
		println("Error retrieving the note change")
	} else if nc.Operation == 1 {
		retrievedNote, err := retrievedChange.RetrieveNote()
        pf("retrievedNote: %s\n", retrievedNote)
		if err != nil {
			println("Error retrieving the note changed")
		} else {
			fmt.Printf("Note created:\n%v\n", retrievedNote)
		}
	} else if nc.Operation == 2 {
		retrievedFrag, err := retrievedChange.RetrieveNoteFrag()
		if err != nil {
			println("Error retrieving the note fragment")
		} else {
			fmt.Printf("Note Fragment created:\n%v\n", retrievedFrag)
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

