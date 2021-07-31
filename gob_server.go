// GOB client server
package main

import (
	"encoding/gob"
	"fmt"
	"go_notes/note"
	"log"
	"net"

	"os"
	"strconv"
)

func synchServer() {
	ln, err := net.Listen("tcp", ":"+SynchPort) // counterpart of net.Dial
	if err != nil {
		fmt.Println("Error setting up server listen on port", SynchPort)
		return
	}
	fmt.Println("Synch Server listening on port: " + SynchPort + " - CTRL-C to quit")

	for {
		conn, err := ln.Accept() // this blocks until connection or error
		if err != nil {
			continue
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	msg := Message{}

	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			log.Println("Error closing conn", err)
		}
	}(conn)

	enc := gob.NewEncoder(conn)
	dec := gob.NewDecoder(conn)

	var localChanges []NoteChange
	var peerId string
	var peer Peer
	var err error
	authorized := false // can set to true when developing auth

	for {
		msg = Message{}
		rcxMsg(dec, &msg)

		switch msg.Type {
		case "Hangup":
			printHangupMsg(conn)
			return

		case "Quit":
			pl("Quit message received. Exiting...")
			os.Exit(1)

			// This is the point of id exchange between the server and client
			// Normal auth process is that the client provides its db signature(peer_id)
			// which the server already has a record of (together with its auth_token)
			// The client provides the required auth_token in the following auth. request message
			// The server must know of the client and auth_token beforehand though.
			// This is achieved one of two ways:
			// (1) The client knows the server's secret token and provides it in the WhoAreYou message
			//     A token is then automatically generated for the client and stored on both ends
			// (2) In a manual operation (At the cmd line for now), the client provides its db signature
			//     to the server in a request for an auth_token
			//     (a) The client does ./go_notes -whoami  # output the client's db signature
			//     (b) At the server a token is generated for the client with
			//         ./go_notes -get_peer_token peer_db_signature
			//     (c) The client saves this token with
			//         ./go_notes -save_peer_token the_token (detail: the_token here is of the format "server_id-auth_token")
		case "WhoAreYou":
			peerId = msg.Param // receive the client db signature
			pd("Client id is:", shortSHA(peerId))
			pl("NoteChg.Guid is:", shortSHA(msg.NoteChg.Guid))
			if msg.NoteChg.Guid == getServerSecret() { // then automatically generate a token
				pt, err := getPeerToken(peerId)
				if err != nil {
					msg.NoteChg.Guid = ""
				} else {
					pl("Auth token generated:", pt)
					msg.NoteChg.Guid = pt // include the auth token in next msg
				}
			} else {
				msg.NoteChg.Guid = "" // no token
			}
			// Always return the server's id
			msg.Param = whoAmI() // reply with the server's db signature
			msg.Type = "WhoIAm"

			// Retrieve the actual peer object which represents the client
			peer, err = getPeerByGuid(peerId)
			if err != nil {
				fmt.Println("Error retrieving peer object for peer:", shortSHA(peerId))
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

		// How client determines if we need to synch
		case "LatestChange":
			if !authorized {
				pl(authFailMsg)
				return
			}
			msg.NoteChg = retrieveLatestChange() // Return server's last local Note Change
			sendMsg(enc, msg)

		case "NumberOfChanges":
			if !authorized {
				pl(authFailMsg)
				return
			}
			// msg.Param will include the synch_point, so send num of changes since synch point
			localChanges = retrieveLocalNoteChangesFromSynchPoint(msg.Param)
			msg.Param = strconv.Itoa(len(localChanges))
			sendMsg(enc, msg)

		case "SendChanges":
			if !authorized {
				pl(authFailMsg)
				return
			}
			msg.Type = "NoteChange"
			msg.Param = ""
			var nte note.Note
			var noteFrag NoteFragment

			for _, change := range localChanges {
				nte = note.Note{}
				noteFrag = NoteFragment{}
				// We have the change but now we need the NoteFragment or Note depending on the operation type
				if change.Operation == 1 {
					db.Where("id = ?", change.NoteId).First(&nte)
					nte.Id = 0
					change.Note = nte
				}
				if change.Operation == 2 {
					db.Where("id = ?", change.NoteFragmentId).First(&noteFrag)
					change.NoteFragment = noteFrag
				}
				msg.NoteChg = change
				msg.NoteChg.Print()
				sendMsg(enc, msg)
			}
		case "NumberOfClientChanges":
			if !authorized {
				pl(authFailMsg)
				return
			}
			numChanges, err := strconv.Atoi(msg.Param)
			if err != nil {
				pl("Could not decode the number of change messages")
				return
			}
			if numChanges < 1 {
				pl("No remote changes.")
				return
			}
			pl(numChanges, "changes")

			peerChanges := make([]NoteChange, numChanges)
			sendMsg(enc, Message{Type: "SendChanges"}) // Send the actual changes
			// Receive changes, extract the NoteChanges, save into peer_changes
			for i := 0; i < numChanges; i++ {
				msg = Message{}
				rcxMsg(dec, &msg)
				peerChanges[i] = msg.NoteChg
			}
			pf("\n%d peer changes received:\n", numChanges)
			processChanges(&peerChanges, &localChanges)

		case "NewSynchPoint": // New synch point at the end of synching
			if !authorized {
				pl(authFailMsg)
				return
			}
			synchNC := msg.NoteChg
			synchNC.Id = 0 // so it will save
			db.Save(&synchNC)
			peer.SynchPos = synchNC.Guid
			db.Save(&peer)
		default:
			pl("Unknown message type received", msg.Type)
			printHangupMsg(conn)
			return
		}
	}
}

func sendMsg(encoder *gob.Encoder, msg Message) {
	err := encoder.Encode(msg)
	if err != nil {
		log.Println("Error on message encode:", err)
	}
	printMsg(msg, false)
}

func rcxMsg(decoder *gob.Decoder, msg *Message) {
	err := decoder.Decode(&msg)
	if err != nil {
		log.Println("error on message decode:", err)
	}
	printMsg(*msg, true)
}

func printHangupMsg(conn net.Conn) {
	fmt.Printf("Closing connection: %+v\n----------------------------------------------\n", conn)
}

func printMsg(msg Message, rcx bool) {
	pl("\n----------------------------------------------")
	if rcx {
		print("Received: ")
	} else {
		print("Sent: ")
	}
	pl("Msg Type:", msg.Type, " Msg Param:", shortSHA(msg.Param))
	msg.NoteChg.Print()
}
