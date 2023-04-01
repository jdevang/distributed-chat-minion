package main

import (
	"distributed-chat/minion/db"
	"distributed-chat/minion/structs"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Message = structs.Message
type User = structs.User

var dbInstance = db.InitDb()

func main() {
	router := gin.Default()
	router.POST("/register", register)

	router.Run("localhost:8080")
}

func register(c *gin.Context) {
	var newUser User
	if err := c.BindJSON(&newUser); err != nil {
		return
	}
	db.CreateUser(dbInstance, newUser)
	c.IndentedJSON(http.StatusCreated, newUser)
}
