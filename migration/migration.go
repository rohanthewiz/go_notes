package migration

import (
	"fmt"
	"go_notes/authen/session"
	"go_notes/dbhandle"
	"go_notes/localsig"
	"go_notes/note"
	"go_notes/peer"
	"go_notes/user"
	"go_notes/utils"
)

func MigrateIfNeeded() {
	// Do we need to migrate?
	if !dbhandle.DB.HasTable(&peer.Peer{}) || !dbhandle.DB.HasTable(&note.Note{}) ||
		!dbhandle.DB.HasTable(&note.NoteChange{}) ||
		!dbhandle.DB.HasTable(&note.NoteFragment{}) ||
		!dbhandle.DB.HasTable(&localsig.LocalSig{}) ||
		!dbhandle.DB.HasTable(&user.User{}) ||
		!dbhandle.DB.HasTable(&session.Session{}) {
		Migrate()
	}
}

func Migrate() {
	// Create or update the table structure as needed
	fmt.Println("Migrating the DB...")
	// If the table is not existing, AutoMigrate will create the table automatically.
	// Fyi, AutoMigrate will only *add new columns*, it won't update column's type or delete unused columns
	// According to GORM: Feel free to change your struct, AutoMigrate will keep your database up-to-date.
	dbhandle.DB.AutoMigrate(
		&note.Note{}, &note.NoteChange{}, &note.NoteFragment{},
		&localsig.LocalSig{}, &user.User{}, &peer.Peer{}, &session.Session{},
	)

	dbhandle.DB.Model(&user.User{}).AddUniqueIndex("idx_user_guid", "guid")
	dbhandle.DB.Model(&user.User{}).AddUniqueIndex("idx_user_email", "email")
	dbhandle.DB.Model(&session.Session{}).AddUniqueIndex("idx_session_key", "session_key")
	dbhandle.DB.Model(&note.Note{}).AddUniqueIndex("idx_note_guid", "guid")
	dbhandle.DB.Model(&note.Note{}).AddUniqueIndex("idx_note_title", "title")
	dbhandle.DB.Model(&note.Note{}).AddIndex("idx_note_user", "user")
	dbhandle.DB.Model(&note.NoteChange{}).AddUniqueIndex("idx_note_change_guid", "guid")
	dbhandle.DB.Model(&note.NoteChange{}).AddIndex("idx_note_change_note_guid", "note_guid")
	dbhandle.DB.Model(&note.NoteChange{}).AddIndex("idx_note_change_created_at", "created_at")

	EnsureDBSig() // Initialize local with a SHA1 signature if it doesn't already have one
	fmt.Println("Migration complete")
}

func EnsureDBSig() {
	var localSigs []localsig.LocalSig
	dbhandle.DB.Find(&localSigs)

	if len(localSigs) == 1 && len(localSigs[0].Guid) == 40 &&
		len(localSigs[0].ServerSecret) == 40 {
		return
	} // all is good

	if len(localSigs) == 0 { // create the signature
		dbhandle.DB.Create(&localsig.LocalSig{Guid: utils.GenerateSHA1(), ServerSecret: utils.GenerateSHA1()})
		if dbhandle.DB.Find(&localSigs); len(localSigs) == 1 && len(localSigs[0].Guid) == 40 { // was it saved?
			fmt.Println("Local signature created")
		}
	} else {
		panic("Error in the 'local_sigs' table. There should be only one and only one good row")
	}
}
