// GOB client server
package main
import(
	//"reflect"
	"net"
	"encoding/gob"
	"fmt"
	"time"
)

func synch_server() { // WIP
	fmt.Println("Server listening on port: " + SYNCH_PORT + " - CTRL-C to quit")
	ln, err := net.Listen("tcp", ":" + SYNCH_PORT) // counterpart of net.Dial
	if err != nil {	println("TODO - handle TCP error") }

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
			msg.Param = "3"
			sendMsg(enc, msg)
		case "SendChanges":
			msg.Type = "NoteChange"
			msg.Param = ""
			msg.NoteChg = NoteChange{Guid: generate_sha1(), Operation: 1, Title: "Synch Note 1",
				Description: "Description for Synch Note 1", Body: "Body for Synch Note 1",
				Tag: "tag_synch_1", CreatedAt: time.Now() }
			sendMsg(enc, msg)
			msg.NoteChg = NoteChange{Guid: generate_sha1(), Operation: 1, Title: "Synch Note 2",
				Description: "Description for Synch Note 2", Body: "Body for Synch Note 2",
				Tag: "tag_synch_2", CreatedAt: time.Now().Add(time.Second) }
			sendMsg(enc, msg)

			msg.NoteChg = NoteChange{Guid: generate_sha1(), Operation: 1, Title: "Synch Note 3",
				Description: "Description for Synch Note 3", Body: "Body for Synch Note 3",
				Tag: "tag_synch_3", CreatedAt: time.Now().Add(time.Millisecond) }
			sendMsg(enc, msg)

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

