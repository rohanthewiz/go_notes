package main
import (
	"time"
	"errors"
)


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
	Salt string
	Peers []Peer // has many peers
	NoteChanges []NoteChange // for synching
	Notes []Note // for notes retrieval
	CreatedAt time.Time
	UpdatedAt time.Time
}
func NewUser(first_name string, last_name string, email string, password string, password_conf string) (* User, error) {
	var user User
	if password != password_conf {
	  	return user, errors.New("Password and password confirmation does not match")
	}
	user = User{FirstName: first_name, LastName: last_name, Email: email}
	user.Guid = generate_sha1()
	user.Salt = generate_sha1()
	user.CryptedPassword = hashPassword(password, user.Salt)
	return user, nil
}

func (u * User) create() {
	db.Create(u)
}