package main

import (
	"os"
	"fmt"
	"flag"
	"time"
	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	queryPtr := flag.String("q", "", "Query - Retrieve notes based on a LIKE search")
	queryIndexPtr := flag.Int("qi", 0, "Query - Retrieve notes based on index")
	limitPtr := flag.Int("l", 9, "Limit the number of notes returned")
	delPtr := flag.String("del", "n", "Delete the notes listed")
	titlePtr := flag.String("t", "", "Title")
	descPtr := flag.String("d", "", "Description")
	bodyPtr := flag.String("b", "", "Body")
	tagsPtr := flag.String("g", "", "Tags - not yet implemented")
	shortPtr := flag.Bool("s", false, "Short Listing - don't show the body")
	adminPtr := flag.String("admin", "", "Privileged actions like 'delete_table'")
	dbPtr := flag.String("db", "", "Sqlite DB path")
	var home string
	if len(*dbPtr) == 0 { 
		home = os.Getenv("HOME")
		if len(home) > 0 {
			home += "/go_notes.sqlite"
		} else {
			home = "/Users/rohan/db/go_notes.sqlite"
		}
		dbPtr = &home
	}
	flag.Parse()

	// For now make sure all vars are used
	if *tagsPtr != "" { fmt.Println("Tags: ", *tagsPtr) }

	db, err := gorm.Open("sqlite3", *dbPtr)
	if err != nil {
		println("There was an error connecting to the DB")
		return
	}
	// conn := db.DB()
	// conn.Ping()

	type Note struct {
		Id int64
		Title string `sql: "size:128"`
		Description string `sql: "size:128"`
		Body string `sql: "type:text"`
		CreatedAt time.Time
		UpdatedAt time.Time
	}

	if *adminPtr == "delete_table" { db.DropTableIfExists(&Note{}) }
	// println(db.HasTable("notes"))
	db.AutoMigrate(&Note{}) // Feel free to change your struct, AutoMigrate will keep your database up-to-date.
// Fyi, AutoMigrate will only *add new columns*, it won't update column's type or delete unused columns, to make sure your data is safe.
// If the table is not existing, AutoMigrate will create the table automatically.

	if *queryPtr == "" && *queryIndexPtr == 0 { // Then try to Create
		if *titlePtr != "" {
			var chk_unique_title []Note
			db.Where("title = ?", *titlePtr).Find(&chk_unique_title)
			if len(chk_unique_title) > 0 {
				println("Error: Title", *titlePtr, "is not unique!")
				return
			}
			// Create new note
			print("Creating new note...")
			note2 := Note{Title: *titlePtr, Description: *descPtr, Body: *bodyPtr}
			db.Create(&note2)
			if ! db.NewRecord(note2) { fmt.Println("Record saved:", note2.Title) }
		} else {
			println("Title (-t=\"A Title\") is required")
		}
	} else { // Query and possibly delete

		var notes []Note
		if *queryIndexPtr != 0 {
			db.Find(&notes, *queryIndexPtr)
		} else if *queryPtr == "all" {
			db.Find(&notes)
		} else {
			db.Where("title LIKE ?", "%"+*queryPtr+"%").Or("description LIKE ?", "%"+*queryPtr+"%").Or("body LIKE ?", "%"+*queryPtr+"%").Limit(*limitPtr).Find(&notes)
		}

		// Print notes found
		for _, n := range notes {
			fmt.Printf("[%d] %s - %s\n", n.Id, n.Title, n.Description)
			if ! *shortPtr { println(n.Body) }
			println("---------------------------------------------")
		}

		// See if there was a delete
		if *delPtr == "yes" || *delPtr == "y" {
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
	}
}
