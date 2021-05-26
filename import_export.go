package main

import (
	"encoding/csv"
	"encoding/gob"
	"fmt"
	"go_notes/note"
	"log"
	"os"
)

func exportCsv(notes []note.Note, out_file string) {
	csv_file, err := os.Create(out_file)
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}
	defer func(csv_file *os.File) {
		err := csv_file.Close()
		if err != nil {
			log.Println("Error in csv file close:", err)
		}
	}(csv_file)
	writer := csv.NewWriter(csv_file)

	for _, n := range notes {
		arr := []string{n.Title, n.Description, n.Body, n.Tag}
		err := writer.Write(arr)
		if err != nil {
			fmt.Println("Error: ", err)
			return
		}
	}
	writer.Flush()
	fmt.Println("Exported to", out_file)
}

func exportGob(notes []note.Note, out_file string) {
	gob_file, err := os.Create(out_file)
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}
	defer func(gob_file *os.File) {
		err := gob_file.Close()
		if err != nil {
			log.Println("Error in GOB file close:", err)
		}
	}(gob_file)
	gob_encoder := gob.NewEncoder(gob_file)

	err = gob_encoder.Encode(notes)
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}
	fmt.Println("Exported to", out_file)
}

func importGob(in_file string) {
	var notes []note.Note

	gob_file, err := os.Open(in_file)
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}
	defer func(gob_file *os.File) {
		err := gob_file.Close()
		if err != nil {
			log.Println("Error in GOB file close:", err)
		}
	}(gob_file)

	decoder := gob.NewDecoder(gob_file)
	err = decoder.Decode(&notes)
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}
	listNotes(notes, false)
	fmt.Printf("%d note(s) retrieved from %s\n", len(notes), in_file)

	// Update, create or discard?
	for _, n := range notes {
		exists, nte := findNoteByTitle(n.Title)
		if exists {
			fmt.Println("This note already exists: ", nte.Title)
			if n.UpdatedAt.After(nte.UpdatedAt) {
				pl("The imported note is newer, updating...")
				nte.Description = n.Description
				nte.Body = n.Body
				nte.Tag = n.Tag
				db.Save(&nte)
				listNotes([]note.Note{nte}, false) // [:] means all of the slice
			} else {
				fmt.Println("The imported note is not newer, ignoring...")
			}
		} else {
			DoCreate(note.Note{Guid: generateSHA1(), Title: n.Title, Description: n.Description, Body: n.Body, Tag: n.Tag})
			fmt.Printf("Created -->Guid: %s, Title: %s - Desc: %s\nBody: %s\nTags: %s\n", shortSHA(n.Guid), n.Title, n.Description, n.Body, n.Tag)
		}
	}
}

func importCsv(in_file string) {
	csv_file, err := os.Open(in_file)
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}
	defer func(csv_file *os.File) {
		err := csv_file.Close()
		if err != nil {
			log.Println("Error in CSV file close:", err)
		}
	}(csv_file)

	reader := csv.NewReader(csv_file)
	reader.FieldsPerRecord = -1 // Todo match this with Note struct

	rawCSVdata, err := reader.ReadAll()
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}

	// sanity check, display to standard output
	for _, f := range rawCSVdata {
		exists, nte := findNoteByTitle(f[0])
		if exists {
			fmt.Println("This note already exists: ", nte.Title)
			// we could check an 'update on import' option here, set the corresponding fields, then save
			// or we could decide to update based on last_updated, but the export would have to save updated times - this would be a gob format
		} else {
			DoCreate(note.Note{Guid: generateSHA1(), Title: f[0], Description: f[1], Body: f[2], Tag: f[3]})
			fmt.Printf("Created -->Title: %s - Desc: %s\nBody: %s\nTags: %s\n", f[0], f[1], f[2], f[3])
		}
	}
}
