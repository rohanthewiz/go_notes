package main

import (
	"fmt"
	"go_notes/config"
	db "go_notes/dbhandle"
	"go_notes/dbutil"
	"go_notes/migration"
	"go_notes/user"
	"os"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

const app_name = "GoNotes"
const version string = "0.14.0"

// Get Commandline Options and Flags
var optsStr, optsIntf = getOpts() // returns map[string]string, map[string]interface{}

func main() {
	err := db.InitDB()
	if err != nil { // Can't err chk db conn outside method, so do it here
		fmt.Println("There was an error connecting to the DB")
		fmt.Println("DBPath: " + config.Opts.DBPath)
		os.Exit(2)
	}

	migration.MigrateIfNeeded()

	if optsIntf["v"].(bool) {
		fmt.Println(app_name, version)
		return
	}

	if optsStr["admin"] == "delete_tables" {
		dbutil.DeleteTables()
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

	if config.Opts.CreateUser != "" {
		_ = user.CreateUser(user.User{
			Email:    config.Opts.Email,
			Username: config.Opts.CreateUser,
		}, config.Opts.Password)
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
		migration.Migrate()

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
