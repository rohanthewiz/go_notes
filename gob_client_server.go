// GOB client server
package main
import(
	//"reflect"
	"net"
	"encoding/gob"
	"fmt"
	"time"
	"log"
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
	msg := Message{} // empty struct (think object)
	enc := gob.NewEncoder(conn)
	dec := gob.NewDecoder(conn)
//	fmt.Printf("encoder is a type of: %v\n", reflect.TypeOf(encoder))
//	fmt.Printf("decoder is a type of: %v\n", reflect.TypeOf(decoder))
	for {
		rcxMsg(dec, &msg)
		if msg.Type == "Hangup" {
			printHangupMsg(conn); conn.Close(); break
		}
		time.Sleep(10) //nano sec
		msg.Type = "Server Response"
		sendMsg(enc, msg)
		println("Delaying for a bit so you can see the interaction..."); time.Sleep(2 * time.Second)
	}
	println("Connection closed")
}

func synch_client() {
	fmt.Println("Synch Client\n-------------")
	conn, err := net.Dial("tcp", "localhost:" + SYNCH_PORT)
	if err != nil {log.Fatal("Connection error", err)}
	defer conn.Close()
	msg := Message{} // init to empty struct
	enc := gob.NewEncoder(conn)
	dec := gob.NewDecoder(conn)
	// Send a message
	sendMsg(enc, Message{Type: "GetSynchPoint", Param: "", NoteChg: NoteChange{Title: "Synch this!", Description: "Just a test"}})
	rcxMsg(dec, &msg) // Decode the response
	sendMsg(enc, Message{Type: "ReturnChangeset", Param: "", NoteChg: NoteChange{Title: "We are really talking now", Description: "Just a test"}})
	rcxMsg(dec, &msg)
	//Send Hangup
	sendMsg(enc, Message{Type: "Hangup", Param: "", NoteChg: NoteChange{}})
	fmt.Println("Client done")
}

func sendMsg(encoder *gob.Encoder, msg Message) {
	encoder.Encode(msg); printMsg(msg, false)
}

func rcxMsg(decoder *gob.Decoder, msg *Message) {
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

