package main

import (
	"bufio"
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"
	"gotut.org/go_notes/options"
	"os"
	"strings"
	"time"
	"encoding/csv"
)

const app_name = "GoNotes"
const version string = "0.8.5"
const line_separator string = "---------------------------------------------------------"
const out_csv = "output.csv"

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
var opts_str, opts_intf = options.Get() //returns map[string]string, map[string]interface{}
// Init db
var db, err = gorm.Open("sqlite3", opts_str["db_path"])

func main() {
	if err != nil { // Can't err chk db conn outside method, so do it here
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
	if opts_str["q"] == "" && opts_intf["qi"].(int) == 0 && opts_str["qg"] == "" {
		createNote() // No query options, we must be trying to CREATE
	} else {
		// QUERY
		notes := queryNotes()

		// List Notes found
		listNotes(notes, true)

		// See if we want to export
		if opts_str["exp"] != "" {
			arr := strings.Split(opts_str["exp"], ".")
			arr_item_last := len(arr) -1
			if arr[arr_item_last] == "csv" {
				exportCsv(notes)
			}
//			if arr[arr_item_last] == "gob" {
//				exportGob(notes)
//			}

		// See if we want to update
		} else if opts_intf["upd"].(bool) {
			updateNotes(notes)

			// See if we want to delete
		} else if opts_intf["del"].(bool) {
			deleteNotes(notes)
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
		print("Creating new note...")
		note2 := Note{Title: opts_str["t"], Description: opts_str["d"], Body: opts_str["b"], Tag: opts_str["g"]}
		db.Create(&note2)
		if !db.NewRecord(note2) {
			fmt.Println("Record saved:", note2.Title)
		}
	} else {
		println("Title (-t) is required if creating a note. Remember to precede option flags with '-'")
	}
}

func queryNotes() []Note {
	var notes []Note
	if opts_intf["qi"].(int) != 0 {
		db.Find(&notes, opts_intf["qi"].(int))
	} else if opts_str["qg"] != "" {
		db.Where("tag LIKE ?", "%"+opts_str["qg"]+"%").
			Limit(opts_intf["ql"].(int)).Find(&notes)
	} else if opts_str["q"] == "all" {
		db.Find(&notes)
	} else if opts_str["q"] != "" {
		db.Where("title LIKE ?", "%"+opts_str["q"]+"%").
			Or("description LIKE ?", "%"+opts_str["q"]+"%").
			Or("body LIKE ?", "%"+opts_str["q"]+"%").
			Or("tag LIKE ?", "%"+opts_str["q"]+"%").
			Limit(opts_intf["ql"].(int)).
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

func exportCsv(notes []Note) {
	csv_file, err := os.Create(out_csv)
	if err != nil { fmt.Println("Error: ", err); return }
//	defer csv_file.Close()
	writer := csv.NewWriter(csv_file)

	for _, n := range notes {
		arr := []string{n.Title, n.Description, n.Body, n.Tag}
		err := writer.Write(arr)
		if err != nil { fmt.Println("Error: ", err); return }
	}
	writer.Flush()
	println("Exported to", out_csv)

	csv_file.Close()

	// Base check of importNotes
	println("Reading the csv back in...")
	importNotes(out_csv)
}

func importNotes(in_file string) {
	csvfile, err := os.Open(in_file)
	if err != nil { fmt.Println("Error: ", err); return }
	defer csvfile.Close()

	reader := csv.NewReader(csvfile)
	reader.FieldsPerRecord = -1 // Todo match this with Note struct

	rawCSVdata, err := reader.ReadAll()
	if err != nil { fmt.Println("Error: ", err); return }

	// sanity check, display to standard output
	for _, f := range rawCSVdata {
		fmt.Printf("Title : %s - Desc: %s\nBody: %s\nTags: %s\n", f[0], f[1], f[2], f[3])
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

			println("\nTitle--> \"" + n.Title + "\"")
			fmt.Println("Enter new Title (or '+ blah' to append, or <ENTER> for no change)")
			tit, _ := reader.ReadString('\n')
			tit = strings.TrimRight(tit, " \r\n")
			if len(tit) > 1 && tit[0:1] == "+" {
				n.Title = n.Title + tit[1:]
			} else if len(tit) > 0 {
				n.Title = tit
			}

			println("Description--> \"" + n.Description + "\"")
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

			println("Body--> \"" + n.Body + "\"")
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

			println("Tags--> \"" + n.Tag + "\"")
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
