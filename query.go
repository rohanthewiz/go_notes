package main

import (
	"errors"
	"fmt"
	"go_notes/config"
	"go_notes/note"
)

func getNote(guid string) (note.Note, error) {
	var n note.Note
	db.Where("guid = ?", guid).First(&n)
	if n.Id != 0 {
		return n, nil
	} else {
		return n, errors.New("Note not found")
	}
}

func findNoteByTitle(title string) (bool, note.Note) {
	var notes []note.Note
	db.Where("title = ?", title).Limit(1).Find(&notes)
	if len(notes) == 1 {
		return true, notes[0]
	} else {
		return false, note.Note{} // yes this is the way you represent an empty Note object/struct
	}
}

func findNoteById(id int64) note.Note {
	var n note.Note
	db.First(&n, id)
	return n
}

// Query by Id, return all notes, query all fields for one param, query a combination of fields and params
func queryNotes() []note.Note {
	var notes []note.Note
	// The order of the if here is very important - esp. for the webserver!
	if optsIntf["ql"] == true {
		db.Order("updated_at desc").Limit(1).Find(&notes)
	} else if optsIntf["qi"] != nil && optsIntf["qi"].(int64) != 0 {
		db.Find(&notes, optsIntf["qi"].(int64))
		// TAG and wildcard
	} else if optsStr["qg"] != "" && optsStr["q"] != "" {
		db.Where("tag LIKE ? AND (title LIKE ? OR description LIKE ? OR body LIKE ?)",
			"%"+optsStr["qg"]+"%",
			"%"+optsStr["q"]+"%",
			"%"+optsStr["q"]+"%",
			"%"+optsStr["q"]+"%",
		).Order("updated_at desc").Limit(optsIntf["l"].(int)).Find(&notes)
		// TITLE and wildcard
	} else if optsStr["qt"] != "" && optsStr["q"] != "" {
		db.Where("title LIKE ? AND (tag LIKE ? OR description LIKE ? OR body LIKE ?)",
			"%"+optsStr["qt"]+"%",
			"%"+optsStr["q"]+"%",
			"%"+optsStr["q"]+"%",
			"%"+optsStr["q"]+"%",
		).Order("updated_at desc").Limit(optsIntf["l"].(int)).Find(&notes)
		//
	} else if optsStr["q"] == "all" {
		db.Order("updated_at desc").Limit(optsIntf["l"].(int)).Find(&notes)
		// General query
	} else if optsStr["q"] != "" {
		db.Where("tag LIKE ? OR title LIKE ? OR description LIKE ? OR body LIKE ?",
			"%"+optsStr["q"]+"%",
			"%"+optsStr["q"]+"%",
			"%"+optsStr["q"]+"%",
			"%"+optsStr["q"]+"%",
		).Order("updated_at desc").Limit(optsIntf["l"].(int)).Find(&notes)
		// ANY combination - without q
	} else {
		db.Where("tag LIKE ? AND title LIKE ? AND description LIKE ? AND body LIKE ?",
			"%"+optsStr["qg"]+"%",
			"%"+optsStr["qt"]+"%",
			"%"+optsStr["qd"]+"%",
			"%"+optsStr["qb"]+"%",
		).Order("updated_at desc").Limit(optsIntf["l"].(int)).Find(&notes)
	}
	// For now on remote webserver  filter out private notes
	if config.Opts.IsRemoteSvr {
		notes = note.FilterOutPrivate(notes)
	}

	return notes
}

func listNotes(notes []note.Note, showCount bool) {
	fmt.Println(LineSeparator)
	for _, n := range notes {
		fmt.Printf("[%d] %s", n.Id, n.Title)
		if n.Description != "" {
			fmt.Printf(" - %s", n.Description)
		}
		fmt.Println("")
		if !optsIntf["s"].(bool) {
			if n.Body != "" {
				fmt.Println(n.Body)
			}
			if n.Tag != "" {
				fmt.Println("Tags:", n.Tag)
			}
		}
		fmt.Println(LineSeparator)
	}
	if showCount {
		var msg string // init'd to ""
		if len(notes) != 1 {
			msg = "s"
		}
		fmt.Printf("(%d note%s found)\n", len(notes), msg)
	}
}
