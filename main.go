package main

import (
	"distributed-chat/minion/structs"

	"github.com/gin-gonic/gin"
)

type Message = structs.Message
type User = structs.User

func main() {
	router := gin.Default()
	router.GET("/register")

	router.Run("localhost:8080")
}
