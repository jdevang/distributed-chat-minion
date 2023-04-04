package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"distributed-chat/minion/structs"
	"encoding/base64"

	"golang.org/x/crypto/bcrypt"
)

type User = structs.User
type Message = structs.Message

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CheckPasswordHash(user User, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(user.Password))
	return err == nil
}

func GenApiKey(user User) (string, error) {
	keyBytes := make([]byte, 32)
	_, err := rand.Read(keyBytes)
	if err != nil {
		return "", err
	}

	hash := sha256.Sum256(keyBytes)

	key := base64.StdEncoding.EncodeToString(hash[:])
	return key, err
}

func VerifyApiKey(username string, apiKey string) bool {
	return true
}
