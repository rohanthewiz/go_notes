package main
import(
	"fmt"
	"log"
	"strconv"
	"net"
	"encoding/gob"
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

		fmt.Println("\nPeer changes received:\n")
		for _, item := range(peer_changes) {
			fmt.Printf("Title: %s, Guid: %s\n", item.Title, item.Guid)
		}

		// Todo: Synch in changes in asc date order
		
	} else {
        println("Peer does not respond to request for database id\nRun peer with -setup_db option or make sure peer version is >= 0.9")
		return
    }

	// Send Hangup
	sendMsg(enc, Message{Type: "Hangup", Param: "", NoteChg: NoteChange{}})
	println("Client done")
}
