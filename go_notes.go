package main

import (
	"fmt"
    "strings"
    "bufio"
	"time"
    "os"
    "gotut.org/go_notes/options"
	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
    opts_str, opts_intf := options.Get()
    //println( "DBPath: " + opts_str["db_path"] )
    
	db, err := gorm.Open("sqlite3", opts_str["db_path"])
	if err != nil {
		println("There was an error connecting to the DB")
		os.Exit(2)
	}

	type Note struct {
		Id int64
		Title string `sql: "size:128"`
		Description string `sql: "size:128"`
		Body string `sql: "type:text"`
		CreatedAt time.Time
		UpdatedAt time.Time
	}

	if opts_str["admin"] == "delete_table" { db.DropTableIfExists(&Note{}) }
	db.AutoMigrate(&Note{}) // Feel free to change your struct, AutoMigrate will keep your database up-to-date.
// Fyi, AutoMigrate will only *add new columns*, it won't update column's type or delete unused columns, to make sure your data is safe.
// If the table is not existing, AutoMigrate will create the table automatically.
    
	if opts_str["q"] == "" && opts_intf["qi"].(int) == 0 { // Then try to Create
		if opts_str["t"] != "" {
			var chk_unique_title []Note
			db.Where("title = ?", opts_str["t"]).Find(&chk_unique_title)
			if len(chk_unique_title) > 0 {
				println("Error: Title", opts_str["t"], "is not unique!")
				return
			}
			print("Creating new note...")
			note2 := Note{Title: opts_str["t"], Description: opts_str["d"], Body: opts_str["b"]}
			db.Create(&note2)
			if ! db.NewRecord(note2) { fmt.Println("Record saved:", note2.Title) }
		} else {
			println("Title (-t \"A Title\") is required")
		}
	} else { // Query and possibly delete/update

		var notes []Note
		if opts_intf["qi"].(int) != 0 {
			db.Find(&notes, opts_intf["qi"].(int))
		} else if opts_str["q"] == "all" {
			db.Find(&notes)
		} else {
			db.Where("title LIKE ?", "%"+opts_str["q"]+"%").
                Or("description LIKE ?", "%"+opts_str["q"]+"%").
                Or("body LIKE ?", "%"+opts_str["q"]+"%").
                Limit(opts_intf["ql"].(int)).
                Find(&notes)
		}

		// Print notes found
        println("---------------------------------------------")
		for _, n := range notes {
			fmt.Printf("[%d] %s - %s\n", n.Id, n.Title, n.Description)
			if ! opts_intf["s"].(bool) { println(n.Body) }
			println("---------------------------------------------")
		}
        msg := "notes found"
        if len(notes) == 1 { msg = "note found" }
        println(len(notes), msg)

		// See if there was a delete
		if opts_intf["del"].(bool) {
			for _, n := range notes {
				save_id := n.Id
				println("---------------------------------------------")
				fmt.Printf("[%d] %s - %s\n", n.Id, n.Title, n.Description)
				print("Delete this note? (y/N) ")
				var input string
				fmt.Scanln(&input) // Get keyboard input
				if input == "y" || input == "Y" {
					db.Delete(&n)
					println("Note", save_id, "deleted")
				}
			}
		}
		// See if there was an update
		if  opts_intf["upd"].(bool) {
			for _, n := range notes {
				println("\n---------------------------------------------")
				fmt.Printf("[%d] %s - %s\n%s\n", n.Id, n.Title, n.Description, n.Body)
				print("Update this note? (y/N) ")
				var input string
				fmt.Scanln(&input) // Get keyboard input
				if input == "y" || input == "Y" {
                    reader := bufio.NewReader(os.Stdin)
                    println("\n" + n.Title)
                    fmt.Print("Enter New Title: (blank for no change) ")
                    tit, _ := reader.ReadString('\n')
                    tit = strings.TrimRight(tit, " \r\n")
                    if len(tit) > 0 { n.Title = tit }
                    println(n.Title)
                    
                    println("\n" + n.Description)
                    fmt.Print("Enter New Description: (blank for no change) ")
                    desc, _ := reader.ReadString('\n')
                    desc = strings.TrimRight(desc, " \r\n")
                    if len(desc) > 0 {n.Description = desc}
                    println(n.Description)
                    
                    println("\n" + n.Body)
                    fmt.Print("Enter New Body: (blank for no change) ")
                    body, _ := reader.ReadString('\n')
                    body = strings.TrimRight(body, " \r\n ")
                    if len(body) > 0 { n.Body = body }
                    println(n.Body)

                    db.Save(&n)
                    println("---------------------------------------------")
                    fmt.Printf("[%d] %s - %s\n%s", n.Id, n.Title, n.Description, n.Body)
				}
			}
		}
	}
}
