package main

import (
	"flag"
	"go_notes/config"
	"os"
	"strings"
)

// Setup commandline options and other configuration for Go Notes
func getOpts() (map[string]string, map[string]interface{}) {
	strOpts := make(map[string]string, 32)
	intfOpts := make(map[string]interface{}, 16)

	qgPtr := flag.String("qg", "", "Query tags based on a LIKE search")
	qtPtr := flag.String("qt", "", "Query title based on a LIKE search")
	qdPtr := flag.String("qd", "", "Query description based on a LIKE search")
	qbPtr := flag.String("qb", "", "Query body based on a LIKE search")
	tPtr := flag.String("t", "", "Create note Title")
	dPtr := flag.String("d", "", "Create note Description")
	bPtr := flag.String("b", "", "Create note Body")
	gPtr := flag.String("g", "", "Comma separated list of Tags for new note")
	qPtr := flag.String("q", "", "Query for notes based on a LIKE search. \"all\" will return all notes")
	pPtr := flag.String("port", "8092", "Specify webserver port")
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
	ptrRemoteSvr := flag.Bool("remote_server", false, "Remote server mode") // run as a remote web server
	verbosePtr := flag.Bool("verbose", true, "verbose mode")                // Todo - turn off for production
	debugPtr := flag.Bool("debug", true, "debug mode")                      // Todo - turn off for production

	flag.Parse()

	// Store options in a couple of maps
	strOpts["q"] = *qPtr
	strOpts["port"] = *pPtr
	strOpts["qg"] = *qgPtr
	strOpts["qt"] = *qtPtr
	strOpts["qd"] = *qdPtr
	strOpts["qb"] = *qbPtr
	strOpts["t"] = *tPtr
	strOpts["d"] = *dPtr
	strOpts["b"] = *bPtr
	strOpts["g"] = *gPtr
	strOpts["admin"] = *adminPtr
	strOpts["db"] = *dbPtr
	strOpts["exp"] = *expPtr
	strOpts["imp"] = *impPtr
	strOpts["synch_client"] = *synchClientPtr
	strOpts["get_peer_token"] = *getPeerTokenPtr
	strOpts["save_peer_token"] = *savePeerTokenPtr
	strOpts["server_secret"] = *serverSecretPtr
	intfOpts["qi"] = *qiPtr
	intfOpts["l"] = *lPtr
	intfOpts["s"] = *sPtr
	intfOpts["ql"] = *qlPtr
	intfOpts["v"] = *vPtr
	intfOpts["whoami"] = *whoamiPtr
	intfOpts["del"] = *delPtr
	intfOpts["upd"] = *updPtr
	intfOpts["get_server_secret"] = *getServerSecretPtr
	intfOpts["setup_db"] = *setupDBPtr

	config.Opts.Verbose = *verbosePtr
	intfOpts["verbose"] = *verbosePtr

	config.Opts.Debug = *debugPtr
	intfOpts["debug"] = *debugPtr

	// This is the better way to pass options
	config.Opts.IsLocalWebSvr = *svrPtr
	config.Opts.IsSynchSvr = *synchServerPtr
	config.Opts.IsRemoteSvr = *ptrRemoteSvr

	separator := "/"
	if strings.Contains(strings.ToUpper(os.Getenv("OS")), "WINDOWS") {
		separator = "\\"
	}
	strOpts["sep"] = separator

	dbFile := "go_notes.sqlite"
	var dbFolder string
	var dbFullPath string
	if len(*dbPtr) == 0 {
		if len(os.Getenv("HOME")) > 0 {
			dbFolder = os.Getenv("HOME")
		} else if len(os.Getenv("HOMEDRIVE")) > 0 && len(os.Getenv("HOMEPATH")) > 0 {
			dbFolder = os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
		} else {
			dbFolder = separator /// last resort
		}
		dbFullPath = dbFolder + separator + dbFile
	} else {
		dbFullPath = *dbPtr
	}
	config.Opts.DBPath = dbFullPath

	return strOpts, intfOpts
}
