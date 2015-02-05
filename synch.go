package main
import(
	"fmt"
	"log"
	"strconv"
	"net"
	"sort"
	"encoding/gob"
	"errors"
	//"time"
)

func synch_client(host string) {
	conn, err := net.Dial("tcp", host + ":" + SYNCH_PORT)
	if err != nil {log.Fatal("Error connecting to server ", err)}
	defer conn.Close()
	msg := Message{} // init to empty struct
	enc := gob.NewEncoder(conn)
	dec := gob.NewDecoder(conn)
    
	// Send handshake
	sendMsg(enc, Message{Type: "WhoAreYou"})
	rcxMsg(dec, &msg) // Decode the response
	if msg.Type == "WhoIAm" {
		peer_id := msg.Param
		println("The server's database id is", peer_id)

		//Do we have a point of last synchronization with this peer?
		var synch_point string
		var peer Peer
		db.Where("guid = ?", peer_id).First(&peer)
		if peer.Id != 0 && peer.SynchPos != "" {
			synch_point = peer.SynchPos
		} else {
			synch_point = ""
		}
		println("Synch position is ", synch_point)

		//TODO: Are we in Sych?
		// If your latest change == my latest change then we are in synch
		synch_point = ""  // hardwire for now // Todo set after synch

		sendMsg(enc, Message{Type: "NumberOfChanges"})
		rcxMsg(dec, &msg) // Decode the response
		numChanges, err := strconv.Atoi(msg.Param)
		if err != nil { println("Could not decode the number of change messages"); return }
		println(numChanges, "changes")

		peer_changes := make([]NoteChange, numChanges)
		sendMsg(enc, Message{Type: "SendChanges"})
		for i := 0; i < numChanges; i++ {
			msg = Message{}
			rcxMsg(dec, &msg)
			peer_changes[i] = msg.NoteChg
		}
		pf("\n%d peer changes received:\n", numChanges)

		println("Disconnecting from Server")
		sendMsg(enc, Message{Type: "Hangup", Param: "", NoteChg: NoteChange{}})

		//PROCESS CHANGES

		sort.Sort(byCreatedAt(peer_changes)) // we will apply in created order
		var note_guid string

		for _, change := range(peer_changes) {
			note_guid = change.NoteGuid

			// Get The latest change for the current note in the local changeset
			last_note_change, err := retrieveLastNoteChangeForNote(note_guid)
			if err != nil {
				pf("This is a new incoming note: %s\n", note_guid[:9])
				//This will be simple synch operation and must be a create,
				// Or, If latest change is a delete op - skip synching this note
				// Weigh the above two points

			} else {
				fmt.Printf("Last local NoteChange for current note: %v\n", last_note_change)
				//Todo - consider this condition in the synch
					// - operation cannot be a create
					// - but could be an update or delete
			}

			// Apply Changes
				performNoteChange(change)
				verifyNoteChangeApplied(change)
			// When done push this note Guid to the completed array
		}

	} else {
        println("Peer does not respond to request for database id\nRun peer with -setup_db option or make sure peer version is >= 0.9")
		return
    }

	println("Synch Operation complete")
}

// Return last noteChange For a note
func retrieveLastNoteChangeForNote(note_guid string) (NoteChange, error) {
	var noteChange NoteChange
	db.Where("note_guid = ?", note_guid).Order("created_at desc").Limit(1).Find(&noteChange)
	if noteChange.Id == 0 {
		return NoteChange{}, errors.New("Note not found")
	}
	return noteChange, nil
}

// Get all local NCs later than the synchPoint
func retrieveLocalNoteChangesFromSynchPoint(synch_guid string) ([]NoteChange) {
	var noteChange NoteChange
	var noteChanges []NoteChange

	db.Where("guid = ?", synch_guid).First(&noteChange) // There should be only one
	if noteChange.Id == 0 {
		db.Find(&noteChanges).Order("created_at, asc")
	} else {
		db.Where("created_at > " + noteChange.CreatedAt.String()).Find(&noteChanges).Order("created_at, asc")
	}
	return noteChanges
}

