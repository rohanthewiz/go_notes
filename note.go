package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"errors"
	"time"
)

type Note struct {
	Id          int64
	Guid		string `sql: "size:40"` //Guid of the note
	Title       string `sql: "size:128"`
	Description string `sql: "size:255"`
	Body        string `sql: "type:text"`
	Tag         string `sql: "size:128"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

const line_separator string = "---------------------------------------------------------"

func createNote() {
	if opts_str["t"] != "" {
		var chk_unique_title []Note
		db.Where("title = ?", opts_str["t"]).Find(&chk_unique_title)
		if len(chk_unique_title) > 0 {
			println("Error: Title", opts_str["t"], "is not unique!")
			return
		}
		do_create( Note{Guid: generate_sha1(), Title: opts_str["t"], Description: opts_str["d"], Body: opts_str["b"], Tag: opts_str["g"]} )
	} else {
		println("Title (-t) is required if creating a note. Remember to precede option flags with '-'")
	}
}

// The core create method
func do_create(note Note) bool {
	print("Creating new note...")
	performNoteChange(
	NoteChange{
		Guid: generate_sha1(), Operation: 1,
		NoteGuid: note.Guid,
		Note: note,
		NoteFragment: NoteFragment{},
	})

	if n, err := getNote(note.Guid); err != nil {
		pf("Error creating note %v\n", note); return false
	} else {
		pf("Record saved: [%d] %s\n", n.Id, n.Title)
	}
	return true
}

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

// Query by Id, return all notes, query all fields for one param, query a combination of fields and params
func queryNotes() []Note {
	var notes []Note
	if opts_intf["qi"] !=nil && opts_intf["qi"].(int) != 0 { // TODO should we be checking options for nil first?
		db.Find(&notes, opts_intf["qi"].(int))
	} else if opts_str["q"] == "all" {
		db.Find(&notes)
	} else if opts_str["q"] != "" {
		db.Where("title LIKE ?", "%"+opts_str["q"]+"%").
		Or("description LIKE ?", "%"+opts_str["q"]+"%").
		Or("body LIKE ?", "%"+opts_str["q"]+"%").
		Or("tag LIKE ?", "%"+opts_str["q"]+"%").
		Limit(opts_intf["ql"].(int)).
		Find(&notes)
	} else {
		query := db.Limit(opts_intf["ql"].(int))
		query = query.Where("tag LIKE ?", "%" + opts_str["qg"] + "%")
		query = query.Where("title LIKE ?", "%" + opts_str["qt"] + "%")
		query = query.Where("description LIKE ?", "%" + opts_str["qd"] + "%")
		query = query.Where("body LIKE ?", "%" + opts_str["qb"] + "%")
		query.Find(&notes)
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
			var nf NoteFragment = NoteFragment{}

			println("\nTitle-->" + n.Title)
			fmt.Println("Enter new Title (or '+ blah' to append, or <ENTER> for no change)")
			tit, _ := reader.ReadString('\n')
			tit = strings.TrimRight(tit, " \r\n")

			orig_title := n.Title
			if len(tit) > 1 && tit[0:1] == "+" {
				n.Title += tit[1:]
			} else if len(tit) > 0 {
				n.Title = tit
			}
			if orig_title != n.Title { //Build NoteFragment
				nf.Title = n.Title
				nf.Bitmask |= 8
			}

			println("Description-->" + n.Description)
			fmt.Println("Enter new Description (or '-' to blank, '+ blah' to append, or <ENTER> for no change)")
			desc, _ := reader.ReadString('\n')
			desc = strings.TrimRight(desc, " \r\n")

			orig_desc := n.Description
			if desc == "-" {
				n.Description = ""
			} else if len(desc) > 1 && desc[0:1] == "+" {
				n.Description += desc[1:]
			} else if len(desc) > 0 {
				n.Description = desc
			}
			if orig_desc != n.Description { //Build NoteFragment
				nf.Description = n.Description
				nf.Bitmask |= 4
			}

			println("Body-->" + n.Body)
			fmt.Println("Enter new Body (or '-' to blank, '+ blah' to append, or <ENTER> for no change)")
			body, _ := reader.ReadString('\n')
			body = strings.TrimRight(body, " \r\n ")

			orig_body := n.Body
			if body == "-" {
				n.Body = ""
			} else if len(body) > 1 && body[0:1] == "+" {
				n.Body += body[1:]
			} else if len(body) > 0 {
				n.Body = body
			}
			if orig_body != n.Body { //Build NoteFragment
				nf.Body = n.Body
				nf.Bitmask |= 2
			}

			println("Tags-->" + n.Tag)
			fmt.Println("Enter new Tags (or '-' to blank, '+ blah' to append, or <ENTER> for no change)")
			tag, _ := reader.ReadString('\n')
			tag = strings.TrimRight(tag, " \r\n ")

			orig_tag := n.Tag
			if tag == "-" {
				n.Tag = ""
			} else if len(tag) > 1 && tag[0:1] == "+" {
				n.Tag += tag[1:]
			} else if len(tag) > 0 {
				n.Tag = tag
			}
			if orig_tag != n.Tag { //Build NoteFragment
				nf.Tag = n.Tag
				nf.Bitmask |= 1
			}

			db.Save(&n)
			nc := NoteChange{ Guid: generate_sha1(), NoteGuid: n.Guid, Operation: op_update, NoteFragment: nf }
			db.Save(&nc)
			if nc.Id > 0 {
				pf("NoteChange (%s) created successfully\n", short_sha(nc.Guid))
			}

			curr_note[0] = n
			listNotes(curr_note[:], false) // [:] means all of the slice
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
			doDelete(n)
			println("Note [", save_id, "] deleted")
		}
	}
}

func doDelete(note Note) {
	db.Delete(&note)
	nc := NoteChange{ Guid: generate_sha1(), NoteGuid: note.Guid, Operation: op_delete }
	db.Save(&nc)
	if nc.Id > 0 { // Hopefully nc was reloaded
		pf("NoteChange (%s) created successfully\n", short_sha(nc.Guid))
	}
}

