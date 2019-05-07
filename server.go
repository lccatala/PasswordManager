package main

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
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
	user.GetData(req)

	var ok bool
	var message string
	switch req.Form.Get("command") {
	case SIGNUP:
		ok, message = user.Signup()
	case LOGIN:
		ok, message = user.Login()
	}

	respond(w, ok, message)
}

// Write response in JSON format
func respond(w io.Writer, ok bool, message string) {
	response := Response{Ok: ok, Message: message}
	JSONResponse, err := json.Marshal(&response)
	CheckError(err)
	w.Write(JSONResponse)
}

// Modify server key so it has an appropiate length
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
