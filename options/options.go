package options
import (
	"os"
	"flag"
    "strings"
)

//Setup commandline options and other configuration for Go Notes
func Get() (map[string]string, map[string]interface{}) {
        opts_str := make( map[string]string )
        opts_intf := make( map[string]interface{} )
        
        qPtr := flag.String("q", "", "Query - Retrieve notes based on a LIKE search")
        tPtr := flag.String("t", "", "Title")
        dPtr := flag.String("d", "", "Description")
        bPtr := flag.String("b", "", "Body")
        gPtr := flag.String("g", "", "Tags - not yet implemented")
        adminPtr := flag.String("admin", "", "Privileged actions like 'delete_table'")
        dbPtr := flag.String("db", "", "Sqlite DB path")
        qiPtr := flag.Int("qi", 0, "Query - Retrieve notes based on index")
        qlPtr := flag.Int("ql", 9, "Limit the number of notes returned")
        sPtr := flag.Bool("s", false, "Short Listing - don't show the body")
        delPtr := flag.Bool("del", false, "Delete the notes queried")
        updPtr := flag.Bool("upd", false, "Update the notes queried")

        flag.Parse()
        opts_str["q"] = *qPtr
        opts_str["t"] = *tPtr
        opts_str["d"] = *dPtr
        opts_str["b"] = *bPtr
        opts_str["g"] = *gPtr
        opts_str["admin"] = *adminPtr
        opts_str["db"] = *dbPtr
        opts_intf["qi"] = *qiPtr
        opts_intf["ql"] = *qlPtr
        opts_intf["s"] =  *sPtr
        opts_intf["del"] = *delPtr
        opts_intf["upd"] = *updPtr
        
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
        //println("DEBUG: db_full_path", db_full_path)
            
        return opts_str, opts_intf
}