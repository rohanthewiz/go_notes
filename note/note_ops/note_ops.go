package note_ops

import (
	"fmt"
	"go_notes/dbhandle"
	"go_notes/note"
	"go_notes/utils"
)

// Create, Update or Delete a note, while tracking the change
func PerformNoteChange(nc note.NoteChange) bool {
	nc.Print()
	// Get The latest change for the current note in the local changeset
	lastNC := note.RetrieveLastChangeForNote(nc.NoteGuid)

	switch nc.Operation {
	case note.OpCreate:
		if lastNC.Id > 0 {
			fmt.Println("Note - Title", lastNC.Note.Title, "Guid:", utils.ShortSHA(lastNC.NoteGuid), "already exists locally - cannot create")
			return false
		}
		nc.Note.Id = 0 // Make sure the embedded note object has a zero id for creation
	case note.OpUpdate:
		nte, err := note.GetNote(nc.NoteGuid)
		if err != nil {
			fmt.Println("Cannot update a non-existent note:", utils.ShortSHA(nc.NoteGuid))
			return false
		}
		UpdateNote(nte, nc)
		nc.NoteFragment.Id = 0 // Make sure the embedded note_fragment has a zero id for creation
	case note.OpDelete:
		if lastNC.Id < 1 {
			fmt.Printf("Cannot delete a non-existent note (Guid:%s)", utils.ShortSHA(nc.NoteGuid))
			return false
		} else {
			dbhandle.DB.Where("guid = ?", lastNC.NoteGuid).Delete(note.Note{})
		}
	default:
		return false
	}
	return note.SaveNoteChange(nc)
}

func UpdateNote(n note.Note, nc note.NoteChange) {
	// Update bitmask allowed fields - this allows us to set a field to ""  // Updates are stored as note fragments
	if nc.NoteFragment.Bitmask&0x8 == 8 {
		n.Title = nc.NoteFragment.Title
	}
	if nc.NoteFragment.Bitmask&0x4 == 4 {
		n.Description = nc.NoteFragment.Description
	}
	if nc.NoteFragment.Bitmask&0x2 == 2 {
		n.Body = nc.NoteFragment.Body
	}
	if nc.NoteFragment.Bitmask&0x1 == 1 {
		n.Tag = nc.NoteFragment.Tag
	}
	dbhandle.DB.Save(&n)
}
