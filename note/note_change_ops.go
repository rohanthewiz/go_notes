package note

import (
	"fmt"
	"go_notes/dbhandle"
	"go_notes/utils"
)

func RetrieveLastChangeForNote(note_guid string) NoteChange {
	var noteChange NoteChange
	dbhandle.DB.Where("note_guid = ?", note_guid).Order("created_at desc").Limit(1).Find(&noteChange)
	return noteChange
}

func RetrieveLatestChange() NoteChange {
	var noteChange NoteChange
	dbhandle.DB.Order("created_at desc").First(&noteChange)
	return noteChange
}

// Save the change object which will create a Note on CreateOp or a NoteFragment on UpdateOp
func SaveNoteChange(nc NoteChange) bool {
	utils.Pf("Saving change object...%s\n", utils.ShortSHA(nc.Guid))
	// Make sure all ids are zeroed - A non-zero Id will not be created
	nc.Id = 0

	dbhandle.DB.Create(&nc)         // will auto create contained objects too and it's smart - 'nil' children will not be created :-)
	if !dbhandle.DB.NewRecord(nc) { // was it saved?
		utils.Pl("Note change saved:", utils.ShortSHA(nc.Guid), ", Operation:", nc.Operation)
		return true
	}
	fmt.Println("Failed to record note changes.", nc.Note.Title, "Changed note Guid:",
		utils.ShortSHA(nc.NoteGuid), "NoteChange Guid:", utils.ShortSHA(nc.Guid))
	return false
}

// Get all local NCs later than the synchPoint
func RetrieveLocalNoteChangesFromSynchPoint(synch_guid string) []NoteChange {
	var noteChange NoteChange
	var noteChanges []NoteChange

	dbhandle.DB.Where("guid = ?", synch_guid).First(&noteChange) // There should be only one
	utils.Pf("Synch point note change is: %v\n", noteChange)
	if noteChange.Id < 1 {
		utils.Pl("Can't find synch point locally - retrieving all note_changes", utils.ShortSHA(synch_guid))
		dbhandle.DB.Find(&noteChanges).Order("created_at, asc")
	} else {
		utils.Pl("Attempting to retrieve note_changes beyond synch_point")
		dbhandle.DB.Find(&noteChanges, "created_at > ?", noteChange.CreatedAt).Order("created_at asc")
	}
	return noteChanges
}

func VerifyNoteChangeApplied(nc NoteChange) {
	utils.Pl("----------------------------------")
	retrievedChange, err := nc.Retrieve()
	if err != nil {
		fmt.Println("Error retrieving the note change")
	} else if nc.Operation == 1 {
		retrievedNote, err := retrievedChange.RetrieveNote()
		utils.Pf("retrievedNote: %s\n", retrievedNote)
		if err != nil {
			fmt.Println("Error retrieving the note changed")
		} else {
			utils.Pf("Note created:\n%v\n", utils.TruncString(retrievedNote.Guid, 12), retrievedNote.Title)
		}
	} else if nc.Operation == 2 {
		retrievedFrag, err := retrievedChange.RetrieveNoteFrag()
		if err != nil {
			fmt.Println("Error retrieving the note fragment")
		} else {
			utils.Pf("Note Fragment created:\n%v\n", retrievedFrag.Id, retrievedFrag.Bitmask, "-", retrievedFrag.Title)
		}
	}
}
