package main

import (
	"fmt"
	"go_notes/authen"
	"go_notes/config"
	db "go_notes/dbhandle"
	"go_notes/migration"
	"go_notes/note/note_ops"
	"go_notes/note/synch"
	"go_notes/peer"
	"go_notes/user"
	"go_notes/web"
	"os"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

const app_name = "GoNotes"
const version string = "0.15.0"

// Get Commandline Options and Flags
var optsStr, optsIntf = getOpts() // returns map[string]string, map[string]interface{}

func main() {
	o := config.Opts // alias - note alias is a copy of config.Opts

	if o.Verbose {
		fmt.Printf("config.Opts ->%#v\n", config.Opts)
		fmt.Println(strings.Repeat("-", 45))
	}

	err := db.InitDB()
	if err != nil { // Can't err chk db conn outside method, so do it here
		fmt.Println("There was an error connecting to the DB")
		fmt.Println("DBPath: " + o.DBPath)
		os.Exit(2)
	}

	migration.MigrateIfNeeded()

	if o.Version {
		fmt.Println(app_name, version)
		return
	}

	// This needs to be put behind auth
	// if optsStr["admin"] == "delete_tables" {
	// 	dbutil.DeleteTables()
	// 	return
	// }

	// Client - Return our db signature
	if o.WhoAmI {
		fmt.Println(authen.WhoAmI())
		return
	}

	// Server - Generate an auth token for a client
	// The format of the generated token is: server_id-auth_token_for_the_client
	if optsStr["get_peer_token"] != "" {
		pt, err := peer.GetPeerToken(optsStr["get_peer_token"])
		if err != nil {
			fmt.Println("Error retrieving token")
			return
		}
		fmt.Printf("Peer token is: %s-%s\nYou will now need to run the client with \n'go_notes -save_peer_token the_token'\n",
			authen.WhoAmI(), pt)
		return
	}

	// Client - Save a token generated for us by a server
	if optsStr["save_peer_token"] != "" {
		peer.SavePeerToken(optsStr["save_peer_token"])
		return
	}

	// Server - Return the server's secret token
	// This is a master key and will allow any client to auth
	// We probably want to use the methods above instead
	if optsIntf["get_server_secret"].(bool) {
		fmt.Println(authen.GetServerSecret())
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
		web.Webserver(o.Port)

	} else if config.Opts.IsSynchSvr {
		synch.SynchServer()

	} else if config.Opts.IsRemoteSvr { // remote web server
		go synch.SynchServer()
		web.Webserver(o.Port)

	} else if optsStr["synch_client"] != "" {
		synch.SynchClient(optsStr["synch_client"], optsStr["server_secret"])

	} else if optsIntf["setup_db"].(bool) { // Migrate the DB
		migration.Migrate()

		// QUERY
	} else if optsStr["q"] != "" || optsIntf["qi"].(int64) != 0 ||
		optsStr["qg"] != "" || optsStr["qt"] != "" ||
		optsStr["qb"] != "" || optsStr["qd"] != "" {
		note_ops.CmdlineQueryAndAction()

		// Other options
	} else if o.Import != "" { // import
		note_ops.HandleImport()

		// Create
	} else if optsStr["t"] != "" { // No query options, we must be trying to CREATE
		note_ops.CreateNote(optsStr["t"], optsStr["d"], optsStr["b"], optsStr["g"])

	} else {
		fmt.Println("starting webserver", o.Port)
		web.Webserver(o.Port)
	}
}
