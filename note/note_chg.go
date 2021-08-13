package note

import (
	"encoding/json"
	"errors"
	"fmt"
	"go_notes/dbhandle"
	"time"
)

// Record note changes so we can replay them on synch
type NoteChange struct {
	Id             int64
	Guid           string `sql:"size:40"` // Guid of the change // TODO - make this PG text
	NoteGuid       string `sql:"size:40"` // Guid of the note
	Operation      int32  // 1: Create, 2: Update, 3: Delete, 9: Synch
	Note           Note
	NoteId         int64
	User           string // (GUID)  //todo - Add Index //A notechange always belongs to a single user
	NoteFragment   NoteFragment
	NoteFragmentId int64
	CreatedAt      time.Time // A note change is never altered once created
}

const OpCreate int32 = 1
const OpUpdate int32 = 2
const OpDelete int32 = 3

// These are changes for a note
type NoteFragment struct {
	Id      int64
	Bitmask int16 // Indicate Active fields (allows for empty string update)
	// 0x8 - Title, 0x4 - Description, 0x2 - Body, 0x1 - Tag
	Title       string `sql:"size:128"`
	Description string `sql:"size:255"`
	Body        string `sql:"type:text"`
	Tag         string `sql:"size:128"`
}

type ByCreatedAt []NoteChange

func (ncs ByCreatedAt) Len() int {
	return len(ncs)
}
func (ncs ByCreatedAt) Less(i int, j int) bool {
	return ncs[i].CreatedAt.Before(ncs[j].CreatedAt)
}
func (ncs ByCreatedAt) Swap(i int, j int) {
	ncs[i], ncs[j] = ncs[j], ncs[i]
}

func (nc *NoteChange) Retrieve() (NoteChange, error) {
	var noteChanges []NoteChange
	dbhandle.DB.Where("guid = ?", nc.Guid).Limit(1).Find(&noteChanges)
	if len(noteChanges) == 1 {
		return noteChanges[0], nil
	} else {
		return NoteChange{}, errors.New("NoteChange not found")
	}
}

func (nc *NoteChange) RetrieveNote() (Note, error) {
	var nte Note
	dbhandle.DB.Model(nc).Related(&nte)
	if nte.Id > 0 {
		return nte, nil
	} else {
		return Note{}, errors.New("Note not found")
	}
}

func (nc *NoteChange) RetrieveNoteFrag() (NoteFragment, error) {
	var noteFrag NoteFragment
	dbhandle.DB.Model(nc).Related(&noteFrag)
	if noteFrag.Id > 0 {
		return noteFrag, nil
	} else {
		return NoteFragment{}, errors.New("NoteFragment not found")
	}
}

func (nc *NoteChange) Print() {
	j_str, err := json.Marshal(*nc)
	if err != nil {
		fmt.Println(string(j_str))
	} else {
		fmt.Printf("%+v\n", nc)
	}
	//	pf("NoteChange: {Id: %d, Guid: %s, NoteGuid: %s, Oper: %d\nNote: {Id: %d, Guid: %s, Title: %s}\nNoteFragment: {Id: %d, Bitmask: %d, Title: %s, Description: %s, Body: %s, Tag: %s}}\n",
	//		nc.Id, shortSHA(nc.Guid), shortSHA(nc.NoteGuid), nc.Operation, nc.NoteId, shortSHA(nc.Note.Guid), nc.Note.Title,
	//		nc.NoteFragment.Id, nc.NoteFragment.Bitmask, nc.NoteFragment.Title, nc.NoteFragment.Description, nc.NoteFragment.Body, nc.NoteFragment.Tag,
	//	)
}
