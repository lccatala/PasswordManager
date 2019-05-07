package main

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/google/uuid"
	"golang.org/x/crypto/scrypt"
)

// User data type
type User struct {
	UUID      uuid.UUID
	Email     string
	Name      string
	Hash      []byte            // Password hash
	Salt      []byte            // Password salt
	Data      map[string]string // Additional data (public and private keys)
	Passwords map[string]string // Key: url, Value: password
}

// Login authenticates the user that calls it on the server
func (user *User) Login() (success bool, message string) {
	storedUser := User{}
	storedUser.Read(user.UUID.String())

	if storedUser.Name == user.Name && bytes.Equal(storedUser.Hash, user.Hash) {
		message = "Logged in with user " + user.Name
	} else {
		message = "Could not log in user " + user.Name
	}

	LogInfo(message)
	return
}

// Signup creates a new user with the data of the one that calls it
func (user *User) Signup() (success bool, message string) {
	if users[user.Name] {
		message = "User " + user.Name + " already exists and cannot be signed up"
		success = false
	} else {
		message = "Signed up user " + user.Name
		success = true
		user.Write()
	}
	LogInfo(message)
	return
}

// EncryptFields encrypts the calling user's fields with it's private key
func (user *User) EncryptFields() {
	//bytekey := []byte(user.Data["privKey"]) // TODO use another key for encrypting

	user.Email = string(Encrypt([]byte(user.Email), KEY))
	user.Name = string(Encrypt([]byte(user.Name), KEY))

	for k, v := range user.Passwords {
		user.Passwords[k] = string(Encrypt([]byte(v), KEY))
	}
}

// Read reads from a json file into it's calling user
func (user *User) Read(username string) {
	fileData, _ := ioutil.ReadFile("users/" + uuid + ".json")
	json.Unmarshal([]byte(fileData), &user)
}

// Write saves the calling user's data to the server's user list (encrypted) and to it's individual json
func (user *User) Write() {

	// Save to user's individual JSON
	user.EncryptFields()
	fileData, err := json.MarshalIndent(user, "", "  ")
	CheckError(err)
	err = ioutil.WriteFile("users/"+user.UUID.String()+".json", fileData, 0644)
	CheckError(err)

	// Add user to users list
	f, err := os.OpenFile("users/users.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	CheckError(err)
	defer f.Close()
	data := Encrypt([]byte(user.Name), KEY)
	_, err = f.Write(data)
	CheckError(err)
	_, err = f.Write([]byte("\n"))
	CheckError(err)
}

// GeneratePassword creates and saves a random password for a given URL
func (user *User) GeneratePassword(url string) {
	pBytes := make([]byte, 9)
	_, err := rand.Read(pBytes)
	CheckError(err)
	password := Encode64(pBytes)
	user.Passwords[url] = password
}

// GetData reads a user's fields from an http request into it's calling user
func (user *User) GetData(req *http.Request) {
	user.Name = req.Form.Get("name")
	var err error
	user.UUID, err = uuid.FromBytes([]byte(user.Name)[:16])
	CheckError(err)
	user.Email = req.Form.Get("email")

	// 16 byte (128 bit) random salt
	user.Salt = make([]byte, 16)
	rand.Read(user.Salt)

	// Get private and public keys
	user.Data = make(map[string]string)
	user.Data["pubKey"] = req.Form.Get("pubKey")
	user.Data["privKey"] = req.Form.Get("privKey")

	// Get password hash
	password := Decode64(req.Form.Get("password"))
	user.Hash, _ = scrypt.Key(password, user.Salt, 16384, 8, 1, 32)
}
