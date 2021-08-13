package note_ops

import (
	"fmt"
	"go_notes/config"
	"go_notes/note"
	"strings"
)

func CmdlineQueryAndAction() {
	o := config.Opts // alias
	notes := note.QueryNotes(note.NotesFilterFromOpts())

	fmt.Println("") // for UI sake
	note.ListNotes(notes, true)

	// Actions that can follow a query
	if o.Export != "" {
		arr := strings.Split(o.Export, ".")
		last := len(arr) - 1

		if arr[last] == "csv" {
			ExportCsv(notes, o.Export)
		}
		if arr[last] == "gob" {
			ExportGob(notes, o.Export)
		}
	} else if o.Update { // update
		note.UpdateNotes(notes)

		// See if we want to delete
	} else if o.Delete {
		note.DeleteNotes(notes)
	}
}
