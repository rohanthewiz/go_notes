package dbhandle

import (
	"go_notes/config"

	"github.com/jinzhu/gorm"
)

var DB *gorm.DB

func InitDB() (err error) {
	DB, err = gorm.Open("sqlite3", config.Opts.DBPath)
	return err
}

// db.LogMode(optsIntf["debug"].(bool)) // Set debug mode for Gorm db
