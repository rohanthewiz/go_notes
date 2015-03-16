package main

import (
	"fmt"
	"os"
	"strings"
	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"
)

const app_name = "GoNotes"
const version string = "0.8.18"

// Get Commandline Options and Flags
var opts_str, opts_intf = getOpts() //returns map[string]string, map[string]interface{}

// Init db
var db, db_err = gorm.Open("sqlite3", opts_str["db_path"])

func migrate() {
	// Create or update the table structure as needed
	println("Migrating the DB...")
	db.AutoMigrate(&Note{}, &NoteChange{}, &NoteFragment{}, &LocalSig{}, &Peer{})
	//According to GORM: Feel free to change your struct, AutoMigrate will keep your database up-to-date.
	// Fyi, AutoMigrate will only *add new columns*, it won't update column's type or delete unused columns, to make sure your data is safe.
	// If the table is not existing, AutoMigrate will create the table automatically.

	db.Model(&Note{}).AddUniqueIndex("idx_note_guid", "guid")
	db.Model(&Note{}).AddUniqueIndex("idx_note_title", "title")
	db.Model(&NoteChange{}).AddUniqueIndex("idx_note_change_guid", "guid")
	db.Model(&NoteChange{}).AddIndex("idx_note_change_note_guid", "note_guid")
	db.Model(&NoteChange{}).AddIndex("idx_note_change_created_at", "created_at")

	ensureDBSig() // Initialize local with a SHA1 signature if it doesn't already have one
	println("Migration complete")
}

func ensureDBSig() {
	var local_sigs []LocalSig
	db.Find(&local_sigs)

	if len(local_sigs) == 1 && len(local_sigs[0].Guid) == 40 &&
		len(local_sigs[0].ServerSecret) == 40 { return } // all is good

	if len(local_sigs) == 0 { // create the signature
		db.Create(&LocalSig{Guid: generate_sha1(), ServerSecret: generate_sha1()})
		if db.Find(&local_sigs); len(local_sigs) == 1 && len(local_sigs[0].Guid) == 40 { // was it saved?
			println("Local signature created")
		}
	} else {
		panic("Error in the 'local_sigs' table. There should be only one and only one good row")
	}
}

func main() {

	if db_err != nil { // Can't err chk db conn outside method, so do it here
		println("There was an error connecting to the DB")
		println("DBPath: " + opts_str["db_path"])
		os.Exit(2)
	}

	//Do we need to migrate?
	if ! db.HasTable(&Peer{}) || ! db.HasTable(&Note{}) || ! db.HasTable(&NoteChange{}) ||
		! db.HasTable(&NoteFragment{}) || ! db.HasTable(&LocalSig{}) { migrate() }

	if opts_intf["v"].(bool) {
		println(app_name, version)
		return
	}

	if opts_str["admin"] == "delete_tables" {
		fmt.Println("Are you sure you want to delete all data? (N/y)")
		var input string
		fmt.Scanln(&input) // Get keyboard input
		println("input", input)
		if input == "y" || input == "Y" {
			db.DropTableIfExists(&Note{})
			db.DropTableIfExists(&NoteChange{})
			db.DropTableIfExists(&NoteFragment{})
			db.DropTableIfExists(&Peer{})
			db.DropTableIfExists(&LocalSig{})
			println("Notes tables deleted")
		}
		return
	}

	// Client
	if opts_intf["whoami"].(bool) {
		println(whoAmI())
		return
	}

	// Server
	if opts_intf["get_server_secret"].(bool) {
		println(get_server_secret())
		return
	}

	// Server
	if opts_str["get_peer_token"] != "" {
		pt, err := getPeerToken(opts_str["get_peer_token"])
		if err != nil {println("Error retrieving token"); return}
		fmt.Printf("Peer token is: %s-%s\nYou will now need to run the client with 'go_notes -save_peer_token the_token'\n",
			whoAmI(), pt)
		return
	}

	// Client
	if opts_str["save_peer_token"] != "" {
		savePeerToken(opts_str["save_peer_token"])
		return
	}

	// CORE PROCESSING

	if opts_intf["svr"].(bool) {
		doWebServer()

	} else if opts_str["t"] != "" { // No query options, we must be trying to CREATE
		createNote()

	} else if opts_str["synch_client"] != "" { // client to test synching
			synch_client(opts_str["synch_client"], opts_str["server_secret"])

	} else if opts_intf["synch_server"].(bool) { // server to test synching
		synch_server()

	} else if opts_intf["setup_db"].(bool) { // Migrate the DB
		migrate()

	} else if opts_str["q"] != "" || opts_intf["qi"].(int) != 0 ||
				opts_str["qg"] != "" || opts_str["qt"] != "" ||
				opts_str["qb"] != "" || opts_str["qd"] != "" {
		// QUERY
		notes := queryNotes()

		// List Notes found
		println("")  // for UI sake
		listNotes(notes, true)

		// Options that can go with Query
		// export
		if opts_str["exp"] != "" {
			arr := strings.Split(opts_str["exp"], ".")
			arr_item_last := len(arr) -1
			if arr[arr_item_last] == "csv" {
				exportCsv(notes, opts_str["exp"])
			}
			if arr[arr_item_last] == "gob" {
				exportGob(notes, opts_str["exp"])
			}
		} else if opts_intf["upd"].(bool) { // update
			updateNotes(notes)

			// See if we want to delete
		} else if opts_intf["del"].(bool) {
			deleteNotes(notes)
		}
		// Other options
	} else if opts_str["imp"] != "" { // import
			arr := strings.Split(opts_str["imp"], ".")
			arr_item_last := len(arr) -1
			if arr[arr_item_last] == "csv" {
				importCsv(opts_str["imp"])
			}
			if arr[arr_item_last] == "gob" {
				importGob(opts_str["imp"])
			}
	}
}
