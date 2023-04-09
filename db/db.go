package db

import (
	"distributed-chat/minion/structs"
	"fmt"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Message = structs.Message
type User = structs.User
type Minion = structs.Minion

func InitDb() gorm.DB {
	db, err := gorm.Open(sqlite.Open("production.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	return *db
}

func CreateDbFromSchema(db gorm.DB) {
	// Migrate the schema
	db.AutoMigrate(&Message{})
	db.AutoMigrate(&User{})
	db.AutoMigrate(&Minion{})
}

func CreateUser(db gorm.DB, user User) (User, error) {
	result := db.Create(&user)
	if result.Error != nil {
		fmt.Println("Error while creating user")
		return user, result.Error
	}
	return user, nil
}

func RetrieveUserByName(db gorm.DB, username string) (User, error) {
	var user User
	err := db.First(&user, "username = ?", username).Error
	return user, err
}

func RetrieveUsersIChatWith(db gorm.DB, username string) []User {
	var users []User
	subQuery1 := db.Table("messages").Select("receiver_name as username").Where("sender_name = ?", username)
	subQuery2 := db.Table("messages").Select("sender_name as username").Where("receiver_name = ?", username)
	subQuery3 := db.Raw("? UNION ?", subQuery1, subQuery2)
	db.Where("id IN (?)", subQuery3).Find(&users)
	return users
}

func CreateMessage(db gorm.DB, message Message) (Message, error) {
	result := db.Create(&message)
	if result.Error != nil {
		fmt.Println("Error while adding message")
	}
	return message, result.Error
}

func RetrieveLatestMessageBySenderAndReceiver(db gorm.DB, senderName string, receiverName string) (Message, error) {
	var message Message
	result := db.Where("sender_name = ? AND receiver_name = ?", senderName, receiverName).Last(&message)
	return message, result.Error
}

func RetrieveAllMessagesBySenderAndReceiver(db gorm.DB, senderName string, receiverName string) []Message {
	var messages []Message
	db.Where("sender_name = ? AND receiver_name = ?", senderName, receiverName).Find(&messages)
	return messages
}

func RetrieveAllMessagesBetweenUsers(db gorm.DB, username1 string, username2 string) []Message {
	var messages []Message
	messages = RetrieveAllMessagesBySenderAndReceiver(db, username1, username2)
	messages = append(messages, RetrieveAllMessagesBySenderAndReceiver(db, username2, username1)...)
	return messages
}

func RetrieveAllUsers(db gorm.DB) []User {
	var users []User
	result := db.Find(&users)
	if result.Error != nil {
		fmt.Println("Error while retrieving users")
	}
	return users
}

func RetrieveSelf(db gorm.DB, minionName string) (Minion, error) {
	var minion Minion
	err := db.First(&minion, "minion_name = ?", minionName).Error
	return minion, err
}

func CreateMinion(db gorm.DB, minion Minion) (Minion, error) {
	result := db.Create(&minion)
	if result.Error != nil {
		fmt.Println("Error while creating user")
		return minion, result.Error
	}
	return minion, nil
}

func DeleteUser(db gorm.DB, user User) {
	db.Delete(&user)
}

func UpdateUser(db gorm.DB, user User) User {
	db.Save(&user)
	return user
}

func UpdateMinion(db gorm.DB, minion Minion) Minion {
	db.Save(&minion)
	return minion
}
