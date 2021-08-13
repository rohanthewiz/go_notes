package synch

import (
	"encoding/gob"
	"fmt"
	"go_notes/authen"
	"go_notes/dbhandle"
	"go_notes/note"
	"go_notes/peer"
	"go_notes/utils"
	"log"
	"net"
	"os"
	"strconv"
)

const authFailMsg = "Authentication failure. Generate authorization token with -synch_auth\nThen store in peer entry on client with -store_synch_auth"

func SynchServer() {
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

	var localChanges []note.NoteChange
	var peerId string
	var pr peer.Peer
	var err error
	authorized := false // can set to true when developing auth

	for {
		msg = Message{}
		RcxMsg(dec, &msg)

		switch msg.Type {
		case "Hangup":
			PrintHangupMsg(conn)
			return

		case "Quit":
			utils.Pl("Quit message received. Exiting...")
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
			utils.Pd("Client id is:", utils.ShortSHA(peerId))
			utils.Pl("NoteChg.Guid is:", utils.ShortSHA(msg.NoteChg.Guid))
			if msg.NoteChg.Guid == authen.GetServerSecret() { // then automatically generate a token
				pt, err := peer.GetPeerToken(peerId)
				if err != nil {
					msg.NoteChg.Guid = ""
				} else {
					utils.Pl("Auth token generated:", pt)
					msg.NoteChg.Guid = pt // include the auth token in next msg
				}
			} else {
				msg.NoteChg.Guid = "" // no token
			}
			// Always return the server's id
			msg.Param = authen.WhoAmI() // reply with the server's db signature
			msg.Type = "WhoIAm"

			// Retrieve the actual peer object which represents the client
			pr, err = peer.GetPeerByGuid(peerId)
			if err != nil {
				fmt.Println("Error retrieving peer object for peer:", utils.ShortSHA(peerId))
				msg.Type = "ERROR"
				msg.Param = "There is no record for this client on the server."
				return
			}
			SendMsg(enc, msg)

		case "AuthMe":
			if pr.Token == msg.Param {
				authorized = true
				msg.Param = "Authorized"
			} else {
				msg.Param = "Declined"
			}
			SendMsg(enc, msg)

		// How client determines if we need to synch
		case "LatestChange":
			if !authorized {
				utils.Pl(authFailMsg)
				return
			}
			msg.NoteChg = note.RetrieveLatestChange() // Return server's last local Note Change
			SendMsg(enc, msg)

		case "NumberOfChanges":
			if !authorized {
				utils.Pl(authFailMsg)
				return
			}
			// msg.Param will include the synch_point, so send num of changes since synch point
			localChanges = note.RetrieveLocalNoteChangesFromSynchPoint(msg.Param)
			msg.Param = strconv.Itoa(len(localChanges))
			SendMsg(enc, msg)

		case "SendChanges":
			if !authorized {
				utils.Pl(authFailMsg)
				return
			}
			msg.Type = "NoteChange"
			msg.Param = ""
			var nte note.Note
			var noteFrag note.NoteFragment

			for _, change := range localChanges {
				nte = note.Note{}
				noteFrag = note.NoteFragment{}
				// We have the change but now we need the NoteFragment or Note depending on the operation type
				if change.Operation == 1 {
					dbhandle.DB.Where("id = ?", change.NoteId).First(&nte)
					nte.Id = 0
					change.Note = nte
				}
				if change.Operation == 2 {
					dbhandle.DB.Where("id = ?", change.NoteFragmentId).First(&noteFrag)
					change.NoteFragment = noteFrag
				}
				msg.NoteChg = change
				msg.NoteChg.Print()
				SendMsg(enc, msg)
			}
		case "NumberOfClientChanges":
			if !authorized {
				utils.Pl(authFailMsg)
				return
			}
			numChanges, err := strconv.Atoi(msg.Param)
			if err != nil {
				utils.Pl("Could not decode the number of change messages")
				return
			}
			if numChanges < 1 {
				utils.Pl("No remote changes.")
				return
			}
			utils.Pl(numChanges, "changes")

			peerChanges := make([]note.NoteChange, numChanges)
			SendMsg(enc, Message{Type: "SendChanges"}) // Send the actual changes
			// Receive changes, extract the NoteChanges, save into peer_changes
			for i := 0; i < numChanges; i++ {
				msg = Message{}
				RcxMsg(dec, &msg)
				peerChanges[i] = msg.NoteChg
			}
			utils.Pf("\n%d peer changes received:\n", numChanges)
			ProcessChanges(&peerChanges, &localChanges)

		case "NewSynchPoint": // New synch point at the end of synching
			if !authorized {
				utils.Pl(authFailMsg)
				return
			}
			synchNC := msg.NoteChg
			synchNC.Id = 0 // so it will save
			dbhandle.DB.Save(&synchNC)
			pr.SynchPos = synchNC.Guid
			dbhandle.DB.Save(&pr)
		default:
			utils.Pl("Unknown message type received", msg.Type)
			PrintHangupMsg(conn)
			return
		}
	}
}

func SendMsg(encoder *gob.Encoder, msg Message) {
	err := encoder.Encode(msg)
	if err != nil {
		log.Println("Error on message encode:", err)
	}
	PrintMsg(msg, false)
}

func RcxMsg(decoder *gob.Decoder, msg *Message) {
	err := decoder.Decode(&msg)
	if err != nil {
		log.Println("error on message decode:", err)
	}
	PrintMsg(*msg, true)
}

func PrintHangupMsg(conn net.Conn) {
	fmt.Printf("Closing connection: %+v\n----------------------------------------------\n", conn)
}

func PrintMsg(msg Message, rcx bool) {
	utils.Pl("\n----------------------------------------------")
	if rcx {
		print("Received: ")
	} else {
		print("Sent: ")
	}
	utils.Pl("Msg Type:", msg.Type, " Msg Param:", utils.ShortSHA(msg.Param))
	msg.NoteChg.Print()
}
