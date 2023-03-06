package db

import (
	"distributed-chat/minion/structs"
	"fmt"

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

func RetrieveLatestMessageBySenderAndReceiver(db gorm.DB, sender_id uint, receiver_id uint) Message {
	var message Message
	db.Where("sender_id = ? AND receiver_id = ?", sender_id, receiver_id).Last(&message)
	return message
}

func RetrieveAllMessagesBySenderAndReceiver(db gorm.DB, sender_id uint, receiver_id uint) []Message {
	var messages []Message
	db.Where("sender_id = ? AND receiver_id = ?", sender_id, receiver_id).Find(&messages)
	return messages
}

func RetrieveAllMessagesBetweenUsers(db gorm.DB, user_id1 uint, user_id2 uint) []Message {
	var messages []Message
	messages = RetrieveAllMessagesBySenderAndReceiver(db, user_id1, user_id2)
	messages = append(messages, RetrieveAllMessagesBySenderAndReceiver(db, user_id2, user_id1)...)
	return messages
}

func RetrieveUsersIChatWith(db gorm.DB, user_id uint) []User {
	var users []User
	subQuery1 := db.Table("messages").Select("receiver_id as user_id").Where("sender_id = ?", user_id)
	subQuery2 := db.Table("messages").Select("sender_id as user_id").Where("receiver_id = ?", user_id)
	subQuery3 := db.Raw("? UNION ?", subQuery1, subQuery2)
	db.Where("id IN (?)", subQuery3).Find(&users)
	return users
}

func CreateUser(db gorm.DB, user User) User {
	result := db.Create(&user)
	if result.Error != nil {
		fmt.Println("Error while creating user")
	}
	return user
}

func CreateMessage(db gorm.DB, message Message) Message {
	result := db.Create(&message)
	if result.Error != nil {
		fmt.Println("Error while adding message")
	}
	return message
}
