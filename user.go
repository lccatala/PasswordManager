package main

import (
	"crypto/rand"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"

	"golang.org/x/crypto/scrypt"
)

// User data type
type User struct {
	Email     string
	Name      string
	Hash      []byte            // Password hash
	Salt      []byte            // Password salt
	Data      map[string]string // Additional data
	Passwords map[string]string // Key: url, Value: password
}

func (user User) login() {
	storedUser := ReadUser(user.Name)

	if storedUser.Name == user.Name {
		LogInfo("Logged in with user " + user.Name)
	} else {
		LogInfo("Could not log in user " + user.Name)
	}
}

func (user User) signup() {
	if users[user.Name] {
		LogInfo("User " + user.Name + " already exists and cannot be signed up")
	} else {
		user.write()
		LogInfo("Signed up user " + user.Name)
	}
}

func (user User) encryptFields() {
	bytekey := []byte(user.Data["privKey"])

	user.Email = string(Encrypt([]byte(user.Email), bytekey))
	user.Name = string(Encrypt([]byte(user.Name), bytekey))

	for k, v := range user.Passwords {
		user.Passwords[k] = string(Encrypt([]byte(v), bytekey))
	}
}

func (user User) write() {

	// Save to user's individual JSON
	user.encryptFields()
	fileData, err := json.MarshalIndent(user, "", "  ")
	CheckError(err)
	err = ioutil.WriteFile("users/"+user.Name+".json", fileData, 0644)
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

func (user User) generatePassword(url string) {
	pBytes := make([]byte, 9)
	_, err := rand.Read(pBytes)
	CheckError(err)
	password := Encode64(pBytes)
	user.Passwords[url] = password
}

func (user User) getData(req *http.Request) {
	user.Name = req.Form.Get("name")
	user.Email = req.Form.Get("email")

	// 16 byte (128 bit) random salt
	user.Salt = make([]byte, 16)
	rand.Read(user.Salt)

	// Get private and public keys
	user.Data = make(map[string]string)
	user.Data["public"] = req.Form.Get("pubKey")
	user.Data["private"] = req.Form.Get("privKey")

	// Get password hash
	password := Decode64(req.Form.Get("password"))
	user.Hash, _ = scrypt.Key(password, user.Salt, 16384, 8, 1, 32)
}
