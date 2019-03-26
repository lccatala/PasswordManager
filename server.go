package main

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"golang.org/x/crypto/scrypt"
)

// User data type
type User struct {
	Email string
	Name  string
	Hash  []byte            // Password hash
	Salt  []byte            // Password salt
	Data  map[string]string // Additional data
}

// Response from server
type Response struct {
	Ok      bool
	Message string
}

func startServer() {
	fmt.Println("Server started")
	http.HandleFunc("/", handler)
	checkError(http.ListenAndServeTLS(":10443", "cert.pem", "key.pem", nil))
}

func handler(w http.ResponseWriter, req *http.Request) {
	req.ParseForm()
	w.Header().Set("Content-Type", "text/plain") // Standard header

	user := User{}
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
	password := decode64(req.Form.Get("password"))
	user.Hash, _ = scrypt.Key(password, user.Salt, 16384, 8, 1, 32)

	switch req.Form.Get("command") {
	case SIGNUP:
		handleSignup(user)
	case LOGIN:

	}
}

func handleSignup(user User) {
	if !userExists(user.Name) {
		writeToFile(user)
	} else {
		fmt.Printf("User " + user.Name + " already exists\n")
	}
}

// Check if username is already taken
func userExists(name string) bool {
	files, err := ioutil.ReadDir("users")
	checkError(err)
	for _, f := range files {
		if f.Name() == name+".json" {
			return true
		}
	}
	return false
}

// Write user struct to json file
func writeToFile(user User) {
	fileData, err := json.MarshalIndent(user, "", "  ")
	checkError(err)
	err = ioutil.WriteFile("users/"+user.Name+".json", fileData, 0644)
	checkError(err)
}

// Write response in JSON format
func respond(w io.Writer, ok bool, message string) {
	response := Response{Ok: ok, Message: message}
	JSONResponse, err := json.Marshal(&response)
	checkError(err)
	w.Write(JSONResponse)
}
