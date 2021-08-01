package session

import (
	"go_notes/user"
	"time"
)

type Session struct {
	Id             int64
	SessionKey     string `sql:"type:text"`
	User           user.User
	LoginTime      time.Time
	LastActiveTime time.Time
}
