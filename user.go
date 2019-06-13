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
func (user *User) Login() (resp Response) {
	storedUser := User{}
	storedUser.Read(user.UUID.String())
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
		user.Write(user.UUID.String())
		users[user.UUID.String()] = true
		LogInfo("Signed up user")
	}

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

// EncryptFields encrypts the calling user's fields with it's private key
func (user *User) EncryptFields() {
	//bytekey := []byte(user.Data["privKey"]) // TODO use another key for encrypting

	user.Email = Encode64(Encrypt([]byte(user.Email), KEY))
	user.Name = Encode64(Encrypt([]byte(user.Name), KEY))

	for k, v := range user.Passwords {
		user.Passwords[k] = Encode64(Encrypt([]byte(v), KEY))
	}
}

// DecryptFields decrypts the calling user's fields with it's private key
func (user *User) DecryptFields() {
	//bytekey := []byte(user.Data["privKey"]) // TODO use another key for encrypting

	user.Email = string(Decrypt(Decode64(user.Email), KEY))
	user.Name = string(Decrypt(Decode64(user.Name), KEY))

	for k, v := range user.Passwords {
		user.Passwords[k] = string(Decrypt(Decode64(v), KEY))
	}
}

// Read reads from a json file into it's calling user
func (user *User) Read(uuid string) {
	if user.Passwords == nil {
		user.Passwords = make(map[string]string)
	}
	fileData, _ := ioutil.ReadFile("users/" + uuid + ".json")
	json.Unmarshal([]byte(fileData), &user)
	user.DecryptFields()
}

// Write saves the calling user's data to the server's user list (encrypted) and to it's individual json
func (user *User) Write(filename string) {

	// Save to user's individual JSON
	user.EncryptFields()
	fileData, err := json.MarshalIndent(user, "", "  ")
	CheckError(err)
	err = ioutil.WriteFile("users/"+filename+".json", fileData, 0644)
	CheckError(err)

	// Add user to users list
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

	// Get password hash
	password := Decode64(req.Form.Get("password"))
	user.Hash, _ = scrypt.Key(password, user.Salt, 16384, 8, 1, 32)
}

// HashPasswordFromFile gets the correct hash for a user by reading the salt from it's file
func (user *User) HashPasswordFromFile(key string) {
	storedUser := User{}
	storedUser.Read(user.UUID.String())
	user.Hash, _ = scrypt.Key(Decode64(key), storedUser.Salt, 16384, 8, 1, 32)
}
