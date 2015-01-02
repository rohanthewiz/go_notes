package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
	"encoding/csv"
	"encoding/gob"
	"net/http"
	"github.com/julienschmidt/httprouter"
	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"
)

const app_name = "GoNotes"
const version string = "0.8.8"
const line_separator string = "---------------------------------------------------------"

type Note struct {
	Id          int64
	Title       string `sql: "size:128"`
	Description string `sql: "size:255"`
	Body        string `sql: "type:text"`
	Tag         string `sql: "size:128"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Get Commandline Options and Flags
var opts_str, opts_intf = getOpts() //returns map[string]string, map[string]interface{}

// Init db // Todo - pass db instead of making it static
var db, db_err = gorm.Open("sqlite3", opts_str["db_path"])

// Handlers for httprouter
func Index(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
	fmt.Fprint(w, "Welcome!\n")
}

//func Hello(w http.ResponseWriter, _ *http.Request, p httprouter.Params) {
//	fmt.Fprintf(w, "Hello, %s!\n", p.ByName("name"))
//}

func Query(w http.ResponseWriter, _ *http.Request, p httprouter.Params) {
	opts_str["q"] = p.ByName("query")  // Overwrite the query param
	notes := queryNotes(opts_str, opts_intf )
	RenderQuery(w, notes)
}

func main() {
	if db_err != nil { // Can't err chk db conn outside method, so do it here
		println("There was an error connecting to the DB")
		println("DBPath: " + opts_str["db_path"])
		os.Exit(2)
	}

	if opts_intf["v"].(bool) {
		println(app_name, version)
		return
	}

	if opts_str["admin"] == "delete_table" {
		db.DropTableIfExists(&Note{})
		println("notes table deleted")
		return
	}

	// Create or update the table structure as needed
	db.AutoMigrate(&Note{}) //According to GORM: Feel free to change your struct, AutoMigrate will keep your database up-to-date.
	// Fyi, AutoMigrate will only *add new columns*, it won't update column's type or delete unused columns, to make sure your data is safe.
	// If the table is not existing, AutoMigrate will create the table automatically.

	// CORE PROCESSING
	if opts_intf["svr"].(bool) {
		router := httprouter.New()
		router.GET("/", Index)
		router.GET("/hello/:name", Hello)
		router.GET("/q/:query", Query)
		println("Server listening on 8080... Ctrl-C to quit")
		log.Fatal(http.ListenAndServe(":8080", router))

	} else if opts_str["t"] != "" { // No query options, we must be trying to CREATE
		createNote()

	} else if opts_str["q"] != "" || opts_intf["qi"].(int) != 0 || opts_str["qg"] != ""{
		// QUERY
		notes := queryNotes(opts_str, opts_intf)

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

func createNote() {
	if opts_str["t"] != "" {
		var chk_unique_title []Note
		db.Where("title = ?", opts_str["t"]).Find(&chk_unique_title)
		if len(chk_unique_title) > 0 {
			println("Error: Title", opts_str["t"], "is not unique!")
			return
		}
		do_create( Note{Title: opts_str["t"], Description: opts_str["d"], Body: opts_str["b"], Tag: opts_str["g"]} )
	} else {
		println("Title (-t) is required if creating a note. Remember to precede option flags with '-'")
	}
}

// The core create method
func do_create(note Note) bool {
	print("Creating new note...")
	db.Create(&note)
	if !db.NewRecord(note) { // was it saved?
		println("Record saved:", note.Title)
		return true
	}
	println("Failed to save:", note.Title)
	return false
}

func find_note_by_title(title string) (bool, Note) {
	var notes []Note
	db.Where("title = ?", title).Limit(1).Find(&notes)
	if len(notes) == 1 {
		return true, notes[0]
	} else {
		return false, Note{} // yes this is the way you represent an empty Note object/struct
	}
}

func exportCsv(notes []Note, out_file string) {
	csv_file, err := os.Create(out_file)
	if err != nil { fmt.Println("Error: ", err); return }
	defer csv_file.Close()
	writer := csv.NewWriter(csv_file)

	for _, n := range notes {
		arr := []string{n.Title, n.Description, n.Body, n.Tag}
		err := writer.Write(arr)
		if err != nil { fmt.Println("Error: ", err); return }
	}
	writer.Flush()
	println("Exported to", out_file)
}

func exportGob(notes []Note, out_file string) {
	gob_file, err := os.Create(out_file)
	if err != nil { fmt.Println("Error: ", err); return }
	defer gob_file.Close()
	gob_encoder := gob.NewEncoder(gob_file)

	err = gob_encoder.Encode(notes)
	if err != nil { fmt.Println("Error: ", err); return }
	println("Exported to", out_file)
}

func importGob(in_file string) {
	var notes []Note

	gob_file, err := os.Open(in_file)
	if err != nil { fmt.Println("Error: ", err); return }
	defer gob_file.Close()

	decoder := gob.NewDecoder(gob_file)
	err = decoder.Decode(&notes)
	if err != nil { fmt.Println("Error: ", err); return }
	listNotes(notes, false)
	fmt.Printf("%d note(s) retrieved from %s\n", len(notes), in_file)

	// Update, create or discard?
	for _, n := range notes {
		exists, note := find_note_by_title(n.Title)
		if exists {
			println("This note already exists: ", note.Title)
			if n.UpdatedAt.After(note.UpdatedAt) {
				println("The imported note is newer, updating...")
				note.Description = n.Description
				note.Body = n.Body
				note.Tag = n.Tag
				db.Save(&note)
				listNotes([]Note{note}, false) // [:] means all of the slice
			}	else { println("The imported note is not newer, ignoring...")}
		} else {
			do_create( Note{ Title: n.Title, Description: n.Description, Body: n.Body, Tag: n.Tag } )
			fmt.Printf("Created -->Title: %s - Desc: %s\nBody: %s\nTags: %s\n", n.Title, n.Description, n.Body, n.Tag)
		}
	}
}

func importCsv(in_file string) {
	csv_file, err := os.Open(in_file)
	if err != nil { fmt.Println("Error: ", err); return }
	defer csv_file.Close()

	reader := csv.NewReader(csv_file)
	reader.FieldsPerRecord = -1 // Todo match this with Note struct

	rawCSVdata, err := reader.ReadAll()
	if err != nil { fmt.Println("Error: ", err); return }

	// sanity check, display to standard output
	for _, f := range rawCSVdata {
		exists, note := find_note_by_title(f[0])
		if exists {
			println("This note already exists: ", note.Title)
			// we could check an 'update on import' option here, set the corresponding fields, then save
			// or we could decide to update based on last_updated, but the export would have to save updated times - this would be a gob format
		} else {
			do_create( Note{Title: f[0], Description: f[1], Body: f[2], Tag: f[3]} )
			fmt.Printf("Created -->Title: %s - Desc: %s\nBody: %s\nTags: %s\n", f[0], f[1], f[2], f[3])
		}
	}
}

func deleteNotes(notes []Note) {
	var curr_note [1]Note //array since listNotes takes a slice
	for _, n := range notes {
		save_id := n.Id
		curr_note[0] = n
		listNotes(curr_note[0:1], false)
		print("Delete this note? (y/N) ")
		var input string
		fmt.Scanln(&input) // Get keyboard input
		if input == "y" || input == "Y" {
			db.Delete(&n)
			println("Note", save_id, "deleted")
		}
	}
}

func updateNotes(notes []Note) {
	var curr_note [1]Note //array since listNotes takes a slice
	for _, n := range notes {
		curr_note[0] = n
		listNotes(curr_note[0:1], false) //pass a slice of the array
		print("Update this note? (y/N) ")
		var input string
		fmt.Scanln(&input) // Get keyboard input
		if input == "y" || input == "Y" {
			reader := bufio.NewReader(os.Stdin)

			println("\nTitle-->" + n.Title)
			fmt.Println("Enter new Title (or '+ blah' to append, or <ENTER> for no change)")
			tit, _ := reader.ReadString('\n')
			tit = strings.TrimRight(tit, " \r\n")
			if len(tit) > 1 && tit[0:1] == "+" {
				n.Title = n.Title + tit[1:]
			} else if len(tit) > 0 {
				n.Title = tit
			}

			println("Description-->" + n.Description)
			fmt.Println("Enter new Description (or '-' to blank, '+ blah' to append, or <ENTER> for no change)")
			desc, _ := reader.ReadString('\n')
			desc = strings.TrimRight(desc, " \r\n")
			if desc == "-" {
				n.Description = ""
			} else if len(desc) > 1 && desc[0:1] == "+" {
				n.Description = n.Description + desc[1:]
			} else if len(desc) > 0 {
				n.Description = desc
			}

			println("Body-->" + n.Body)
			fmt.Println("Enter new Body (or '-' to blank, '+ blah' to append, or <ENTER> for no change)")
			body, _ := reader.ReadString('\n')
			body = strings.TrimRight(body, " \r\n ")
			if body == "-" {
				n.Body = ""
			} else if len(body) > 1 && body[0:1] == "+" {
				n.Body = n.Body + body[1:]
			} else if len(body) > 0 {
				n.Body = body
			}

			println("Tags-->" + n.Tag)
			fmt.Println("Enter new Tags (or '-' to blank, '+ blah' to append, or <ENTER> for no change)")
			tag, _ := reader.ReadString('\n')
			tag = strings.TrimRight(tag, " \r\n ")
			if tag == "-" {
				n.Tag = ""
			} else if len(tag) > 1 && tag[0:1] == "+" {
				n.Tag = n.Tag + tag[1:]
			} else if len(tag) > 0 {
				n.Tag = tag
			}

			db.Save(&n)
			curr_note[0] = n
			listNotes(curr_note[:], false) // [:] means all of the slice
		}
	}
}

func queryNotes(str_options map[string]string, intf_options map[string]interface{}) []Note {
	var notes []Note
	if intf_options["qi"] !=nil && intf_options["qi"].(int) != 0 { // TODO should we be checking options for nil first?
		db.Find(&notes, intf_options["qi"].(int))
	} else if str_options["qg"] != "" {
		db.Where("tag LIKE ?", "%"+str_options["qg"]+"%").
		Limit(intf_options["ql"].(int)).Find(&notes)
	} else if str_options["q"] == "all" {
		db.Find(&notes)
	} else if str_options["q"] != "" {
		db.Where("title LIKE ?", "%"+str_options["q"]+"%").
		Or("description LIKE ?", "%"+str_options["q"]+"%").
		Or("body LIKE ?", "%"+str_options["q"]+"%").
		Or("tag LIKE ?", "%"+str_options["q"]+"%").
		Limit(intf_options["ql"].(int)).
		Find(&notes)
	}
	return notes
}

func listNotes(notes []Note, show_count bool) {
	println(line_separator)
	for _, n := range notes {
		fmt.Printf("[%d] %s", n.Id, n.Title)
		if n.Description != "" {
			fmt.Printf(" - %s", n.Description)
		}
		println("")
		if !opts_intf["s"].(bool) {
			if n.Body != "" {
				println(n.Body)
			}
			if n.Tag != "" {
				println("Tags:", n.Tag)
			}
		}
		println(line_separator)
	}
	if show_count {
		var msg string // init'd to ""
		if len(notes) != 1 {
			msg = "s"
		}
		fmt.Printf("(%d note%s found)\n", len(notes), msg)
	}
}
