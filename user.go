package main
import "time"


// A User can have multiple peers
// Locally we will have only one user
// The server will manage users and populate notes with the user's GUID. On initial migration, the user will be asked to setup the user
// with prompts at the command line
// The user's GUID will be a hash of their email
type User struct {
	Id uint64
	FirstName string
	LastName string
	Email string // will be the users unique identifier
	Guid string  // GUID will be hash of users email //todo - Add Index
	CryptedPassword string
	Seed string
	Peers []Peer // has many peers
	CreatedAt time.Time
	UpdatedAt time.Time
}
