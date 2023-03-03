package main

import (
	"distributed-chat/minion/db"
	"fmt"
)

func main() {
	connectedDb := db.InitDb()
	fmt.Println(db.RetrieveMessageById(connectedDb, 2))
}
