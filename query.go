package main

import(
	"fmt"
	"errors"
)

func getNote(guid string) (Note, error) {
	var note Note
	db.Where("guid = ?", guid).First(&note)
	if note.Id != 0 {
		return note, nil
	} else {
		return note, errors.New("Note not found")
	}
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

func find_note_by_id(id int64) (Note) {
	var note Note
	db.First(&note, id)
	return note
}

// Query by Id, return all notes, query all fields for one param, query a combination of fields and params
func queryNotes() []Note {
	var notes []Note
	db.LogMode(true)
	// The order of the if here is very important - esp. for the webserver!
	if opts_intf["ql"] == true {
		db.Order("updated_at desc").Limit(1).Find(&notes)
	} else if opts_intf["qi"] !=nil && opts_intf["qi"].(int64) != 0 {
		db.Find(&notes, opts_intf["qi"].(int64))
		// TAG and wildcard
	} else if opts_str["qg"] != "" && opts_str["q"] != "" {
		db.Where("tag LIKE ? AND (title LIKE ? OR description LIKE ? OR body LIKE ?)",
					"%"+opts_str["qg"]+"%",
					"%"+opts_str["q"]+"%",
					"%"+opts_str["q"]+"%",
					"%"+opts_str["q"]+"%",
		).Order("updated_at desc").Limit(opts_intf["l"].(int)).Find(&notes)
		// TITLE and wildcard
	} else if opts_str["qt"] != "" && opts_str["q"] != "" {
		db.Where("title LIKE ? AND (tag LIKE ? OR description LIKE ? OR body LIKE ?)",
					"%"+opts_str["qt"]+"%",
					"%"+opts_str["q"]+"%",
					"%"+opts_str["q"]+"%",
					"%"+opts_str["q"]+"%",
		).Order("updated_at desc").Limit(opts_intf["l"].(int)).Find(&notes)
		//
	} else if opts_str["q"] == "all" {
		db.Order("updated_at desc").Limit(opts_intf["l"].(int)).Find(&notes)
		// General query
	} else if opts_str["q"] != "" {
		db.Where("tag LIKE ? OR title LIKE ? OR description LIKE ? OR body LIKE ?",
					"%"+opts_str["q"]+"%",
					"%"+opts_str["q"]+"%",
					"%"+opts_str["q"]+"%",
					"%"+opts_str["q"]+"%",
		).Order("updated_at desc").Limit(opts_intf["l"].(int)).Find(&notes)
		// ANY combination - without q
	} else {
		db.Where("tag LIKE ? AND title LIKE ? AND description LIKE ? AND body LIKE ?",
					"%"+opts_str["qg"]+"%",
					"%"+opts_str["qt"]+"%",
					"%"+opts_str["qd"]+"%",
					"%"+opts_str["qb"]+"%",
		).Order("updated_at desc").Limit(opts_intf["l"].(int)).Find(&notes)
	}

	return notes
}

func listNotes(notes []Note, show_count bool) {
	pl(line_separator)
	for _, n := range notes {
		fmt.Printf("[%d] %s", n.Id, n.Title)
		if n.Description != "" {
			fmt.Printf(" - %s", n.Description)
		}
		pl("")
		if !opts_intf["s"].(bool) {
			if n.Body != "" {
				pl(n.Body)
			}
			if n.Tag != "" {
				pl("Tags:", n.Tag)
			}
		}
		pl(line_separator)
	}
	if show_count {
		var msg string // init'd to ""
		if len(notes) != 1 {
			msg = "s"
		}
		fmt.Printf("(%d note%s found)\n", len(notes), msg)
	}
}
