package note

import (
	"errors"
	"go_notes/config"
	"go_notes/dbhandle"
	"strings"
)

func GetNote(guid string) (Note, error) {
	var n Note
	dbhandle.DB.Where("guid = ?", guid).First(&n)
	if n.Id != 0 {
		return n, nil
	} else {
		return n, errors.New("Note not found")
	}
}

func FindNoteByTitle(title string) (bool, Note) {
	var notes []Note
	dbhandle.DB.Where("title = ?", title).Limit(1).Find(&notes)
	if len(notes) == 1 {
		return true, notes[0]
	} else {
		return false, Note{} // yes this is the way you represent an empty Note object/struct
	}
}

func FindNoteById(id int64) Note {
	var n Note
	dbhandle.DB.First(&n, id)
	return n
}

// Query by Id, return all notes, query all fields for one param, query a combination of fields and params

// Query by Tag words not just phrases
func filterByTag(notes []Note, tag string) (out []Note) {
	if tag == "" {
		return notes
	}

	out = make([]Note, 0, len(notes)) // it will likely be less than the input, but saves on allocs

	for _, note := range notes {
		tags := strings.Split(note.Tag, ",")
		for _, t := range tags {
			if strings.TrimSpace(t) == tag {
				out = append(out, note)
				break
			}
		}
	}
	return out
}

func QueryNotes(nf *NotesFilter) (notes []Note) {
	// The order of the `if` here is very important - esp. for the webserver!
	if nf.Last {
		dbhandle.DB.Order("updated_at desc").Limit(1).Find(&notes)

	} else if nf.Id > 0 {
		dbhandle.DB.Find(&notes, nf.Id)

		// TAG and wildcard. For tags, we match words and possible pharses with like.
		// Later we will filter on just word matches
	} else if len(nf.Tags) > 0 && nf.QueryStr != "" { // TODO - we'll decide whether to do fuzzy or by word
		dbhandle.DB.Where("tag LIKE ? AND (title LIKE ? OR description LIKE ? OR body LIKE ?)",
			"%"+nf.Tags[0]+"%",
			"%"+nf.QueryStr+"%",
			"%"+nf.QueryStr+"%",
			"%"+nf.QueryStr+"%",
		).Order("updated_at desc").Limit(nf.Limit).Find(&notes)
		notes = filterByTag(notes, nf.Tags[0]) // len must be checked above

		// TITLE and wildcard
	} else if nf.Title != "" && nf.QueryStr != "" {
		dbhandle.DB.Where("title LIKE ? AND (tag LIKE ? OR description LIKE ? OR body LIKE ?)",
			"%"+nf.Title+"%",
			"%"+nf.QueryStr+"%",
			"%"+nf.QueryStr+"%",
			"%"+nf.QueryStr+"%",
		).Order("updated_at desc").Limit(nf.Limit).Find(&notes)
		//
	} else if nf.QueryStr == "all" {
		dbhandle.DB.Order("updated_at desc").Limit(nf.Limit).Find(&notes)
		// General query
	} else if nf.QueryStr != "" {
		dbhandle.DB.Where("tag LIKE ? OR title LIKE ? OR description LIKE ? OR body LIKE ?",
			"%"+nf.QueryStr+"%",
			"%"+nf.QueryStr+"%",
			"%"+nf.QueryStr+"%",
			"%"+nf.QueryStr+"%",
		).Order("updated_at desc").Limit(nf.Limit).Find(&notes)
	}

	// For now on remote webserver  filter out private notes
	if config.Opts.IsRemoteSvr {
		notes = FilterOutPrivate(notes)
	}

	return notes
}
