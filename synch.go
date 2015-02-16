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
	defer sendMsg(enc, Message{Type: "Hangup", Param: "", NoteChg: NoteChange{}})

	// Send handshake
	sendMsg(enc, Message{Type: "WhoAreYou"})
	rcxMsg(dec, &msg) // Decode the response
	if msg.Type == "WhoIAm" {
		peer_id := msg.Param
		println("The server's id is", short_sha(peer_id))

		//Do we have a point of last synchronization with this peer?
		var synch_point string  // defaults to empty
		var peer Peer
		db.Where("guid = ?", peer_id).First(&peer)
		if peer.Id > 0 && peer.SynchPos != "" {
			// If we are already in synch, abort
			last_change := retrieveLatestChange()
			if last_change.Id > 0 && last_change.Guid == peer.SynchPos {
				synch_point = peer.SynchPos // otherwize synch_point will be ""
			}
		}
		pf("Last known Synch position is \"%s\"\n", short_sha(synch_point))
		sendMsg(enc, Message{Type: "NumberOfChanges", Param: synch_point})
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


		//PROCESS CHANGES

		sort.Sort(byCreatedAt(peer_changes)) // we will apply in created order

		for _, change := range(peer_changes) {
			println("____________________________________________________________________")
			// If we already have this NoteChange locally then skip
			var local_change NoteChange
			db.Where("guid = ?", change.Guid).First(&local_change)
			if local_change.Id > 1 { continue } // we already have that NC
			// Apply Changes
			performNoteChange(change)
			verifyNoteChangeApplied(change)
			// When done push this note Guid to the completed array
		}

		// Save the last synch point - //TODO - How does success above influence the last synch point?
		var lastSynchPoint string
		if ln := len(peer_changes); ln > 0 {
			lastSynchPoint = peer_changes[ln - 1].Guid
		}
		if peer.Id > 0 {
			peer.SynchPos = lastSynchPoint
			db.Save(&peer)
		} else {
			db.Create(&Peer{Guid: peer_id, SynchPos: lastSynchPoint})
		}
		db.Where("guid = ?", peer_id).First(&peer)
		if peer.SynchPos == lastSynchPoint {
			println("Peer Synch Point saved:", short_sha(lastSynchPoint))
		} else {
			println("Warning! Could not save a synch point for peer:", short_sha(peer_id),
					"Future synchs with this peer may be unreliable")
		}

	} else {
        println("Peer does not respond to request for database id\nRun peer with -setup_db option or make sure peer version is >= 0.9")
		return
    }

	println("Synch Operation complete")
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
	fmt.Printf("Operation: %d, Title: %s, Guid: %s, NoteGuid: %s, CreatedAt: %s\n",
			nc.Operation, nc.Note.Title, short_sha(nc.Guid), short_sha(nc.NoteGuid), nc.CreatedAt)
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
	fmt.Printf("Saving change object...\n%s", short_sha(nc.Guid))
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

// Currently unused // Get all local NCs later than the synchPoint
func retrieveLocalNoteChangesFromSynchPoint(synch_guid string) ([]NoteChange) {
	var noteChange NoteChange
	var noteChanges []NoteChange

	db.Where("guid = ?", synch_guid).First(&noteChange) // There should be only one
	if noteChange.Id < 1 {
		println("Can't find synch point locally", short_sha(synch_guid))
		db.Find(&noteChanges).Order("created_at, asc") // Can't find the synch point so send them all
	} else {
		db.Where("created_at > '" + noteChange.CreatedAt.String() + "'").Find(&noteChanges).Order("created_at asc")
	}
	return noteChanges
}

// VERIFICATION

func verifyNoteChangeApplied(nc NoteChange) {
	// Verify
	println("----------------------------------")
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
		return NoteChange{}, errors.New("NoteChange not found")
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

func printNoteChange(nc NoteChange) {
	pf("NoteChange: {Id: %d, Guid: %s, NoteGuid: %s, Oper: %d\nNote: {Id: %d, Guid: %s, Title: %s}\nNoteFragment: {Id: %d, Bitmask: %d, Title: %s, Description: %s, Body: %s, Tag: %s}}\n",
		nc.Id, short_sha(nc.Guid), short_sha(nc.NoteGuid), nc.Operation, nc.NoteId, short_sha(nc.Note.Guid), nc.Note.Title,
		nc.NoteFragment.Id, nc.NoteFragment.Bitmask, nc.NoteFragment.Title, nc.NoteFragment.Description, nc.NoteFragment.Body, nc.NoteFragment.Tag,
	)
}
