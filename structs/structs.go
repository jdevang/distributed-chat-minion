package structs

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Name     string
	Password string
}

type Message struct {
	gorm.Model
	SenderID         uint
	ReceiverID       uint
	SenderMinionID   uint
	ReceiverMinionID uint
	Content          string
}

type Gru struct {
	gorm.Model
	Name string
	Url  string
}
