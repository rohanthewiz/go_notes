package main

import (
	"bufio"
	"fmt"
	"go_notes/dbhandle"
	"go_notes/note"
	"go_notes/note/note_change"
	"go_notes/utils"
	"log"
	"os"
	"strings"
	"time"
)

const LineSeparator string = "---------------------------------------------------------"

func CreateNote(title string, desc string, body string, tag string) uint64 {
	if title != "" {
		var notes []note.Note
		dbhandle.DB.Where("title = ?", title).Find(&notes)
		if len(notes) > 0 {
			_, _ = fmt.Println("Error: Title", title, "is not unique!")
			return 0
		}
		return DoCreate(note.Note{Guid: utils.GenerateSHA1(), Title: title, Description: desc,
			Body: body, Tag: tag})
	} else {
		_, _ = fmt.Println("Title (-t) is required if creating a note. Remember to precede option flags with '-'")
	}
	return 0
}

// The core create method
func DoCreate(nte note.Note) (id uint64) {
	utils.Pl("Creating new note...")
	performNoteChange(
		note_change.NoteChange{
			Guid: utils.GenerateSHA1(), Operation: 1,
			NoteGuid:     nte.Guid,
			Note:         nte,
			NoteFragment: note_change.NoteFragment{},
		})

	if n, err := getNote(nte.Guid); err != nil {
		utils.Pf("Error creating note %v\n", nte)
		return 0
	} else {
		utils.Pf("Record saved: [%d] %s\n", n.Id, n.Title)
		id = n.Id
	}
	return id
}

func AllFieldsUpdate(nte note.Note) { // note is an unsaved note prepared with Id and all other fields even if not changed
	var orig note.Note
	dbhandle.DB.Where("id = ?", nte.Id).First(&orig) // get the original for comparision
	// Actual update
	dbhandle.DB.Table("notes").Where("id = ?", nte.Id).Updates(map[string]interface{}{
		"title": nte.Title, "description": nte.Description, "body": nte.Body, "tag": nte.Tag,
		"updated_at": time.Now(),
	})
	var nf note_change.NoteFragment = note_change.NoteFragment{}
	if orig.Title != nte.Title { //Build NoteFragment
		nf.Title = nte.Title
		nf.Bitmask |= 8
	}
	if orig.Description != nte.Description { //Build NoteFragment
		nf.Description = nte.Description
		nf.Bitmask |= 4
	}
	if orig.Body != nte.Body { //Build NoteFragment
		nf.Body = nte.Body
		nf.Bitmask |= 2
	}
	if orig.Tag != nte.Tag { //Build NoteFragment
		nf.Tag = nte.Tag
		nf.Bitmask |= 1
	}
	nc := note_change.NoteChange{Guid: utils.GenerateSHA1(), NoteGuid: orig.Guid, Operation: note_change.OpUpdate, NoteFragment: nf}
	dbhandle.DB.Save(&nc)
	if nc.Id > 0 {
		utils.Pf("NoteChange (%s) created successfully\n", utils.ShortSHA(nc.Guid))
	}
}

func UpdateNotes(notes []note.Note) {
	for _, n := range notes {
		// curr_note[0] = n
		listNotes([]note.Note{n}, false)
		print("Update this note? (y/N) ")
		var input string
		_, err := fmt.Scanln(&input)
		if err != nil {
			log.Println("Error while scanning input:", err)
		} // Get keyboard input
		if input == "y" || input == "Y" {
			reader := bufio.NewReader(os.Stdin)
			var nf = note_change.NoteFragment{}

			_, _ = fmt.Println("\nTitle-->" + n.Title)
			fmt.Println("Enter new Title (or '+ blah' to append, or <ENTER> for no change)")
			tit, _ := reader.ReadString('\n')
			tit = strings.TrimRight(tit, " \r\n")

			origTitle := n.Title
			if len(tit) > 1 && tit[0:1] == "+" {
				n.Title += tit[1:]
			} else if len(tit) > 0 {
				n.Title = tit
			}
			if origTitle != n.Title { // Build NoteFragment
				nf.Title = n.Title
				nf.Bitmask |= 8
			}

			_, _ = fmt.Println("Description-->" + n.Description)
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
			if orig_desc != n.Description { // Build NoteFragment
				nf.Description = n.Description
				nf.Bitmask |= 4
			}

			fmt.Println("Body-->" + n.Body)
			fmt.Println("Enter new Body (or '-' to blank, '+ blah' to append, or <ENTER> for no change)")
			body, _ := reader.ReadString('\n')
			body = strings.TrimRight(body, " \r\n ")

			origBody := n.Body
			if body == "-" {
				n.Body = ""
			} else if len(body) > 1 && body[0:1] == "+" {
				n.Body += body[1:]
			} else if len(body) > 0 {
				n.Body = body
			}
			if origBody != n.Body { // Build NoteFragment
				nf.Body = n.Body
				nf.Bitmask |= 2
			}

			fmt.Println("Tags-->" + n.Tag)
			fmt.Println("Enter new Tags (or '-' to blank, '+ blah' to append, or <ENTER> for no change)")
			tag, _ := reader.ReadString('\n')
			tag = strings.TrimRight(tag, " \r\n ")

			origTag := n.Tag
			if tag == "-" {
				n.Tag = ""
			} else if len(tag) > 1 && tag[0:1] == "+" {
				n.Tag += tag[1:]
			} else if len(tag) > 0 {
				n.Tag = tag
			}
			if origTag != n.Tag { //Build NoteFragment
				nf.Tag = n.Tag
				nf.Bitmask |= 1
			}

			dbhandle.DB.Save(&n)
			nc := note_change.NoteChange{Guid: utils.GenerateSHA1(), NoteGuid: n.Guid, Operation: note_change.OpUpdate, NoteFragment: nf}
			dbhandle.DB.Save(&nc)
			if nc.Id > 0 {
				utils.Pf("NoteChange (%s) created successfully\n", utils.ShortSHA(nc.Guid))
			}

			listNotes([]note.Note{n}, false) // [:] means all of the slice
		}
	}
}

func DeleteNotes(notes []note.Note) {
	for _, n := range notes {
		save_id := n.Id
		listNotes([]note.Note{n}, false)
		print("Delete this note? (y/N) ")
		var input string
		_, err := fmt.Scanln(&input) // Get keyboard input
		if err != nil {
			log.Println("Error scanning cmd line:", err)
		}
		if input == "y" || input == "Y" {
			DoDelete(n)
			fmt.Println("Note [", save_id, "] deleted")
		}
	}
}

func DoDelete(nte note.Note) {
	if nte == (note.Note{}) {
		utils.Pf("Internal error: cannot delete non-existent note")
		return
	}
	dbhandle.DB.Delete(&nte)
	nc := note_change.NoteChange{Guid: utils.GenerateSHA1(), NoteGuid: nte.Guid, Operation: note_change.OpDelete}
	dbhandle.DB.Save(&nc)
	if nc.Id > 0 { // Hopefully nc was reloaded
		utils.Pf("NoteChange (%s) created successfully\n", utils.ShortSHA(nc.Guid))
	}
}
