// GOB client server
package main
import(
	//"reflect"
	"os"
	"net"
	"encoding/gob"
	"fmt"
	"strconv"
)

func synch_server() { // WIP
	ln, err := net.Listen("tcp", ":" + SYNCH_PORT) // counterpart of net.Dial
	if err != nil {
		pl("Error setting up server listen on port", SYNCH_PORT)
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

	var local_changes []NoteChange
	var peer_id string
	var peer Peer
	var er error
	var authorized bool = false // true so things will work while we develop auth

	for {
		msg = Message{}
		rcxMsg(dec, &msg)

		switch msg.Type {
		case "Hangup":
			printHangupMsg(conn); return
		case "Quit":
			pl("Quit message received. Exiting..."); os.Exit(1)
		case "WhoAreYou":
			peer_id = msg.Param // receive the client db signature here also
			pl("Client id is:", short_sha(peer_id))
			pl("NoteChg.Guid is:", short_sha(msg.NoteChg.Guid))
			if msg.NoteChg.Guid == get_server_secret() { // then automatically generate a token
				pt, err := getPeerToken(peer_id)
				pl("Auth token generated:", pt)
				if err == nil {
					msg.NoteChg.Guid = pt // include the auth token in next msg
				}
			} else {
				msg.NoteChg.Guid = "" // no token
			}
			msg.Param = whoAmI()
			msg.Type = "WhoIAm"
			peer, er = getPeerByGuid(peer_id)
			if er != nil {
				pl("Error retrieving peer object for peer:", short_sha(peer_id));
				msg.Type = "ERROR"
				msg.Param = "There is no record for this client on the server."
				return
			}
			sendMsg(enc, msg)
		case "AuthMe":
			if peer.Token == msg.Param {
				authorized = true
				msg.Param = "Authorized"
			} else {
				msg.Param = "Declined"
			}
			sendMsg(enc, msg)
		case "LatestChange":
			if !authorized { pl(authFailMsg); return }
			msg.NoteChg = retrieveLatestChange()
			sendMsg(enc, msg)
		case "NumberOfChanges":
			if !authorized { pl(authFailMsg); return }
			// msg.Param will include the synch_point, so send num of changes since synch point
			local_changes = retrieveLocalNoteChangesFromSynchPoint(msg.Param)
			msg.Param = strconv.Itoa(len(local_changes))
			sendMsg(enc, msg)
		case "SendChanges":
			if !authorized { pl(authFailMsg); return }
			msg.Type = "NoteChange"
			msg.Param = ""
			var note Note
			var note_frag NoteFragment
			for _, change := range(local_changes) {
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
			if !authorized { pl(authFailMsg); return }
			numChanges, err := strconv.Atoi(msg.Param)
			if err != nil {
				pl("Could not decode the number of change messages"); return
			}
			if numChanges < 1 { pl("No remote changes."); return }
			pl(numChanges, "changes")
			peer_changes := make([]NoteChange, numChanges)
			sendMsg(enc, Message{Type: "SendChanges"}) // Send the actual changes
			// Receive changes, extract the NoteChanges, save into peer_changes
			for i := 0; i < numChanges; i++ {
				msg = Message{}
				rcxMsg(dec, &msg)
				peer_changes[i] = msg.NoteChg
			}
			pf("\n%d peer changes received:\n", numChanges)
			processChanges(&peer_changes, &local_changes)
		case "NewSynchPoint": // New synch point at the end of synching
			if !authorized { pl(authFailMsg); return }
			synch_nc := msg.NoteChg
			synch_nc.Id = 0 // so it will save
			db.Save(&synch_nc)
			peer.SynchPos = synch_nc.Guid
			db.Save(&peer)
		default:
			pl("Unknown message type received", msg.Type)
			printHangupMsg(conn); return
		}
	}
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
	pl("\n----------------------------------------------")
	if rcx { print("Received: ")
	} else {
		print("Sent: ")
	}
	pl("Msg Type:", msg.Type, " Msg Param:", short_sha(msg.Param))
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
