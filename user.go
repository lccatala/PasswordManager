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
	Data      map[string]string // public and private keys
	DataKey   []byte            // Key to encrypt user's data
	Passwords map[string]string // Key: url, Value: password
	Notes     []SecureNote      // Secure notes
}

// Login authenticates the user that calls it on the server
func (user *User) Login() (resp Response) {
	storedUser := User{}
	storedUser.Read(user.UUID.String(), user.DataKey)
	LogTrace("Gotten name " + storedUser.Name)
	if user.Passwords == nil {
		user.Passwords = make(map[string]string)
	}

	resp.UserData.Passwords = make(map[string]string)
	resp.UserData.Name = user.Name
	resp.UserData.UUID = user.UUID
	if storedUser.Name == user.Name && bytes.Equal(storedUser.Hash, user.Hash) {
		LogInfo("User logged in")
		resp.Ok = true

		for k, v := range storedUser.Passwords {
			resp.UserData.Passwords[k] = v
		}
	} else {
		LogInfo("User failed logging in")
		resp.Ok = false
	}

	return
}

// Signup creates a new user with the data of the one that calls it
func (user *User) Signup() (resp Response) {
	resp.UserData.Name = user.Name
	resp.UserData.UUID = user.UUID
	if users[user.UUID.String()] {
		resp.Ok = false
		LogInfo("Failed attempt to sign up user")
	} else {
		resp.Ok = true
		user.Passwords = make(map[string]string)
		user.WriteToJSON()
		user.WriteToList()
		users[user.UUID.String()] = true
		LogInfo("Signed up user")
	}

	return
}

// AddNote adds a SecureNote to the calling user
func (user *User) AddNote(noteTitle string, noteContent string) (resp Response) {
	newNote := SecureNote{noteTitle, noteContent}
	user.Notes = append(user.Notes, newNote)
	resp.Ok = true
	return
}

// AddPassword adds a password for a given url to the calling user
func (user *User) AddPassword(password string, url string) (resp Response) {
	if user.Passwords == nil {
		user.Passwords = make(map[string]string)
	}

	user.Passwords[url] = password
	resp.Ok = true
	return
}

// Read reads from a json file into it's calling user
func (user *User) Read(uuid string, key []byte) {
	if user.Passwords == nil {
		user.Passwords = make(map[string]string)
	}
	fileData, _ := ioutil.ReadFile("users/" + uuid + ".json")
	fileData = Decrypt(fileData, key)
	json.Unmarshal([]byte(fileData), &user)
}

// WriteToJSON saves the calling user's data to it's individual json file
func (user *User) WriteToJSON() {
	filename := user.UUID.String()
	fileData, err := json.MarshalIndent(user, "", "  ")
	fileData = Encrypt(fileData, user.DataKey)
	CheckError(err)
	err = ioutil.WriteFile("users/"+filename+".json", fileData, 0644)
	CheckError(err)
}

// WriteToList appends the user's UUID to the server's user list
func (user *User) WriteToList() {
	f, err := os.OpenFile("users/users.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	CheckError(err)
	defer f.Close()
	data := []byte(user.UUID.String())
	_, err = f.Write(data)
	CheckError(err)
	_, err = f.Write([]byte("\n"))
	CheckError(err)
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
	user.Data["loginKey"] = req.Form.Get("loginKey")
	user.DataKey = Decode64(req.Form.Get("dataKey"))

	// Get password hash
	password := Decode64(req.Form.Get("loginKey"))
	user.Hash, _ = scrypt.Key(password, user.Salt, 16384, 8, 1, 32)
}

// HashPasswordFromFile gets the correct hash for a user by reading the salt from it's file
func (user *User) HashPasswordFromFile() {
	storedUser := User{}
	storedUser.Read(user.UUID.String(), user.DataKey)
	user.Hash, _ = scrypt.Key(Decode64(user.Data["loginKey"]), storedUser.Salt, 16384, 8, 1, 32)
}
