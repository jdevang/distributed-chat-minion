package main

import (
	"bytes"
	"distributed-chat/minion/auth"
	"distributed-chat/minion/db"
	"distributed-chat/minion/structs"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Message = structs.Message
type User = structs.User
type HTTPStatusMessage = structs.HTTPStatusMessage
type Minion = structs.Minion

type MessageApiWrapper struct {
	ApiKey       string
	SenderName   string
	ReceiverName string
	Content      string
}

type UserApiWrapper struct {
	ApiKey   string
	Username string
}
type UsersApiWrapper struct {
	ApiKey       string
	Username     string
	ReceiverName string
}

var dbInstance = db.InitDb()
var minionName = os.Getenv("MINION_NAME")
var minionUrlIdentifier = os.Getenv("MINION_URL_IDENTIFIER")
var masterUrl = os.Getenv("MASTER_URL")

func main() {
	router := gin.Default()
	db.CreateDbFromSchema(dbInstance)
	setupMinionAtMaster()
	router.POST("/register", register)
	router.POST("/login", login)
	router.POST("/send", send)
	router.POST("/receive", receive)
	router.POST("/checkNewMessages", checkNewMessages)
	router.POST("/getUsersIChatWith", getUsersIChatWith)
	router.POST("/getMessagesBetweenMeAndUser", getMessagesBetweenMeAndUser)

	router.GET("/users", users)
	router.GET("/messages", messages)
	router.GET("/alive", alive)

	router.SetTrustedProxies(nil)

	router.Run(":8080")
}

func setupMinionAtMaster() {
	minion, err := db.RetrieveSelf(dbInstance, minionName)
	if err != nil {
		minion, _ = db.CreateMinion(dbInstance, Minion{MinionName: minionName, UrlIdentifier: minionUrlIdentifier, SetAtMaster: false})
	}

	if !minion.SetAtMaster {
		url := masterUrl + "/registerMinion"
		payload, _ := json.Marshal(minion)
		req, _ := http.NewRequest("POST", url, bytes.NewBuffer(payload))
		req.Header.Set("Content-Type", "application/json")
		client := &http.Client{}
		_, err := client.Do(req)
		if err != nil {
			os.Exit(1)
		}
		minion.SetAtMaster = true
		db.UpdateMinion(dbInstance, minion)
	}
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
	message.SenderMinionUrlIdentifier = minionUrlIdentifier
	user, err := db.RetrieveUserByName(dbInstance, username)
	if err != nil {
		fmt.Println("Error verifying apiKey")
		c.IndentedJSON(http.StatusUnauthorized, HTTPStatusMessage{Message: "invalid apikey"})
		return
	}

	if auth.VerifyApiKey(user, apiKey) {
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
			message.ReceiverMinionUrlIdentifier = minionUrlIdentifier
		}

		status = messageSender(user, message)

		if status == "success" {
			fmt.Println("Success")
			c.IndentedJSON(http.StatusOK, HTTPStatusMessage{Message: "Success"})
		} else if status == "invalid" {
			fmt.Println("Error: Invalid Message Body")
			c.IndentedJSON(http.StatusBadRequest, HTTPStatusMessage{Message: "invalid message"})
			return
		} else {
			fmt.Println("Error: Failed to get response from receiver")
			c.IndentedJSON(http.StatusRequestTimeout, HTTPStatusMessage{Message: "failed to get response from receiver"})
			return
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
	url := masterUrl + "/retrieveMinionUrlIdentifier"
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

	body, err := io.ReadAll(resp.Body)

	if err != nil {

		return "invalid"
	}

	result := struct {
		ReceiverMinionUrlIdentifier string
	}{}

	if err := json.Unmarshal(body, &result); err != nil { // Parse []byte to go struct pointer
		fmt.Println("Can not unmarshal JSON")
	}

	return result.ReceiverMinionUrlIdentifier
}

func messageSender(user User, message Message) string {
	var url string
	if message.ReceiverMinionUrlIdentifier != minionUrlIdentifier {
		url = "https://" + message.ReceiverMinionUrlIdentifier + ".minion.chat.junglesucks.com/receive"

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
	user, err := db.RetrieveUserByName(dbInstance, message.ReceiverName)
	if err != nil {
		fmt.Println("No user")
		c.IndentedJSON(http.StatusNotAcceptable, HTTPStatusMessage{Message: "no such user here"})
		return
	}

	status := messageSender(user, message)

	if status == "success" {
		fmt.Println("Success")
		c.IndentedJSON(http.StatusOK, HTTPStatusMessage{Message: "Success"})
	} else if status == "invalid" {
		fmt.Println("Error: Invalid Message Body")
		c.IndentedJSON(http.StatusBadRequest, HTTPStatusMessage{Message: "invalid message"})
		return
	} else {
		fmt.Println("Error: Failed to get response from receiver")
		c.IndentedJSON(http.StatusRequestTimeout, HTTPStatusMessage{Message: "failed to get response from receiver"})
		return
	}

	_, err = db.CreateMessage(dbInstance, message)

	if err != nil {
		message.ReceiverMinionUrlIdentifier = ""
		db.CreateMessage(dbInstance, message)
	}
}

func getUsersIChatWith(c *gin.Context) {
	var userApiWrapper UserApiWrapper
	err := c.BindJSON(&userApiWrapper)
	if err != nil {
		fmt.Println("Error in reading json")
		c.IndentedJSON(http.StatusBadRequest, HTTPStatusMessage{Message: "faulty request"})
		return
	}
	user, err := db.RetrieveUserByName(dbInstance, userApiWrapper.Username)
	if err != nil {
		fmt.Println("Error verifying apiKey")
		c.IndentedJSON(http.StatusUnauthorized, HTTPStatusMessage{Message: "invalid apikey"})
		return
	}

	if auth.VerifyApiKey(user, userApiWrapper.ApiKey) {
		users := db.RetrieveUsersIChatWith(dbInstance, user.Username)
		c.IndentedJSON(http.StatusOK, users)

	} else {
		fmt.Println("Error: Invalid ApiKey")
		c.IndentedJSON(http.StatusUnauthorized, HTTPStatusMessage{Message: "invalid apikey"})
		return
	}
}

func getMessagesBetweenMeAndUser(c *gin.Context) {
	var usersApiWrapper UsersApiWrapper
	err := c.BindJSON(&usersApiWrapper)
	if err != nil {
		fmt.Println("Error in reading json")
		c.IndentedJSON(http.StatusBadRequest, HTTPStatusMessage{Message: "faulty request"})
		return
	}
	user1, err := db.RetrieveUserByName(dbInstance, usersApiWrapper.Username)
	if err != nil {
		fmt.Println("Error verifying apiKey")
		c.IndentedJSON(http.StatusUnauthorized, HTTPStatusMessage{Message: "invalid apikey"})
		return
	}
	user2, err := db.RetrieveUserByName(dbInstance, usersApiWrapper.ReceiverName)
	if err != nil {
		fmt.Println("No User")
		c.IndentedJSON(http.StatusUnauthorized, []string{})
		return
	}

	if auth.VerifyApiKey(user1, usersApiWrapper.ApiKey) {
		messages := db.RetrieveAllMessagesBetweenUsers(dbInstance, user1.Username, user2.Username)
		c.IndentedJSON(http.StatusOK, messages)
	} else {
		fmt.Println("Error: Invalid ApiKey")
		c.IndentedJSON(http.StatusUnauthorized, HTTPStatusMessage{Message: "invalid apikey"})
		return
	}
}

func checkNewMessages(c *gin.Context) {
	var usersApiWrapper UsersApiWrapper
	err := c.BindJSON(&usersApiWrapper)
	if err != nil {
		fmt.Println("Error in reading json")
		c.IndentedJSON(http.StatusBadRequest, HTTPStatusMessage{Message: "faulty request"})
		return
	}
	user, err := db.RetrieveUserByName(dbInstance, usersApiWrapper.Username)
	if err != nil {
		fmt.Println("Error verifying apiKey")
		c.IndentedJSON(http.StatusUnauthorized, HTTPStatusMessage{Message: "invalid apikey"})
		return
	}

	if auth.VerifyApiKey(user, usersApiWrapper.ApiKey) {
		message, _ := db.RetrieveLatestMessageBySenderAndReceiver(dbInstance, usersApiWrapper.ReceiverName, usersApiWrapper.Username)
		c.IndentedJSON(http.StatusOK, message)
		return
	}
}

func users(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, db.RetrieveAllUsers(dbInstance))
}

func messages(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, db.RetrieveAllMessagesBetweenUsers(dbInstance, "testusername1", "testusername2"))
}

func alive(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, HTTPStatusMessage{Message: "I'm Alive!"})
}
