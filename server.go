package main

import (
	"encoding/json"
	"io"
	"net/http"
	"os"

	"github.com/google/uuid"
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
	if command != ADDPASS {
		user.GetData(req)
	}

	resp := Response{}
	switch command {
	case SIGNUP:
		resp = user.Signup()
	case LOGIN:
		user.HashPasswordFromFile()
		resp = user.Login()
	case ADDPASS:
		newPassword := req.Form.Get("password")
		newURL := req.Form.Get("url")

		user := User{}
		//path, _ := uuid.Parse(req.Form.Get("uuid"))

		//user.Read(path.String(), string(privKey))
		resp = user.AddPassword(newPassword, newURL)
		user.WriteToJSON()
	case ADDNOTE:
		noteTitle := req.Form.Get("noteTitle")
		noteContent := req.Form.Get("noteContent")

		user := User{}
		path, _ := uuid.Parse(req.Form.Get("uuid"))

		user.Read(path.String(), user.DataKey)
		resp = user.AddNote(noteTitle, noteContent)
		user.WriteToJSON()
	}

	respond(w, resp)
}

// Write response in JSON format
func respond(w io.Writer, resp Response) {
	JSONResp, err := json.Marshal(&resp)
	CheckError(err)
	w.Write(JSONResp)
}
