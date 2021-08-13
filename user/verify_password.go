package user

import (
	"go_notes/dbhandle"
	"go_notes/utils"
)

// TODO route and form
// Session cookie
func VerifyPassword(pWord string, usrGuid string) (pass bool) {
	usr := User{}
	dbhandle.DB.Where("guid = ?", usrGuid).Limit(1).Find(&usr)

	return utils.Blake384(usr.Salt+pWord) == usr.HashedPassword
}
