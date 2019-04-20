package main

import "crypto/rand"

type User struct {
	Email     string
	Name      string
	Hash      []byte            // Password hash
	Salt      []byte            // Password salt
	Data      map[string]string // Additional data
	Passwords map[string]string // Key: url, Value: password
}

func (user User) login() {
	storedUser := ReadUser(user.Name)

	if storedUser.Name == user.Name {
		LogInfo("Logged in with user " + user.Name)
	} else {
		LogInfo("Could not log in user " + user.Name)
	}
}

func (user User) signup() {
	if users[user.Name] != 0 {
		LogInfo("User " + user.Name + " already exists and cannot be signed up")
	} else {
		WriteUser(user)
		LogInfo("Signed up user " + user.Name)
	}
}

func (user User) generatePassword(url string) {
	pBytes := make([]byte, 9)
	_, err := rand.Read(pBytes)
	CheckError(err)
	password := Encode64(pBytes)
	user.Passwords[url] = password
}
