package util

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"os"
)

// Encode64 converts a []byte in base64 to a string
func Encode64(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

// Decode64 converts a string to a []byte in base64
func Decode64(s string) []byte {
	b, err := base64.StdEncoding.DecodeString(s)
	CheckError(err)
	return b
}

// WriteUser writes a user struct to a json file
func WriteUser(user User) {
	// TODO encrypt user data before saving it to file
	fileData, err := json.MarshalIndent(user, "", "  ")
	CheckError(err)
	err = ioutil.WriteFile("users/"+user.Name+".json", fileData, 0644)
	CheckError(err)
}

// ReadUser reads a user struct from a json file
func ReadUser(username string) (User, bool) {
	user := User{}
	fileData, err := ioutil.ReadFile(username + ".json")
	json.Unmarshal([]byte(fileData), &user)
	return user, err != nil
}

// ReadAllUsers reads all users contained in users/users.txt to the users map
func ReadAllUsers() {
	Users = make(map[string]int)
	file, err := os.Open("users/users.txt")
	CheckError(err)
	defer file.Close()

	scanner := bufio.NewScanner(file)

	id := 0
	for scanner.Scan() {
		Users[scanner.Text()] = id
		id++
	}
}
