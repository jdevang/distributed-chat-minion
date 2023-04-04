package main

import (
	"bytes"
	"distributed-chat/minion/auth"
	"distributed-chat/minion/db"
	"distributed-chat/minion/structs"
	"distributed-chat/minion/utils"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Message = structs.Message
type User = structs.User
type HTTPStatusMessage = structs.HTTPStatusMessage

type MessageApiWrapper struct {
	ApiKey       string
	SenderName   string
	ReceiverName string
	Content      string
}

var dbInstance = db.InitDb()
var config, _ = utils.ReadConfigFile("config.yaml")

func main() {
	router := gin.Default()
	db.CreateDbFromSchema(dbInstance)

	router.POST("/register", register)
	router.POST("/login", login)
	router.GET("/users", users)
	router.POST("/send", send)
	router.POST("/receive", receive)

	router.Run("localhost:8080")
}

func register(c *gin.Context) {
	var newUser User
	err := c.BindJSON(&newUser)
	if err != nil {
		fmt.Println("Error in reading json")
		c.IndentedJSON(http.StatusBadRequest, HTTPStatusMessage{Message: "faulty request"})
		return
	}

	newUser.Password, err = auth.HashPassword(newUser.Password)
	if err != nil {
		fmt.Println("Error in generating password hash")
		c.IndentedJSON(http.StatusConflict, HTTPStatusMessage{Message: "invalid password"})
		return
	}

	newUser.ApiKey, err = auth.GenApiKey(newUser)
	if err != nil {
		fmt.Println("Error in generating Api key")
		c.IndentedJSON(http.StatusConflict, HTTPStatusMessage{Message: "could not create user"})
		return
	}

	newUser, err = db.CreateUser(dbInstance, newUser)
	if err != nil {
		c.IndentedJSON(http.StatusConflict, HTTPStatusMessage{Message: "username not available"})
		return
	}
	c.IndentedJSON(http.StatusCreated, HTTPStatusMessage{Message: "user created"})
}

func login(c *gin.Context) {
	var user User
	err := c.BindJSON(&user)
	if err != nil {
		fmt.Println("Error in reading json")
		c.IndentedJSON(http.StatusBadRequest, HTTPStatusMessage{Message: "faulty request"})
		return
	}

	dbUser, err := db.RetrieveUserByName(dbInstance, user.Username)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		fmt.Println("Error: No such user")
		c.IndentedJSON(http.StatusConflict, HTTPStatusMessage{Message: "no such user"})
		return
	}
	if auth.CheckPasswordHash(user, dbUser.Password) {
		c.IndentedJSON(http.StatusOK, HTTPStatusMessage{ApiKey: dbUser.ApiKey})
		return
	} else {
		fmt.Println("Error: Password doesn't match")
		fmt.Println(dbUser.Password)
		fmt.Println(dbUser.Username)
		fmt.Println(user.Username)
		fmt.Println(user.Password)
		c.IndentedJSON(http.StatusConflict, HTTPStatusMessage{Message: "invalid password"})
		return
	}
}

func send(c *gin.Context) {
	var newMessageApiWrapper MessageApiWrapper
	err := c.BindJSON(&newMessageApiWrapper)
	if err != nil {
		fmt.Println("Error in reading json")
		c.IndentedJSON(http.StatusBadRequest, HTTPStatusMessage{Message: "faulty request"})
		return
	}

	username, apiKey := newMessageApiWrapper.SenderName, newMessageApiWrapper.ApiKey
	message := Message{SenderName: newMessageApiWrapper.SenderName, ReceiverName: newMessageApiWrapper.ReceiverName, Content: newMessageApiWrapper.Content}
	message.SenderMinionUrlIdentifier = config["minionUrlIdentifier"]

	if auth.VerifyApiKey(username, apiKey) {
		status := ""
		user, err := db.RetrieveUserByName(dbInstance, message.ReceiverName)
		if err != nil {
			precedenceMessage, err := db.RetrieveLatestMessageBySenderAndReceiver(dbInstance, message.SenderName, message.ReceiverName)
			if err != nil || precedenceMessage.ReceiverMinionUrlIdentifier == "" {
				message.ReceiverMinionUrlIdentifier = retrieveMinionUrlIdentifierFromMaster(message.ReceiverName)
			} else {
				message.ReceiverMinionUrlIdentifier = precedenceMessage.ReceiverMinionUrlIdentifier
			}

		} else { //Receiver is coupled to this minion
			message.ReceiverMinionUrlIdentifier = config["minionUrlIdentifier"]
		}

		status = messageSender(user, message)

		if status == "success" {
			fmt.Println("Success")
			c.IndentedJSON(http.StatusOK, HTTPStatusMessage{Message: "Success"})
		} else if status == "invalid" {
			fmt.Println("Error: Invalid Message Body")
			c.IndentedJSON(http.StatusBadRequest, HTTPStatusMessage{Message: "invalid message"})
		} else {
			fmt.Println("Error: Failed to get response from receiver")
			c.IndentedJSON(http.StatusRequestTimeout, HTTPStatusMessage{Message: "failed to get response from receiver"})
		}

		_, err = db.CreateMessage(dbInstance, message)

		if err != nil {
			message.ReceiverMinionUrlIdentifier = ""
			db.CreateMessage(dbInstance, message)
		}

		return
	} else {
		fmt.Println("Error: Invalid ApiKey")
		c.IndentedJSON(http.StatusUnauthorized, HTTPStatusMessage{Message: "invalid apikey"})
		return
	}
}

func retrieveMinionUrlIdentifierFromMaster(username string) string {
	url := "https://master.example.com/retrieveMinionUrlIdentifier"
	payload, err := json.Marshal(username)
	if err != nil {
		return "invalid"
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return "invalid"
	}

	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil || resp.Status != "200 OK" {

		return "timeout"
	}

	body, err := ioutil.ReadAll(resp.Body)

	result := struct {
		ReceiverMinionUrlIdentifier string
	}{}

	if err := json.Unmarshal(body, &result); err != nil { // Parse []byte to go struct pointer
		fmt.Println("Can not unmarshal JSON")
	}

	return result.ReceiverMinionUrlIdentifier
}

func messageSender(user User, message Message) string {
	url := "https://" + user.ClientUrlIdentifier + ".messageclient.example.com/receive"
	payload, err := json.Marshal(message)
	if err != nil {
		return "invalid"
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return "invalid"
	}

	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil || resp.Status != "200 OK" {

		return "timeout"
	}

	return "success"
}

func receive(c *gin.Context) {
	var message Message
	err := c.BindJSON(&message)
	if err != nil {
		fmt.Println("Error in reading json")
		c.IndentedJSON(http.StatusBadRequest, HTTPStatusMessage{Message: "faulty request"})
		return
	}
}

func users(c *gin.Context) {
	c.IndentedJSON(http.StatusCreated, db.RetrieveAllUsers(dbInstance))
}
