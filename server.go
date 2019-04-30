package main

import (
	"crypto/rand"
	"encoding/json"
	"io"
	"net/http"
	"os"

	"golang.org/x/crypto/scrypt"
)

// Response from server
type Response struct {
	Ok      bool
	Message string
}

// Login and signup modes
const (
	LOGIN  = "login"
	SIGNUP = "signup"
)

// KEY for encrypting user list
var KEY []byte

// List of signed up users
var users map[string]bool

func startServer() {
	LogInfo("Starting server...")

	KEY = parseKey([]byte(os.Args[2]))
	users = ReadAllUsers(KEY)

	http.HandleFunc("/", handler)
	LogInfo("Done")
	err := http.ListenAndServeTLS(":10443", "cert.pem", "key.pem", nil)
	CheckError(err)
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
	password := Decode64(req.Form.Get("password"))
	user.Hash, _ = scrypt.Key(password, user.Salt, 16384, 8, 1, 32)

	switch req.Form.Get("command") {
	case SIGNUP:
		user.signup()
	case LOGIN:
		user.login()
	}
}

// Write response in JSON format
func respond(w io.Writer, ok bool, message string) {
	response := Response{Ok: ok, Message: message}
	JSONResponse, err := json.Marshal(&response)
	CheckError(err)
	w.Write(JSONResponse)
}

// Modify key so it has an appropiate length
func parseKey(key []byte) []byte {
	if len(key) > 16 {
		return key[0:16]
	} else if len(key) < 16 {
		var l = len(key)
		for i := 0; i < 16-l; i++ {
			key = append(key, key[i])
		}
	}
	return key
}
