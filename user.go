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

	resp.UserData.Passwords = make(map[string]string)
	resp.UserData.Name = user.Name
	if storedUser.Name == user.Name && bytes.Equal(storedUser.Hash, user.Hash) {
		resp.Message = "Logged in with user "
		resp.Ok = true

		for k, v := range storedUser.Passwords {
			resp.UserData.Passwords[k] = v
		}
	} else {
		resp.Message = "Could not log in user "
		resp.Ok = false
	}

	user.DecryptFields()
	resp.Message += user.Name

	LogInfo(resp.Message)
	return
}

// Signup creates a new user with the data of the one that calls it
func (user *User) Signup() (resp Response) {
	resp.UserData.Name = user.Name
	if users[user.UUID.String()] {
		resp.Ok = false
		resp.Message = "Could not sign up repeated user "
	} else {
		resp.Ok = true
		user.Write()
		users[user.UUID.String()] = true
		resp.Message = "Signed up user "
	}

	user.DecryptFields()
	resp.UserData = *user

	LogInfo(resp.Message)
	return
}

// EncryptFields encrypts the calling user's fields with it's private key
func (user *User) EncryptFields() {
	//bytekey := []byte(user.Data["privKey"]) // TODO use another key for encrypting

	user.Email = string(Encrypt([]byte(user.Email), KEY))
	user.Email = Encode64([]byte(user.Email))

	user.Name = string(Encrypt([]byte(user.Name), KEY))
	user.Name = Encode64([]byte(user.Name))

	user.Hash = Encrypt(user.Hash, KEY)
	user.Hash = []byte(Encode64(user.Hash))

	for k, v := range user.Passwords {
		user.Passwords[k] = string(Encrypt([]byte(v), KEY))
	}
}

// DecryptFields decrypts the calling user's fields with it's private key
func (user *User) DecryptFields() {
	//bytekey := []byte(user.Data["privKey"]) // TODO use another key for encrypting

	user.Email = string(Decode64(user.Email))
	user.Email = string(Decrypt([]byte(user.Email), KEY))

	user.Name = string(Decode64(user.Name))
	user.Name = string(Decrypt([]byte(user.Name), KEY))

	user.Hash = Decode64(string(user.Hash))
	user.Hash = Decrypt([]byte(user.Hash), KEY)

	for k, v := range user.Passwords {
		user.Passwords[k] = string(Decrypt(Decode64(v), KEY))
	}
}

// Read reads from a json file into it's calling user
func (user *User) Read(uuid string) {
	fileData, _ := ioutil.ReadFile("users/" + uuid + ".json")
	json.Unmarshal([]byte(fileData), &user)
	user.DecryptFields()
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
	data := []byte(user.UUID.String())
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

// HashPasswordFromFile gets the correct hash for a user by reading the salt from it's file
func (user *User) HashPasswordFromFile(key string) {
	storedUser := User{}
	storedUser.Read(user.UUID.String())
	user.Hash, _ = scrypt.Key(Decode64(key), storedUser.Salt, 16384, 8, 1, 32)
}
