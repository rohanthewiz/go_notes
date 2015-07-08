package main
import (
	"testing"
	"github.com/jinzhu/gorm"
	"log"
)

func setup() gorm.DB {
	db, err := gorm.Open("sqlite3", "./test/test.sqlite")
	if err != nil {
		log.Fatal("Error opening the test db")
	}
	db.AutoMigrate(&User{})
	return db
}

func teardown() {
		db.Close()
}

func TestUserCreate( t *testing.T ) {
	lpl("Testing User create...")
	db = setup()

	user, err := NewUser("John", "Brown", "jb@one.com", "jb111", "jb111")
	if err != nil || user.FirstName != "John" || user.LastName != "Brown" {
		t.Error("User creation unsuccessful")
	}

	if db.NewRecord(user) {
		t.Error("User was not saved")
	}

	// Good password should authenticate
	if ! user.Auth("jb111") {
		t.Error("User cannot authenticate")
	}

	// Bad password should not auth
	if user.Auth("jbill") {
		t.Error("Bad password should not authenticate")
	}

	teardown()
	lpl("Done")
}