func performNoteChange(nc NoteChange) bool {
	//fmt.Printf("Applying NoteChange: %v\n", nc)
	fmt.Printf("Operation: %d, Title: %s, Guid: %s, NoteGuid: %s, CreatedAt: %s\n",
		nc.Operation, nc.Note.Title, nc.Guid[:9], nc.NoteGuid[:9], nc.CreatedAt)
	switch nc.Operation {

	case op_create:
		if _, err := getNote(nc.NoteGuid); err == nil {
			println("Note - Title", nc.Note.Title, "Guid:", nc.NoteGuid[:9], "already exists locally")
			return false
		}
		return saveNoteChange(nc)

	case op_update:
		note, err := getNote(nc.NoteGuid)
		if err != nil {
			println("Cannot update a non-existent note:", nc.NoteGuid[:9])
			return false
		}
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
//		fmt.Printf("NoteFragment.Bitmask: %v", nc.NoteFragment.Bitmask)
//		fmt.Printf("NoteFragment.Bitmask & 0x8: %v", nc.NoteFragment.Bitmask & 0x8)
		db.Save(&note)
		return saveNoteChange(nc)

	case op_delete:
		if note, err := getNote(nc.NoteGuid); err != nil {
			return false
		} else {
			db.Delete(&note)
		}

	default:
		return false
	}

	return true
}

// Save the change object which will create a Note on CreateOp or a NoteFragment on UpdateOp
func saveNoteChange(note_change NoteChange) bool {
	fmt.Printf("Saving change object...\n%v", note_change) // TODO - remove this for production
	db.Create(&note_change) // will auto create contained objects too and it's smart - 'nil' children will not be created :-)
	if !db.NewRecord(note_change) { // was it saved?
		println("Note change saved:", note_change.Guid, ", Operation:", note_change.Operation)
		return true
	}
	println("Failed to record note changes.", note_change.Note.Title, "Changed note Guid:",
			note_change.NoteGuid[:9], "NoteChange Guid:", note_change.Guid[:9])
	return false
}

func getNote(guid string) (Note, error) {
	var note Note
	db.Where("guid = ?", guid).First(&note)
	if note.Id != 0 {
		return note, nil
	} else {
		return note, errors.New("Note not found")
	}
}

func verifyNoteChangeApplied(nc NoteChange) {
	// Verify
	retrievedChange, err := retrieveNoteChangeByObject(nc)
	if err != nil {
		println("Error retrieving the note change")
	} else if nc.Operation == 1 {
		retrievedNote, err := retrieveChangedNote(retrievedChange)
		if err != nil {
			println("Error retrieving the note changed")
		} else {
			fmt.Printf("Note created:\n%v\n", retrievedNote)
		}
	} else if nc.Operation == 2 {
		retrievedFrag, err := retrieveNoteFrag(retrievedChange)
		if err != nil {
			println("Error retrieving the note fragment")
		} else {
			fmt.Printf("Note Fragment created:\n%v\n", retrievedFrag)
		}
	}
}

func retrieveNoteChangeByObject(nc NoteChange) (NoteChange, error) {
	var noteChanges []NoteChange
	db.Where("guid = ?", nc.Guid).Limit(1).Find(&noteChanges)
	if len(noteChanges) == 1 {
		return noteChanges[0], nil
	} else {
		return NoteChange{}, errors.New("Note not found")
	}
}

func retrieveChangedNote(nc NoteChange) (Note, error) {
	var note Note
	db.Model(&nc).Related(&note)
	if note.Id > 0 {
		return note, nil
	} else {
		return Note{}, errors.New("Note not found")
	}
}

func retrieveNoteFrag(nc NoteChange) (NoteFragment, error) {
	var noteFrag NoteFragment
	db.Model(&nc).Related(&noteFrag)
	if noteFrag.Id > 0 {
		return noteFrag, nil
	} else {
		return NoteFragment{}, errors.New("NoteFragment not found")
	}
}
