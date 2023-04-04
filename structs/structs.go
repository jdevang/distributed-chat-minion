package structs

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Username            string `gorm:"uniqueIndex:idx_name,sort:desc"`
	Password            string
	ApiKey              string `gorm:"uniqueIndex:idx_api"`
	ClientUrlIdentifier string
}

type Message struct {
	gorm.Model
	SenderName                  string
	ReceiverName                string
	SenderMinionUrlIdentifier   string
	ReceiverMinionUrlIdentifier string
	Content                     string
}

type Gru struct {
	gorm.Model
	Name          string
	UrlIdentifier string `gorm:"uniqueIndex:idx_gru"`
}

type Minion struct {
	gorm.Model
	Name          string
	UrlIdentifier string `gorm:"uniqueIndex:idx_minion"`
}

type HTTPStatusMessage struct {
	Message string
	ApiKey  string
}
