package main

import (
	"fmt"
	"go_notes/config"
	"go_notes/note"
	"os"
	"strings"

	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"
)

const app_name = "GoNotes"
const version string = "0.14.0"

// Get Commandline Options and Flags
var optsStr, optsIntf = getOpts() // returns map[string]string, map[string]interface{}

// Init db
var db, dbErr = gorm.Open("sqlite3", optsStr["db_path"])

func migrate() {
	// Create or update the table structure as needed
	pl("Migrating the DB...")
	// If the table is not existing, AutoMigrate will create the table automatically.
	// Fyi, AutoMigrate will only *add new columns*, it won't update column's type or delete unused columns
	// According to GORM: Feel free to change your struct, AutoMigrate will keep your database up-to-date.
	db.AutoMigrate(&note.Note{}, &NoteChange{}, &NoteFragment{}, &LocalSig{}, &Peer{})

	db.Model(&note.Note{}).AddUniqueIndex("idx_note_guid", "guid")
	db.Model(&note.Note{}).AddUniqueIndex("idx_note_title", "title")
	db.Model(&NoteChange{}).AddUniqueIndex("idx_note_change_guid", "guid")
	db.Model(&NoteChange{}).AddIndex("idx_note_change_note_guid", "note_guid")
	db.Model(&NoteChange{}).AddIndex("idx_note_change_created_at", "created_at")

	ensureDBSig() // Initialize local with a SHA1 signature if it doesn't already have one
	pl("Migration complete")
}

func ensureDBSig() {
	var localSigs []LocalSig
	db.Find(&localSigs)

	if len(localSigs) == 1 && len(localSigs[0].Guid) == 40 &&
		len(localSigs[0].ServerSecret) == 40 {
		return
	} // all is good

	if len(localSigs) == 0 { // create the signature
		db.Create(&LocalSig{Guid: generateSHA1(), ServerSecret: generateSHA1()})
		if db.Find(&localSigs); len(localSigs) == 1 && len(localSigs[0].Guid) == 40 { // was it saved?
			pl("Local signature created")
		}
	} else {
		panic("Error in the 'local_sigs' table. There should be only one and only one good row")
	}
}

func main() {

	if dbErr != nil { // Can't err chk db conn outside method, so do it here
		fmt.Println("There was an error connecting to the DB")
		fmt.Println("DBPath: " + optsStr["db_path"])
		os.Exit(2)
	}

	// Do we need to migrate?
	if !db.HasTable(&Peer{}) || !db.HasTable(&note.Note{}) || !db.HasTable(&NoteChange{}) ||
		!db.HasTable(&NoteFragment{}) || !db.HasTable(&LocalSig{}) {
		migrate()
	}

	if optsIntf["v"].(bool) {
		fmt.Println(app_name, version)
		return
	}

	// db.LogMode(optsIntf["debug"].(bool)) // Set debug mode for Gorm db

	if optsStr["admin"] == "delete_tables" {
		fmt.Println("Are you sure you want to delete all data? (N/y)")
		var input string
		_, err := fmt.Scanln(&input) // Get keyboard input
		if err != nil {
			return
		}
		pd("input", input)
		if input == "y" || input == "Y" {
			db.DropTableIfExists(&note.Note{})
			db.DropTableIfExists(&NoteChange{})
			db.DropTableIfExists(&NoteFragment{})
			db.DropTableIfExists(&Peer{})
			db.DropTableIfExists(&LocalSig{})
			pl("Notes tables deleted")
		}
		return
	}

	// Client - Return our db signature
	if optsIntf["whoami"].(bool) {
		fmt.Println(whoAmI())
		return
	}

	// Server - Generate an auth token for a client
	// The format of the generated token is: server_id-auth_token_for_the_client
	if optsStr["get_peer_token"] != "" {
		pt, err := getPeerToken(optsStr["get_peer_token"])
		if err != nil {
			fmt.Println("Error retrieving token")
			return
		}
		fmt.Printf("Peer token is: %s-%s\nYou will now need to run the client with \n'go_notes -save_peer_token the_token'\n",
			whoAmI(), pt)
		return
	}

	// Client - Save a token generated for us by a server
	if optsStr["save_peer_token"] != "" {
		savePeerToken(optsStr["save_peer_token"])
		return
	}

	// Server - Return the server's secret token
	// This is a master key and will allow any client to auth
	// We probably want to use the methods above instead
	if optsIntf["get_server_secret"].(bool) {
		fmt.Println(getServerSecret())
		return
	}

	// CORE PROCESSING

	// when -remote require auth and start synch server in background

	if config.Opts.IsLocalWebSvr { // local only webserver - security is relaxed
		webserver(optsStr["port"])

	} else if config.Opts.IsSynchSvr {
		synchServer()

	} else if config.Opts.IsRemoteSvr { // remote web server
		go synchServer()
		webserver(optsStr["port"])

	} else if optsStr["synch_client"] != "" {
		synchClient(optsStr["synch_client"], optsStr["server_secret"])

	} else if optsIntf["setup_db"].(bool) { // Migrate the DB
		migrate()

	} else if optsStr["q"] != "" || optsIntf["qi"].(int64) != 0 ||
		optsStr["qg"] != "" || optsStr["qt"] != "" ||
		optsStr["qb"] != "" || optsStr["qd"] != "" {
		// QUERY
		notes := queryNotes()

		// List Notes found
		fmt.Println("") // for UI sake
		listNotes(notes, true)

		// Options that can go with Query
		// export
		if optsStr["exp"] != "" {
			arr := strings.Split(optsStr["exp"], ".")
			arr_item_last := len(arr) - 1
			if arr[arr_item_last] == "csv" {
				exportCsv(notes, optsStr["exp"])
			}
			if arr[arr_item_last] == "gob" {
				exportGob(notes, optsStr["exp"])
			}
		} else if optsIntf["upd"].(bool) { // update
			UpdateNotes(notes)

			// See if we want to delete
		} else if optsIntf["del"].(bool) {
			DeleteNotes(notes)
		}
		// Other options
	} else if optsStr["imp"] != "" { // import
		arr := strings.Split(optsStr["imp"], ".")
		arr_item_last := len(arr) - 1
		if arr[arr_item_last] == "csv" {
			importCsv(optsStr["imp"])
		}
		if arr[arr_item_last] == "gob" {
			importGob(optsStr["imp"])
		}
		// Create
	} else if optsStr["t"] != "" { // No query options, we must be trying to CREATE
		CreateNote(optsStr["t"], optsStr["d"], optsStr["b"], optsStr["g"])
	} else {
		webserver(optsStr["port"])
	}
}
