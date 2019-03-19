package main

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"golang.org/x/crypto/scrypt"
)

// User data type
type User struct {
	Email string
	Hash  []byte            // Password hash
	Salt  []byte            // Password salt
	Data  map[string]string // Additional data
}

// Server response
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

	switch req.Form.Get("command") {
	case SIGNUP:
		user := User{}
		user.Email = req.Form.Get("email")
		fmt.Println("Email is " + user.Email)

		// 16 byte (128 bit) random salt
		user.Salt = make([]byte, 16)
		rand.Read(user.Salt)

		user.Data = make(map[string]string)

		// Get private and public keys
		user.Data["public"] = req.Form.Get("pubKey")
		user.Data["private"] = req.Form.Get("privKey")

		// Password hash
		password := decode64(req.Form.Get("password"))
		user.Hash, _ = scrypt.Key(password, user.Salt, 16384, 8, 1, 32)
		respond(w, true, "Attempted to register with email "+user.Email)
		// TODO check if user already exists
	case LOGIN:
	}
}

// Write response in JSON format
func respond(w io.Writer, ok bool, message string) {
	response := Response{Ok: ok, Message: message}
	JSONResponse, err := json.Marshal(&response)
	checkError(err)
	w.Write(JSONResponse)
}
