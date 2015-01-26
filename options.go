package main
import (
	"os"
	"flag"
  "strings"
)

//Setup commandline options and other configuration for Go Notes
func getOpts() (map[string]string, map[string]interface{}) {
        opts_str := make( map[string]string )
        opts_intf := make( map[string]interface{} )
        
        qPtr := flag.String("q", "", "Query for notes based on a LIKE search. \"all\" will return all notes")
        qgPtr := flag.String("qg", "", "Query tags based on a LIKE search")
        tPtr := flag.String("t", "", "Create note Title")
        dPtr := flag.String("d", "", "Create note Description")
        bPtr := flag.String("b", "", "Create note Body")
        gPtr := flag.String("g", "", "Comma separated list of Tags for new note")
        adminPtr := flag.String("admin", "", "Privileged actions like 'delete_table'")
        dbPtr := flag.String("db", "", "Sqlite DB path")
        expPtr := flag.String("exp", "", "Export the notes queried to the format of the file given")
        impPtr := flag.String("imp", "", "Import the notes queried from the file given")
        synchClientPtr := flag.String("synch_client", "", "Synch client mode")

        qiPtr := flag.Int("qi", 0, "Query for notes based on ID")
        qlPtr := flag.Int("ql", -1, "Limit the number of notes returned")
        sPtr := flag.Bool("s", false, "Short Listing - don't show the body")
        vPtr := flag.Bool("v", false, "Show version")
        setupDBPtr := flag.Bool("setup_db", false, "Setup the Database")
        delPtr := flag.Bool("del", false, "Delete the notes queried")
        updPtr := flag.Bool("upd", false, "Update the notes queried")
        svrPtr := flag.Bool("svr", false, "Web server mode")
        synchServerPtr := flag.Bool("synch_server", false, "Synch server mode")

        flag.Parse()

        // Store options in a couple of maps
        opts_str["q"] = *qPtr
        opts_str["qg"] = *qgPtr
        opts_str["t"] = *tPtr
        opts_str["d"] = *dPtr
        opts_str["b"] = *bPtr
        opts_str["g"] = *gPtr
        opts_str["admin"] = *adminPtr
        opts_str["db"] = *dbPtr
        opts_str["exp"] = *expPtr
        opts_str["imp"] = *impPtr
        opts_str["synch_client"] = *synchClientPtr

        opts_intf["qi"] = *qiPtr
        opts_intf["ql"] = *qlPtr
        opts_intf["s"] =  *sPtr
        opts_intf["v"] =  *vPtr
        opts_intf["del"] = *delPtr
        opts_intf["upd"] = *updPtr
        opts_intf["svr"] = *svrPtr
        opts_intf["synch_server"] = *synchServerPtr
        opts_intf["setup_db"] = *setupDBPtr

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
