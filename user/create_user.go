package user

import (
	"errors"
	"fmt"
	db "go_notes/dbhandle"
	"go_notes/utils"
	"time"
)

func CreateUser(usr User, password string) (err error) {
	const errStage = " when creating user"

	if usr.Email == "" {
		err = errors.New("email is required when creating user")
		fmt.Println(err)
		return
	}

	if password == "" {
		err = errors.New("password is required when creating user")
		fmt.Println(err)
		return
	}

	if usr.Username == "" { // this should never happen, but just in case
		usr.Username = usr.Email
	}

	var users []User
	db.DB.Where("email = ?", usr.Email).Find(&users)

	if len(users) > 0 {
		err = errors.New("user with email already exists")
		fmt.Println(err.Error())
		return
	}

	gSalt, err := utils.RandomTokenBase64(64)
	if err != nil {
		fmt.Println(err.Error() + errStage)
		return err
	}
	usr.Guid = utils.Blake256(gSalt + usr.Email)

	pSalt, err := utils.RandomTokenBase64(64)
	if err != nil {
		fmt.Println(err.Error() + errStage)
		return err
	}
	usr.Salt = pSalt

	pHash := utils.Blake384(pSalt + password)
	usr.HashedPassword = pHash

	usr.CreatedAt = time.Now()

	fmt.Printf("usr %#v\n", usr)
	db.DB.Create(&usr)
	return db.DB.Error
}
