// GOB client server
package main
import(
	//"reflect"
	"net"
	"encoding/gob"
	"fmt"
	"time"
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
	for {
		msg = Message{}
		rcxMsg(dec, &msg)

		switch msg.Type {
		case "Hangup":
			printHangupMsg(conn); return
		case "WhoAreYou":
			msg.Param = whoAmI()
			msg.Type = "WhoIAm"
			sendMsg(enc, msg)
		case "NumberOfChanges":
			// msg.Param will include the synch_point, so send # changes since synch point
			note_changes :=	retrieveLocalNoteChangesFromSynchPoint(msg.Param)
			msg.Param = strconv.Itoa(len(note_changes))
//			fmt.Sprintf(msg.Param, "%d", len(note_changes))
			sendMsg(enc, msg)
		case "SendChanges":
			msg.Type = "NoteChange"
			msg.Param = ""

			for _, change := range(note_changes) {
				msg.NoteChg = change
				pf("NoteChg is %v", msg.NoteChg)
				sendMsg(enc, msg)
			}
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

		default:
			println("Unknown message type received")
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
	time.Sleep(10)
}

func rcxMsg(decoder *gob.Decoder, msg *Message) {
	time.Sleep(10)
	decoder.Decode(&msg); printMsg(*msg, true)
}

func printHangupMsg(conn net.Conn) {
	fmt.Printf("Closing connection: %+v...\n----------------------------------------------\n", conn)
}

func printMsg(msg Message, rcx bool) {
	if rcx { print("Received: ")
	} else {
		print("Sent: ")
	}
	fmt.Printf("%+v\n----------------------------------------------\n", msg)
}

// CODE_SCRAP
//	fmt.Printf("encoder is a type of: %v\n", reflect.TypeOf(encoder))

