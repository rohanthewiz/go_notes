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
    
	// Send a message
	sendMsg(enc, Message{Type: "WhoAreYou"})
	rcxMsg(dec, &msg) // Decode the response

    if msg.Type == "WhoIAm" {
        peer_id := msg.Param
        println("The server's database id is", peer_id)

		var synch_point string
		//Do we have a point of last synchronization with this peer?
		var peers []Peer
		db.Where("guid = ?", peer_id).Limit(1).Find(&peers)
		if len(peers) == 1 && peers[0].SynchPos != "" {
			synch_point = peers[0].SynchPos
		} else {
			synch_point = ""
		}
		println("Synch position is ", synch_point)

		//TODO: Are we in Sych?
		// If your latest change == my latest change then we are in synch

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

		sort.Sort(byCreatedAt(peer_changes)) // sort changes
		// Apply Changes
		fmt.Printf("\n%d peer changes received:\n", numChanges)
		for _, item := range(peer_changes) {
			applyChange(item)
		}

	} else {
        println("Peer does not respond to request for database id\nRun peer with -setup_db option or make sure peer version is >= 0.9")
		return
    }

	// Send Hangup
	sendMsg(enc, Message{Type: "Hangup", Param: "", NoteChg: NoteChange{}})
	println("Client done")
}

func applyChange(nc NoteChange) bool {
	fmt.Printf("Title: %s, Operation: %d, CreatedAt: %s, Guid: %s\n", nc.Note.Title, nc.Operation, nc.CreatedAt, nc.Guid)
	switch nc.Operation {
	case op_create:
		if _, err := getNote(nc.Guid); err == nil {
			println("Note - Title", nc.Note.Title, "Guid:", nc.Note.Guid, "already exists locally")
			return false
		}
	case op_update:
		note, err := getNote(nc.NoteFragment.Guid)
		if err != nil {
			println("Cannot update a non-existent note:", nc.Guid)
			return false
		}
		// shouldn't need this // nc.Note = Note{} // Force no create
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
		fmt.Printf("NoteFragment.Bitmask: %v", nc.NoteFragment.Bitmask)
		fmt.Printf("NoteFragment.Bitmask & 0x8: %v", nc.NoteFragment.Bitmask & 0x8)
		fmt.Printf("Updated note: %v", note)
		db.Save(&note)
	case op_delete:
		if _, err := getNote(nc.Guid); err != nil {
			return false
		} else { deleteFromNoteChange(nc.Guid) }
	default:
		return false
	}
	createNoteChange(nc) // Save the change locally
	return true
}

func getNote(guid string) (Note, error) {
	var notes []Note
	db.Where("guid = ?", guid).Limit(1).Find(&notes)
	if len(notes) == 1 && notes[0].Guid == guid {
		return notes[0], nil
	} else {
		return Note{}, errors.New("Note not found")
	}
}

func createFromNoteChange(nc NoteChange) bool {
	println("Creating new note from NoteChange", nc.Note.Title, "-", nc.Guid)
//	nc.Note = Note{}
//	note := nc.Note // Note{Guid: nc.Guid, Title: nc.Title, Description: nc.Description, Body: nc.Body, Tag: nc.Tag}
	// db.Create(&note)
	createNoteChange(nc)
//	if !db.NewRecord(note) { // was it saved?
//		println("Record saved:", note.Title)
		// Save the NoteChange on this machine. Maintain the same Guid
		//nc.CreatedAt = time.Now()  // Not necessary // Todo verify
		//createNoteChange(NoteChange{Guid: nc.Guid, Operation: op_create, Title: opts_str["t"], Description: opts_str["d"], Body: opts_str["b"], Tag: opts_str["g"]} )
//		return true
//	}
//	println("Failed to save:", note.Title)
	return true
//	return false
}

func updateFromNoteChange(nc NoteChange, note Note) {
	println("We would update the note", note.Title)
}

func deleteFromNoteChange(guid string) {
	println("We would delete the note with Guid", guid)
}
