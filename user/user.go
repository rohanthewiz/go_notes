package user

import (
	"time"
)

// A User can have multiple peers
// Locally we will have only one user
// The server will manage users and populate notes with the user's GUID. On initial migration, the user will be asked to setup the user
// with prompts at the command line
// The user's GUID could be a hash of their email
type User struct {
	Id             int64
	FirstName      string
	LastName       string
	Email          string // will be the users unique identifier
	Username       string // this is the user's handle in the system // login will be done via email though
	Guid           string `sql:"type:text"` // GUID could be hash of users email //todo - Add Index
	HashedPassword string `sql:"type:text"`
	Salt           string `sql:"type:text"`
	IsDisabled     bool
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
