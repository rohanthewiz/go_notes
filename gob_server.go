// GOB client server
package main
import(
	//"reflect"
	"os"
	"net"
	"encoding/gob"
	"fmt"
	//"time"
	"strconv"
)

func synch_server() { // WIP
	ln, err := net.Listen("tcp", ":" + SYNCH_PORT) // counterpart of net.Dial
	if err != nil {
		println("Error setting up server listen on port", SYNCH_PORT)
		return
	}
	fmt.Println("Server listening on port: " + SYNCH_PORT + " - CTRL-C to quit")

	for {
		conn, err := ln.Accept() // this blocks until connection or error
		if err != nil { continue }
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	msg := Message{}
	defer conn.Close()
	enc := gob.NewEncoder(conn)
	dec := gob.NewDecoder(conn)

	var note_changes []NoteChange
	var peer_id string
	var peer Peer

	for {
		msg = Message{}
		rcxMsg(dec, &msg)

		switch msg.Type {
		case "Hangup":
			printHangupMsg(conn); return
		case "Quit":
			println("Quit message received. Exiting..."); os.Exit(1)
		case "WhoAreYou":
			peer_id = msg.Param // client db signature
			msg.Param = whoAmI()
			msg.Type = "WhoIAm"
			sendMsg(enc, msg)
		case "NumberOfChanges":
			// msg.Param will include the synch_point, so send num of changes since synch point
			note_changes = retrieveLocalNoteChangesFromSynchPoint(msg.Param)
			msg.Param = strconv.Itoa(len(note_changes))
			sendMsg(enc, msg)
		case "SendChanges":
			msg.Type = "NoteChange"
			msg.Param = ""
			var note Note
			var note_frag NoteFragment
			for _, change := range(note_changes) {
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
		case "NumberOfClientChanges":
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
			// TODO - go processChanges(peer, &peer_changes)
			db.Where("guid = ?", peer_id).First(&peer) // Do we know of this peer?
		default:
			println("Unknown message type received", msg.Type)
			printHangupMsg(conn); return
		}
	}
}

func whoAmI() string {
	var sig []LocalSig
	db.Find(&sig)
	if len(sig) > 0 {
		return sig[0].Guid
	}
	return ""
}

func sendMsg(encoder *gob.Encoder, msg Message) {
	encoder.Encode(msg); printMsg(msg, false)
	//time.Sleep(10)
}

func rcxMsg(decoder *gob.Decoder, msg *Message) {
	//time.Sleep(10)
	decoder.Decode(&msg); printMsg(*msg, true)
}

func printHangupMsg(conn net.Conn) {
	fmt.Printf("Closing connection: %+v\n----------------------------------------------\n", conn)
}

func printMsg(msg Message, rcx bool) {
	println("\n----------------------------------------------")
	if rcx { print("Received: ")
	} else {
		print("Sent: ")
	}
	println("Msg Type:", msg.Type, " Msg Param:", short_sha(msg.Param))
	msg.NoteChg.Print()
}


// CODE_SCRAP // Yes. A compiled language allows us to do this without any runtime penalty
//	fmt.Printf("encoder is a type of: %v\n", reflect.TypeOf(encoder))

//			// Send a Create Change
//			noteGuid := generate_sha1() // we use the note guid in two places (a little denormalization)
//			note1Guid := noteGuid
//			msg.NoteChg = NoteChange{
//				Operation: 1,
//				Guid: generate_sha1(),
//				NoteGuid: noteGuid,
//				Note: Note{
//					Guid: noteGuid, Title: "Synch Note 1",
//					Description: "Description for Synch Note 1", Body: "Body for Synch Note 1",
//					Tag: "tag_synch_1", CreatedAt: time.Now()},
//				NoteFragment: NoteFragment{},
//			}
//			sendMsg(enc, msg)
//
//			// Send another Create Change
//			noteGuid = generate_sha1()
//			msg.NoteChg = NoteChange{
//				Operation: 1,
//				Guid: generate_sha1(),
//				NoteGuid: noteGuid,
//				Note: Note{
//					Guid: noteGuid, Title: "Synch Note 2",
//					Description: "Description for Synch Note 2", Body: "Body for Synch Note 2",
//					Tag: "tag_synch_2", CreatedAt: time.Now().Add(time.Second)},
//				NoteFragment: NoteFragment{},
//			}
//			second_note_guid := msg.NoteChg.NoteGuid // save for use in update op
//			sendMsg(enc, msg)
//
//			// Send an update operation
//			msg.NoteChg = NoteChange{
//				Operation: 2,
//				Guid: generate_sha1(),
//				NoteGuid: second_note_guid,
//				Note: Note{},
//				NoteFragment: NoteFragment{
//						Bitmask: 0xC, Title: "Synch Note 2 - Updated",
//						Description: "Updated!"},
//			}
//			sendMsg(enc, msg)
//
//			// Send a Delete Change
//			msg.NoteChg = NoteChange{
//				Operation: 3,
//				Guid: generate_sha1(),
//				NoteGuid: note1Guid,
//				Note: Note{},
//				NoteFragment: NoteFragment{},
//			}
//			sendMsg(enc, msg)
