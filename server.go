package main

import (
	"encoding/json"
	"io"
	"net/http"
	"os"

	"github.com/google/uuid"
)

// Response from server
type Response struct {
	Ok       bool
	UserData User
}

// Login and signup modes
const (
	LOGIN  = "login"
	SIGNUP = "signup"
	ADD    = "add"
)

// KEY for encrypting user list
var KEY []byte

// List of signed up users
var users map[string]bool

func startServer() {
	LogInfo("Starting server...")

	KEY = parseKey([]byte(os.Args[2]))
	users = ReadAllUsers()

	http.HandleFunc("/", handler)
	LogInfo("Done")
	err := http.ListenAndServeTLS(":10443", "cert.pem", "key.pem", nil)
	CheckError(err)
}

func handler(w http.ResponseWriter, req *http.Request) {
	req.ParseForm()
	w.Header().Set("Content-Type", "text/plain") // Standard header

	command := req.Form.Get("command")
	user := User{}
	if command != ADD {
		user.GetData(req)
	}

	resp := Response{}
	switch command {
	case SIGNUP:
		resp = user.Signup()
	case LOGIN:
		user.HashPasswordFromFile(req.Form.Get("password"))
		resp = user.Login()
	case ADD:
		newPassword := req.Form.Get("password")
		newURL := req.Form.Get("url")

		user := User{}
		path, _ := uuid.Parse(req.Form.Get("uuid"))

		user.Read(path.String())
		resp = user.AddPassword(newPassword, newURL)
		user.WriteToJSON(path.String())
	}

	respond(w, resp)
}

// Write response in JSON format
func respond(w io.Writer, resp Response) {
	JSONResp, err := json.Marshal(&resp)
	CheckError(err)
	w.Write(JSONResp)
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
