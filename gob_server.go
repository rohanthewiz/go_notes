// GOB client server
package main
import(
	//"reflect"
	"os"
	"net"
	"encoding/gob"
	"fmt"
	"errors"
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
	var er error

	for {
		msg = Message{}
		rcxMsg(dec, &msg)

		switch msg.Type {
		case "Hangup":
			printHangupMsg(conn); return
		case "Quit":
			println("Quit message received. Exiting..."); os.Exit(1)
		case "WhoAreYou":
			peer_id = msg.Param // receive the client db signature here also
			if len(peer_id) == 40 {
				msg.Param = whoAmI()
				msg.Type = "WhoIAm"
				peer, er = getOrCreatePeer(peer_id)
				if er != nil { println("Error retrieving peer object"); return }
			} else {
				msg.Type = "InvalidPeerId"
			}
			sendMsg(enc, msg)
		case "LatestChange":
			msg.NoteChg = retrieveLatestChange()
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
			if err != nil {
				println("Could not decode the number of change messages"); return
			}
			if numChanges < 1 { println("No remote changes."); return }
			println(numChanges, "changes")
			peer_changes := make([]NoteChange, numChanges)
			sendMsg(enc, Message{Type: "SendChanges"}) // Send the actual changes
			// Receive changes, extract the NoteChanges, save into peer_changes
			for i := 0; i < numChanges; i++ {
				msg = Message{}
				rcxMsg(dec, &msg)
				peer_changes[i] = msg.NoteChg
			}
			pf("\n%d peer changes received:\n", numChanges)
			processChanges(peer, &peer_changes)
		case "NewSynchPoint": // New synch point at the end of synching
			synch_nc := msg.NoteChg
			synch_nc.Id = 0 // so it will save
			db.Save(&synch_nc)
			peer.SynchPos = synch_nc.Guid
			db.Save(&peer)
		default:
			println("Unknown message type received", msg.Type)
			printHangupMsg(conn); return
		}
	}
}

// Get local DB signature
func whoAmI() string {
	var local_sig LocalSig
	db.First(&local_sig)
	if local_sig.Id < 1 {
		ensureDBSig()
		db.First(&local_sig)
		if local_sig.Id < 1 {
			println("Could not locate or create local database signature.\nYou should back up your notes, delete the local database, import your notes then try again")
			return ""
		}
	}
	return local_sig.Guid
}

func getOrCreatePeer(peer_id string) (Peer, error) {
	var peer Peer
	db.Where("guid = ?", peer_id).First(&peer)
	if peer.Id < 1 {
		println("Creating new peer:", peer_id)
		db.Create(&Peer{Guid: peer_id})
		db.Where("guid = ?", peer_id).First(&peer)
		if peer.Id < 1 {
			return peer, errors.New("Could not create peer")
		}
	}
	return peer, nil
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
