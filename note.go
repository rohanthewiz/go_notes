package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"
)

type Note struct {
	Id          uint64
	Guid        string `sql: "size:40"` //Guid of the note
	Title       string `sql: "size:128"`
	Description string `sql: "size:255"`
	Body        string `sql: "type:text"`
	Tag         string `sql: "size:128"`
	User        string // who's account is this currently in (GUID) //todo - Add Index
	Creator     string // (GUID) who originally created the note
	SharedBy    string // (GUID) if it was shared to me, by who?
	Public      bool   // Was it made public for all users
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

const line_separator string = "---------------------------------------------------------"

func createNote(title string, desc string, body string, tag string) uint64 {
	if title != "" {
		var chk_unique_title []Note
		db.Where("title = ?", title).Find(&chk_unique_title)
		if len(chk_unique_title) > 0 {
			fpl("Error: Title", title, "is not unique!")
			return 0
		}
		return do_create(Note{Guid: generate_sha1(), Title: title, Description: desc,
			Body: body, Tag: tag})
	} else {
		fpl("Title (-t) is required if creating a note. Remember to precede option flags with '-'")
	}
	return 0
}

// The core create method
func do_create(note Note) uint64 {
	print("Creating new note...")
	performNoteChange(
		NoteChange{
			Guid: generate_sha1(), Operation: 1,
			NoteGuid:     note.Guid,
			Note:         note,
			NoteFragment: NoteFragment{},
		})

	if n, err := getNote(note.Guid); err != nil {
		pf("Error creating note %v\n", note)
		return 0
	} else {
		pf("Record saved: [%d] %s\n", n.Id, n.Title)
		return n.Id
	}
	return 0
}

func allFieldsUpdate(note Note) { // note is an unsaved note prepared with Id and all other fields even if not changed
	var orig Note
	db.Where("id = ?", note.Id).First(&orig) // get the original for comparision
	// Actual update
	db.Table("notes").Where("id = ?", note.Id).Updates(map[string]interface{}{
		"title": note.Title, "description": note.Description, "body": note.Body, "tag": note.Tag,
	})
	var nf NoteFragment = NoteFragment{}
	if orig.Title != note.Title { //Build NoteFragment
		nf.Title = note.Title
		nf.Bitmask |= 8
	}
	if orig.Description != note.Description { //Build NoteFragment
		nf.Description = note.Description
		nf.Bitmask |= 4
	}
	if orig.Body != note.Body { //Build NoteFragment
		nf.Body = note.Body
		nf.Bitmask |= 2
	}
	if orig.Tag != note.Tag { //Build NoteFragment
		nf.Tag = note.Tag
		nf.Bitmask |= 1
	}
	nc := NoteChange{Guid: generate_sha1(), NoteGuid: orig.Guid, Operation: op_update, NoteFragment: nf}
	db.Save(&nc)
	if nc.Id > 0 {
		pf("NoteChange (%s) created successfully\n", short_sha(nc.Guid))
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

			fpl("\nTitle-->" + n.Title)
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

			fpl("Description-->" + n.Description)
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

			fpl("Body-->" + n.Body)
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

			fpl("Tags-->" + n.Tag)
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
			nc := NoteChange{Guid: generate_sha1(), NoteGuid: n.Guid, Operation: op_update, NoteFragment: nf}
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
			fpl("Note [", save_id, "] deleted")
		}
	}
}

func doDelete(note Note) {
	if note == (Note{}) {
		pf("Internal error: cannot delete non-existent note")
		return
	}
	db.Delete(&note)
	nc := NoteChange{Guid: generate_sha1(), NoteGuid: note.Guid, Operation: op_delete}
	db.Save(&nc)
	if nc.Id > 0 { // Hopefully nc was reloaded
		pf("NoteChange (%s) created successfully\n", short_sha(nc.Guid))
	}
}
