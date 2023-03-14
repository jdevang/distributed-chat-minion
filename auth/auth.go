package auth

import (
	"distributed-chat/minion/db"
	"distributed-chat/minion/structs"
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User = structs.User
type Message = structs.Message

func LoginUser(connectedDb gorm.DB, username string, password string) bool {
	passwordHash, err := hashPassword(password)
	if err != nil {
		fmt.Println("Error in generating passwrod hash")
		return false
	}
	user, err := db.RetrieveUserByName(connectedDb, username)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		fmt.Println("Error: No such user")
		return false
	}
	return checkPasswordHash(user, passwordHash)
}

func RegisterUser(connectedDb gorm.DB, username string, password string) User {
	passwordHash, err := hashPassword(password)
	if err != nil {
		fmt.Println("Error in generating password hash")
		return User{}
	}
	user := db.CreateUser(connectedDb, User{Name: username, Password: passwordHash})
	return user
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func checkPasswordHash(user User, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(user.Password))
	return err == nil
}
