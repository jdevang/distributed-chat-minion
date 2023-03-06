package main

import (
	"distributed-chat/minion/structs"
)

type Message = structs.Message
type User = structs.User

func main() {
	// connectedDb := db.InitDb()
	// message := db.RetrieveMessageById(connectedDb, 2)
	// var users []User
	// db.CreateUser(connectedDb, User{Name: "Joe"})
	// db.CreateUser(connectedDb, User{Name: "Jane"})
	// connectedDb.Find(&users)
	// db.CreateMessage(connectedDb, Message{SenderID: 1, ReceiverID: 2, Content: "Hi"})
	// db.CreateMessage(connectedDb, Message{SenderID: 2, ReceiverID: 1, Content: "Hey!"})
	// db.CreateMessage(connectedDb, Message{SenderID: 1, ReceiverID: 2, Content: "You good?"})
	// db.CreateMessage(connectedDb, Message{SenderID: 2, ReceiverID: 1, Content: "Yup!"})
	// db.CreateMessage(connectedDb, Message{SenderID: 1, ReceiverID: 3, Content: "Hey?"})
	// fmt.Println(message.SenderID, message.ReceiverID)
	// users := db.RetrieveUsersIChatWith(connectedDb, 1)
	// fmt.Println(users)
}
