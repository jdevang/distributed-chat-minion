package db

import (
	"distributed-chat/minion/structs"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Message = structs.Message
type Gru = structs.Gru
type User = structs.User

func InitDb() gorm.DB {
	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{}) // change to postgres after setting up dockerise
	if err != nil {
		panic("failed to connect database")
	}
	return *db
}

func CreateDbFromSchema(db gorm.DB) {
	// Migrate the schema
	db.AutoMigrate(&Message{})
	db.AutoMigrate(&Gru{})
	db.AutoMigrate(&User{})
}

func RetrieveMessageById(db gorm.DB, message_id uint) Message {
	var message Message
	db.First(&message, message_id)
	return message
}

func RetrieveUserById(db gorm.DB, user_id uint) User {
	var user User
	db.First(&user, user_id)
	return user
}
