package main
import "time"


// A User can have multiple peers
// Locally we will have only one user
// The server will manage users and populate notes with the user's GUID.
// A user must be setup beforehand on the server - preferably through registration
// The user's GUID will be provided by the server
type User struct {
	Id uint64
	FirstName string
	LastName string
	Email string // will be the users unique identifier
	Guid string  // GUID will be hash of users email
	CryptedPassword string
	Seed string
	Peers []Peer // has many peers
	NoteChanges []NoteChange // for synching
	Notes []Note // for notes retrieval
	CreatedAt time.Time
	UpdatedAt time.Time
}
