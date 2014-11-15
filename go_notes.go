package main

import (
	"fmt"
	"flag"
	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	queryPtr := flag.String("q", "most_recent", "Query - read operation")
	limitPtr := flag.Int("l", 5, "Limit")
	titlePtr := flag.String("t", "", "Title")
	descPtr := flag.String("d", "", "Description")
	bodyPtr := flag.String("b", "", "Body")
	tagsPtr := flag.String("g", "", "Tags")
	dbPtr := flag.String("db", "/Users/rohan/db/notes.sqlite", "Sqlite DB path")
	flag.Parse()

	fmt.Println("Query: ", *queryPtr)
	fmt.Println("Limit: ", *limitPtr)
	fmt.Println("Title: ", *titlePtr)
	fmt.Println("Description: ", *descPtr)
	fmt.Println("Body: ", *bodyPtr)
	fmt.Println("Tags: ", *tagsPtr)
	fmt.Println("DB path: ", *dbPtr)

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
	}

	db.DropTableIfExists(&Note{})
	db.CreateTable(&Note{})

	note := Note{Title: "First Note", Description: "This is my first note", Body: "Notes are good things"}
	db.Create(&note)
	if db.NewRecord(note) { fmt.Println("Failed to save note", note.Title) }

	note2 := Note{Title: *titlePtr, Description: *descPtr, Body: *bodyPtr}
	db.Create(&note2)
	if ! db.NewRecord(note2) { fmt.Println("Record saved:", note2.Title) }

	// var note_retrieved Note
	// db.First(&note_retrieved)
	// fmt.Println("Here is the first note retrieved: \n", note_retrieved.Title)

	var notes []Note
	db.Limit(*limitPtr).Find(&notes)

	fmt.Println("All notes:")
	for _, n := range notes { println("Title:", n.Title) }

}
