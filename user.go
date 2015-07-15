package main
import (
	"time"
	"errors"
	"github.com/elithrar/simple-scrypt"
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
	if password != password_conf {
	  	return nil, errors.New("Password and password confirmation do not match")
	}
	user := new(User)
	user.FirstName = first_name
	user.LastName = last_name
	user.Email = email
	user.Guid = random_sha1()
	user.Salt = random_sha1()
	hash, err := scrypt.GenerateFromPassword([]byte(password + user.Salt), scrypt.DefaultParams)
	if err != nil {
		lpl("Failed to generate a password hash for", user.Email)
		return nil, err
	}
	user.CryptedPassword = string(hash)
	user.Create()
	return user, nil
}

func (u * User) Create() {
	db.Create(u)
}

func (u *User) Auth( word string) bool {
	err := scrypt.CompareHashAndPassword([]byte(u.CryptedPassword), []byte(word + u.Salt))
	if err != nil {
		return false
	}
	return true
}
