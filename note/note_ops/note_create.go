package note_ops

import (
	"fmt"
	"go_notes/dbhandle"
	"go_notes/note"
	"go_notes/utils"
)

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
	PerformNoteChange(
		note.NoteChange{
			Guid: utils.GenerateSHA1(), Operation: 1,
			NoteGuid:     nte.Guid,
			Note:         nte,
			NoteFragment: note.NoteFragment{},
		})

	if n, err := note.GetNote(nte.Guid); err != nil {
		utils.Pf("Error creating note %v\n", nte)
		return 0
	} else {
		utils.Pf("Record saved: [%d] %s\n", n.Id, n.Title)
		id = n.Id
	}
	return id
}
