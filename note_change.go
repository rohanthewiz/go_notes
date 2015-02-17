package main

import (
	"time"
	"errors"
)

// Record note changes so we can replay them on synch
type NoteChange struct {
Id          int64
Guid		string `sql: "size:40"` //Guid of the change
NoteGuid		string `sql: "size:40"` // Guid of the note
Operation	int32  // 1: Create, 2: Update, 3: Delete
Note Note
NoteId int64
NoteFragment NoteFragment
NoteFragmentId int64
CreatedAt   time.Time // A note change is never altered once created
}

const op_create int32 = 1
const op_update int32 = 2
const op_delete int32 = 3

type NoteFragment struct {
	Id          int64
	Bitmask		int16	// Indicate Active fields (allows for empty string update)
	// 0x8 - Title, 0x4 - Description, 0x2 - Body, 0x1 - Tag
	Title       string `sql: "size:128"`
	Description string `sql: "size:255"`
	Body        string `sql: "type:text"`
	Tag         string `sql: "size:128"`
}

type byCreatedAt []NoteChange

func (ncs byCreatedAt) Len() int {
	return len(ncs)
}
func (ncs byCreatedAt) Less(i int, j int) bool {
	return ncs[i].CreatedAt.Before(ncs[j].CreatedAt)
}
func (ncs byCreatedAt) Swap(i int, j int) {
	ncs[i], ncs[j] = ncs[j], ncs[i]
}


func retrieveNoteChangeByObject(nc NoteChange) (NoteChange, error) {
	var noteChanges []NoteChange
	db.Where("guid = ?", nc.Guid).Limit(1).Find(&noteChanges)
	if len(noteChanges) == 1 {
		return noteChanges[0], nil
	} else {
		return NoteChange{}, errors.New("NoteChange not found")
	}
}

func retrieveChangedNote(nc NoteChange) (Note, error) {
	var note Note
	db.Model(&nc).Related(&note)
	if note.Id > 0 {
		return note, nil
	} else {
		return Note{}, errors.New("Note not found")
	}
}

func retrieveNoteFrag(nc NoteChange) (NoteFragment, error) {
	var noteFrag NoteFragment
	db.Model(&nc).Related(&noteFrag)
	if noteFrag.Id > 0 {
		return noteFrag, nil
	} else {
		return NoteFragment{}, errors.New("NoteFragment not found")
	}
}

func printNoteChange(nc NoteChange) {
	pf("NoteChange: {Id: %d, Guid: %s, NoteGuid: %s, Oper: %d\nNote: {Id: %d, Guid: %s, Title: %s}\nNoteFragment: {Id: %d, Bitmask: %d, Title: %s, Description: %s, Body: %s, Tag: %s}}\n",
		nc.Id, short_sha(nc.Guid), short_sha(nc.NoteGuid), nc.Operation, nc.NoteId, short_sha(nc.Note.Guid), nc.Note.Title,
		nc.NoteFragment.Id, nc.NoteFragment.Bitmask, nc.NoteFragment.Title, nc.NoteFragment.Description, nc.NoteFragment.Body, nc.NoteFragment.Tag,
	)
}
