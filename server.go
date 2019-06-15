package main

import (
	"encoding/json"
	"io"
	"net/http"
)

// SecureNote stores a typical note with title and content, but it's secure
type SecureNote struct {
	Title   string
	Content string
}

// Response from server
type Response struct {
	Ok       bool
	UserData User
}

// Login and signup modes
const (
	LOGIN   = "login"
	SIGNUP  = "signup"
	ADDPASS = "addpass"
	ADDNOTE = "addnote"
)

// List of signed up users
var users map[string]bool

func startServer() {
	LogInfo("Starting server...")

	//KEY = parseKey([]byte(os.Args[2]))
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

	user.GetAuthData(req)

	resp := Response{}
	switch command {
	case SIGNUP:
		resp = user.Signup()
	case LOGIN:
		resp = user.Login()
	case ADDPASS:
		resp = user.AddPassword(req)
	case ADDNOTE:
		resp = user.AddNote(req)
	}

	respond(w, resp)
}

// Write response in JSON format
func respond(w io.Writer, resp Response) {
	JSONResp, err := json.Marshal(&resp)
	CheckError(err)
	w.Write(JSONResp)
}
