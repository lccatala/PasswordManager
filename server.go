package main

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
)

// Response from server
type Response struct {
	Ok        bool
	Message   string
	Username  string
	Passwords map[string]string
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
	users = ReadAllUsers()

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

	resp := Response{}
	switch req.Form.Get("command") {
	case SIGNUP:
		resp = user.Signup()
		LogTrace("User name: " + user.Name)
		LogTrace("Resp name: " + resp.Username)
	case LOGIN:
		user.HashPasswordFromFile(req.Form.Get("password"))
		resp = user.Login()
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
