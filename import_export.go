package main
import (
	"fmt"
	"encoding/csv"
	"encoding/gob"
	"os"
)

func exportCsv(notes []Note, out_file string) {
	csv_file, err := os.Create(out_file)
	if err != nil { fmt.Println("Error: ", err); return }
	defer csv_file.Close()
	writer := csv.NewWriter(csv_file)

	for _, n := range notes {
		arr := []string{n.Title, n.Description, n.Body, n.Tag}
		err := writer.Write(arr)
		if err != nil { fmt.Println("Error: ", err); return }
	}
	writer.Flush()
	pl("Exported to", out_file)
}

func exportGob(notes []Note, out_file string) {
	gob_file, err := os.Create(out_file)
	if err != nil { fmt.Println("Error: ", err); return }
	defer gob_file.Close()
	gob_encoder := gob.NewEncoder(gob_file)

	err = gob_encoder.Encode(notes)
	if err != nil { fmt.Println("Error: ", err); return }
	pl("Exported to", out_file)
}

func importGob(in_file string) {
	var notes []Note

	gob_file, err := os.Open(in_file)
	if err != nil { fmt.Println("Error: ", err); return }
	defer gob_file.Close()

	decoder := gob.NewDecoder(gob_file)
	err = decoder.Decode(&notes)
	if err != nil { fmt.Println("Error: ", err); return }
	listNotes(notes, false)
	fmt.Printf("%d note(s) retrieved from %s\n", len(notes), in_file)

	// Update, create or discard?
	for _, n := range notes {
		exists, note := find_note_by_title(n.Title)
		if exists {
			pl("This note already exists: ", note.Title)
			if n.UpdatedAt.After(note.UpdatedAt) {
				pl("The imported note is newer, updating...")
				note.Description = n.Description
				note.Body = n.Body
				note.Tag = n.Tag
				db.Save(&note)
				listNotes([]Note{note}, false) // [:] means all of the slice
			}	else { pl("The imported note is not newer, ignoring...")}
		} else {
			do_create( Note{ Guid: generate_sha1(), Title: n.Title, Description: n.Description, Body: n.Body, Tag: n.Tag } )
			fmt.Printf("Created -->Guid: %s, Title: %s - Desc: %s\nBody: %s\nTags: %s\n", short_sha(n.Guid), n.Title, n.Description, n.Body, n.Tag)
		}
	}
}

func importCsv(in_file string) {
	csv_file, err := os.Open(in_file)
	if err != nil { fmt.Println("Error: ", err); return }
	defer csv_file.Close()

	reader := csv.NewReader(csv_file)
	reader.FieldsPerRecord = -1 // Todo match this with Note struct

	rawCSVdata, err := reader.ReadAll()
	if err != nil { fmt.Println("Error: ", err); return }

	// sanity check, display to standard output
	for _, f := range rawCSVdata {
		exists, note := find_note_by_title(f[0])
		if exists {
			pl("This note already exists: ", note.Title)
			// we could check an 'update on import' option here, set the corresponding fields, then save
			// or we could decide to update based on last_updated, but the export would have to save updated times - this would be a gob format
		} else {
			do_create( Note{Guid: generate_sha1(), Title: f[0], Description: f[1], Body: f[2], Tag: f[3]} )
			fmt.Printf("Created -->Title: %s - Desc: %s\nBody: %s\nTags: %s\n", f[0], f[1], f[2], f[3])
		}
	}
}
