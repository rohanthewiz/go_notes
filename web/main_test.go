package web

import (
	"fmt"
	"go_notes/config"
	db "go_notes/dbhandle"
	"go_notes/migration"
	"os"
	"strings"
	"testing"
)

func TestMain(m *testing.M) {
	println(strings.Repeat("-", 20), "Setting up tests", strings.Repeat("-", 20))
	const testDB = "/home/ro/goprojs/go_notes/test_db/test1.sqlite"

	_, _ = config.GetOpts()

	config.Opts.DBPath = testDB
	println("Overrided tests to use db:", testDB)

	err := db.InitDB()
	if err != nil { // Can't err chk db conn outside method, so do it here
		fmt.Println("There was an error connecting to the DB", err)
		fmt.Println("DBPath: " + config.Opts.DBPath)
		os.Exit(2)
	}

	migration.MigrateIfNeeded()

	println(strings.Repeat("-", 20), "Running tests", strings.Repeat("-", 20))
	ret := m.Run()
	println("Testing complete")
	os.Exit(ret)
}
