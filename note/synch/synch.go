package synch

import (
	"encoding/gob"
	"fmt"
	"go_notes/authen"
	"go_notes/dbhandle"
	"go_notes/note"
	"go_notes/note/note_ops"
	"go_notes/peer"
	"go_notes/utils"
	"log"
	"net"
	"sort"
	"strconv"
)

type Message struct {
	Type    string
	Param   string
	NoteChg note.NoteChange
}

const SynchPort string = "8090"

func SynchClient(host string, serverSecret string) {
	conn, err := net.Dial("tcp", host+":"+SynchPort)
	if err != nil {
		log.Fatal("Error connecting to server ", err)
	}
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			log.Println("Failed to close connection", err)
		}
	}(conn)
	msg := Message{} // init to empty struct
	enc := gob.NewEncoder(conn)
	dec := gob.NewDecoder(conn)
	defer SendMsg(enc, Message{Type: "Hangup", Param: "", NoteChg: note.NoteChange{}})

	// Send handshake - Client initiates
	SendMsg(enc, Message{
		Type: "WhoAreYou", Param: authen.WhoAmI(), NoteChg: note.NoteChange{Guid: serverSecret}, // borrow NoteChange.Guid
	})

	RcxMsg(dec, &msg) // Decode the response
	if msg.Type == "WhoIAm" {
		peerId := msg.Param // retrieve the server's guid
		utils.Pl("The server's id is", utils.ShortSHA(peerId))
		if len(peerId) != 40 {
			fmt.Println("The server's id is invalid. Run the server once with the -setup_db option")
			return
		}
		// Is there a token for us?
		if len(msg.NoteChg.Guid) == 40 {
			err := peer.SetPeerToken(peerId, msg.NoteChg.Guid)
			if err != nil {
				log.Println("Failed to set peer token:", err)
			} // make sure to save new auth token
		}
		// Obtain the peer object which represents the server
		pr, err := peer.GetPeerByGuid(peerId)
		if err != nil {
			fmt.Println("Error retrieving peer object")
			return
		}
		msg.NoteChg.Guid = "" // hide the evidence

		// Auth
		msg.Type = "AuthMe"
		msg.Param = pr.Token // This is set for the server(peer) by some access granting mechanism
		SendMsg(enc, msg)
		RcxMsg(dec, &msg)
		if msg.Param != "Authorized" {
			fmt.Println("The server declined the authorization request")
			return
		}

		// Do we need to Synch?
		// (SynchPos is the NoteChg.Guid of the last change applied in a synch operation)
		if pr.SynchPos != "" { // Do we have a point of last synch with this peer?
			lastChange := note.RetrieveLatestChange() // Retrieve last local Note Change
			if lastChange.Id > 0 && lastChange.Guid == pr.SynchPos {
				// Get server last change
				msg.Type = "LatestChange"
				SendMsg(enc, msg)
				msg = Message{}
				RcxMsg(dec, &msg)
				if msg.NoteChg.Id > 0 && msg.NoteChg.Guid == pr.SynchPos {
					utils.Pf("We are already in synch with peer: %s at note_change: %s\n",
						utils.ShortSHA(peerId), utils.ShortSHA(lastChange.Guid))
					return
				}
			} // else we probably have never synched so carry on
		}
		utils.Pf("Last known Synch position is \"%s\"\n", utils.ShortSHA(pr.SynchPos))

		// Get Server Changes
		SendMsg(enc, Message{Type: "NumberOfChanges", Param: pr.SynchPos}) // heads up on number of changes
		RcxMsg(dec, &msg)                                                  // Decode the response
		numChanges, err := strconv.Atoi(msg.Param)
		if err != nil {
			fmt.Println("Could not decode the number of change messages")
			return
		}
		utils.Pl(numChanges, "changes")

		peerChanges := make([]note.NoteChange, numChanges) // preallocate slice (optimization)
		SendMsg(enc, Message{Type: "SendChanges"})         // please send the actual changes
		for i := 0; i < numChanges; i++ {
			msg = Message{}
			RcxMsg(dec, &msg)
			peerChanges[i] = msg.NoteChg
		}
		utils.Pf("\n%d server changes received:\n", numChanges)

		// Get Local Changes
		localChanges := note.RetrieveLocalNoteChangesFromSynchPoint(pr.SynchPos)
		utils.Pf("%d local changes after synch point found\n", len(localChanges))
		// Push local changes to server
		if len(localChanges) > 0 {
			SendMsg(enc, Message{Type: "NumberOfClientChanges",
				Param: strconv.Itoa(len(localChanges))})
			RcxMsg(dec, &msg)
			if msg.Type == "SendChanges" {
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
			}
		}

		// Process remote changes received
		if len(peerChanges) > 0 {
			ProcessChanges(&peerChanges, &localChanges)
		}

		// Mark Synch Point with a special NoteChange (Operation: 9)
		// Save on client and server
		if len(peerChanges) > 0 || len(localChanges) > 0 {
			synch_nc := note.NoteChange{Guid: utils.GenerateSHA1(), Operation: 9}
			dbhandle.DB.Save(&synch_nc)
			pr.SynchPos = synch_nc.Guid
			dbhandle.DB.Save(&pr)
			// Mark the server with the same NoteChange
			msg.NoteChg = synch_nc
			msg.Type = "NewSynchPoint"
			SendMsg(enc, msg)
		}

	} else {
		fmt.Println("Peer does not respond to request for database id")
		fmt.Println("Make sure both server and client databases have been properly setup(migrated) with the -setup_db option")
		fmt.Println("or make sure peer version is >= 0.9")
		return
	}

	defer fmt.Println("Synch Operation complete")
}

func ProcessChanges(peerChanges *[]note.NoteChange, localChanges *[]note.NoteChange) {
	utils.Pl("Processing received changes...")
	sort.Sort(note.ByCreatedAt(*peerChanges)) // we will apply in created order
	var localChange note.NoteChange
	var skip bool

	for _, peerChange := range *peerChanges {
		// If we already have this NoteChange locally then skip // same change
		localChange = note.NoteChange{} // make sure local_change is inited here
		// otherwise GORM uses its id in the query - weird!
		dbhandle.DB.Where("guid = ?", peerChange.Guid).First(&localChange)
		if localChange.Id > 1 {
			utils.Pf("We already have NoteChange: %s -- skipping\n", utils.ShortSHA(localChange.Guid))
			continue // we already have that NC
		}
		// If there is a newer local change of the same note and field then skip
		for _, localChange = range *localChanges {
			if localChange.NoteGuid == peerChange.NoteGuid && // same note
				localChange.CreatedAt.After(peerChange.CreatedAt) && ( // local newer
			// any field in peer_change matches a field in local_change
			localChange.NoteFragment.Bitmask&0x8 == peerChange.NoteFragment.Bitmask&0x8 ||
				localChange.NoteFragment.Bitmask&0x4 == peerChange.NoteFragment.Bitmask&0x4 ||
				localChange.NoteFragment.Bitmask&0x2 == peerChange.NoteFragment.Bitmask&0x2 ||
				localChange.NoteFragment.Bitmask&0x1 == peerChange.NoteFragment.Bitmask&0x1) {
				skip = true
			}
		}
		if skip {
			skip = false // reset
			continue
		}

		// Apply Changes
		utils.Pl("____________________APPLYING CHANGE_________________________________")
		note_ops.PerformNoteChange(peerChange)
		note.VerifyNoteChangeApplied(peerChange)
	}
}
