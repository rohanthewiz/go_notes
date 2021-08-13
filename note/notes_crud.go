package note

import (
	"bufio"
	"fmt"
	"go_notes/dbhandle"
	"go_notes/utils"
	"log"
	"os"
	"strings"
	"time"
)

const lineSeparator string = "---------------------------------------------------------"

func AllFieldsUpdate(nte Note) { // note is an unsaved note prepared with Id and all other fields even if not changed
	var orig Note
	dbhandle.DB.Where("id = ?", nte.Id).First(&orig) // get the original for comparision

	// Actual update
	dbhandle.DB.Table("notes").Where("id = ?", nte.Id).Updates(map[string]interface{}{
		"title": nte.Title, "description": nte.Description, "body": nte.Body, "tag": nte.Tag,
		"updated_at": time.Now(),
	})

	nf := NoteFragment{}
	if orig.Title != nte.Title { // Build NoteFragment
		nf.Title = nte.Title
		nf.Bitmask |= 8
	}
	if orig.Description != nte.Description { // Build NoteFragment
		nf.Description = nte.Description
		nf.Bitmask |= 4
	}
	if orig.Body != nte.Body { // Build NoteFragment
		nf.Body = nte.Body
		nf.Bitmask |= 2
	}
	if orig.Tag != nte.Tag { // Build NoteFragment
		nf.Tag = nte.Tag
		nf.Bitmask |= 1
	}
	nc := NoteChange{Guid: utils.GenerateSHA1(), NoteGuid: orig.Guid, Operation: OpUpdate, NoteFragment: nf}
	dbhandle.DB.Save(&nc)
	if nc.Id > 0 {
		utils.Pf("NoteChange (%s) created successfully\n", utils.ShortSHA(nc.Guid))
	}
}

func UpdateNotes(notes []Note) {
	for _, n := range notes {
		// curr_note[0] = n
		ListNotes([]Note{n}, false)
		print("Update this note? (y/N) ")
		var input string
		_, err := fmt.Scanln(&input)
		if err != nil {
			log.Println("Error while scanning input:", err)
		} // Get keyboard input
		if input == "y" || input == "Y" {
			reader := bufio.NewReader(os.Stdin)
			var nf = NoteFragment{}

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
			if origTag != n.Tag { // Build NoteFragment
				nf.Tag = n.Tag
				nf.Bitmask |= 1
			}

			dbhandle.DB.Save(&n)
			nc := NoteChange{Guid: utils.GenerateSHA1(), NoteGuid: n.Guid, Operation: OpUpdate, NoteFragment: nf}
			dbhandle.DB.Save(&nc)
			if nc.Id > 0 {
				utils.Pf("NoteChange (%s) created successfully\n", utils.ShortSHA(nc.Guid))
			}

			ListNotes([]Note{n}, false) // [:] means all of the slice
		}
	}
}

func DeleteNotes(notes []Note) {
	for _, n := range notes {
		save_id := n.Id
		ListNotes([]Note{n}, false)
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

func DoDelete(nte Note) {
	if nte == (Note{}) {
		utils.Pf("Internal error: cannot delete non-existent note")
		return
	}
	dbhandle.DB.Delete(&nte)
	nc := NoteChange{Guid: utils.GenerateSHA1(), NoteGuid: nte.Guid, Operation: OpDelete}
	dbhandle.DB.Save(&nc)
	if nc.Id > 0 { // Hopefully nc was reloaded
		utils.Pf("NoteChange (%s) created successfully\n", utils.ShortSHA(nc.Guid))
	}
}
