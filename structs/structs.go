package structs

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Username            string `gorm:"uniqueIndex:idx_username"`
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

type Minion struct {
	gorm.Model
	MinionName    string
	UrlIdentifier string `gorm:"uniqueIndex:idx_minion"`
	SetAtMaster   bool
}

type HTTPStatusMessage struct {
	Message string
	ApiKey  string
}
