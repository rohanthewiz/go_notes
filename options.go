package main

import (
	"flag"
	"os"
	"strings"
)

const CentralServer = true // share codebase with central server

//Setup commandline options and other configuration for Go Notes
func getOpts() (map[string]string, map[string]interface{}) {
	opts_str := make(map[string]string)
	opts_intf := make(map[string]interface{})

	qgPtr := flag.String("qg", "", "Query tags based on a LIKE search")
	qtPtr := flag.String("qt", "", "Query title based on a LIKE search")
	qdPtr := flag.String("qd", "", "Query description based on a LIKE search")
	qbPtr := flag.String("qb", "", "Query body based on a LIKE search")
	tPtr := flag.String("t", "", "Create note Title")
	dPtr := flag.String("d", "", "Create note Description")
	bPtr := flag.String("b", "", "Create note Body")
	gPtr := flag.String("g", "", "Comma separated list of Tags for new note")
	qPtr := flag.String("q", "", "Query for notes based on a LIKE search. \"all\" will return all notes")
	pPtr := flag.String("port", "8080", "Specify webserver port")
	adminPtr := flag.String("admin", "", "Privileged actions like 'delete_table'")
	dbPtr := flag.String("db", "", "Sqlite DB path")
	expPtr := flag.String("exp", "", "Export the notes queried to the format of the file given")
	impPtr := flag.String("imp", "", "Import the notes queried from the file given")
	synchClientPtr := flag.String("synch_client", "", "Synch client mode")
	getPeerTokenPtr := flag.String("get_peer_token", "", "Get a token for interacting with this as server")
	savePeerTokenPtr := flag.String("save_peer_token", "", "Save a token for interacting with this as server")
	serverSecretPtr := flag.String("server_secret", "", "Include Server Secret")

	qiPtr := flag.Int64("qi", 0, "Query for notes based on ID")
	lPtr := flag.Int("l", -1, "Limit the number of notes returned")
	sPtr := flag.Bool("s", false, "Short Listing - don't show the body")
	qlPtr := flag.Bool("ql", false, "Query for the last note updated")
	vPtr := flag.Bool("v", false, "Show version")
	whoamiPtr := flag.Bool("whoami", false, "Show Client GUID")
	setupDBPtr := flag.Bool("setup_db", false, "Setup the Database")
	delPtr := flag.Bool("del", false, "Delete the notes queried")
	updPtr := flag.Bool("upd", false, "Update the notes queried")
	svrPtr := flag.Bool("svr", false, "Web server mode")
	getServerSecretPtr := flag.Bool("get_server_secret", false, "Show Server Secret")
	synchServerPtr := flag.Bool("synch_server", false, "Synch server mode")
	verbosePtr := flag.Bool("verbose", true, "verbose mode") // Todo - turn off for production
	debugPtr := flag.Bool("debug", true, "debug mode") // Todo - turn off for production

	flag.Parse()

	// Store options in a couple of maps
	opts_str["q"] = *qPtr
	opts_str["port"] = *pPtr
	opts_str["qg"] = *qgPtr
	opts_str["qt"] = *qtPtr
	opts_str["qd"] = *qdPtr
	opts_str["qb"] = *qbPtr
	opts_str["t"] = *tPtr
	opts_str["d"] = *dPtr
	opts_str["b"] = *bPtr
	opts_str["g"] = *gPtr
	opts_str["admin"] = *adminPtr
	opts_str["db"] = *dbPtr
	opts_str["exp"] = *expPtr
	opts_str["imp"] = *impPtr
	opts_str["synch_client"] = *synchClientPtr
	opts_str["get_peer_token"] = *getPeerTokenPtr
	opts_str["save_peer_token"] = *savePeerTokenPtr
	opts_str["server_secret"] = *serverSecretPtr
	opts_intf["qi"] = *qiPtr
	opts_intf["l"] = *lPtr
	opts_intf["s"] = *sPtr
	opts_intf["ql"] = *qlPtr
	opts_intf["v"] = *vPtr
	opts_intf["whoami"] = *whoamiPtr
	opts_intf["del"] = *delPtr
	opts_intf["upd"] = *updPtr
	opts_intf["svr"] = *svrPtr
	opts_intf["synch_server"] = *synchServerPtr
	opts_intf["get_server_secret"] = *getServerSecretPtr
	opts_intf["setup_db"] = *setupDBPtr
	opts_intf["verbose"] = *verbosePtr
	opts_intf["debug"] = *debugPtr

	separator := "/"
	if strings.Contains(strings.ToUpper(os.Getenv("OS")), "WINDOWS") {
		separator = "\\"
	}
	opts_str["sep"] = separator

	db_file := "go_notes.sqlite"
	var db_folder string
	var db_full_path string
	if len(*dbPtr) == 0 {
		if len(os.Getenv("HOME")) > 0 {
			db_folder = os.Getenv("HOME")
		} else if len(os.Getenv("HOMEDRIVE")) > 0 && len(os.Getenv("HOMEPATH")) > 0 {
			db_folder = os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
		} else {
			db_folder = separator /// last resort
		}
		db_full_path = db_folder + separator + db_file
	} else {
		db_full_path = *dbPtr
	}
	opts_str["db_path"] = db_full_path

	return opts_str, opts_intf
}
