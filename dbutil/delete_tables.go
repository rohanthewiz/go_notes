package dbutil

import (
	"fmt"
	db "go_notes/dbhandle"
	"go_notes/localsig"
	"go_notes/note"
	"go_notes/note/note_change"
	"go_notes/peer"
	"go_notes/utils"
)

func DeleteTables() {
	fmt.Println("Are you sure you want to delete all data? (N/y)")
	var input string
	_, err := fmt.Scanln(&input) // Get keyboard input
	if err != nil {
		return
	}
	utils.Pd("input", input)
	if input == "y" || input == "Y" {
		db.DB.DropTableIfExists(&note.Note{})
		db.DB.DropTableIfExists(&note_change.NoteChange{})
		db.DB.DropTableIfExists(&note_change.NoteFragment{})
		db.DB.DropTableIfExists(&peer.Peer{})
		db.DB.DropTableIfExists(&localsig.LocalSig{})
		utils.Pl("Notes tables deleted")
	}
}
