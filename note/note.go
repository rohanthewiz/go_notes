package note

import "time"

type Note struct {
	Id          uint64
	Guid        string `sql:"size:40"` // todo make this PG text
	Title       string `sql:"size:128"`
	Description string `sql:"size:255"`
	Body        string `sql:"type:text"`
	Tag         string `sql:"size:128"`
	User        string // who's account is this currently in (GUID)
	Creator     string // (GUID) who originally created the note
	SharedBy    string // (GUID) if it was shared to me, by who?
	Public      bool   // Was it made public for all users
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
