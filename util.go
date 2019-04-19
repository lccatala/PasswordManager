package main

import (
	"bufio"
	"bytes"
	"compress/zlib"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

// CheckError is a standard error checking function
func CheckError(e error) {
	if e != nil {
		panic(e)
	}
}

// Encrypt encrypts with AES in CTR mode, adding an IV at the beggining
func Encrypt(data, key []byte) (out []byte) {
	out = make([]byte, len(data)+16) // Allocate space for IV + data
	rand.Read(out[:16])              // generate IV
	blk, err := aes.NewCipher(key)   // AES block cipher, requires a key
	CheckError(err)

	ctr := cipher.NewCTR(blk, out[:16]) // Flow (stream?) cipher in CTR mode, requires IV
	ctr.XORKeyStream(out[16:], data)    // Encrypt the data
	return
}

// Compress compresses lol
func Compress(data []byte) []byte {
	var b bytes.Buffer

	// Use a writer to compress over b
	w := zlib.NewWriter(&b)
	w.Write(data)
	w.Close()

	return b.Bytes()
}

// Decompress decompresses lol
func Decompress(data []byte) []byte {
	var b bytes.Buffer

	// The reader decompresses while reading
	r, err := zlib.NewReader(bytes.NewReader(data))
	CheckError(err)

	// Copy from decompressor to buffer
	io.Copy(&b, r)
	r.Close()

	return b.Bytes()
}

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

	// Add user to users list
	f, err := os.OpenFile("users/users.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	CheckError(err)
	defer f.Close()

	data := Encrypt([]byte(user.Name), KEY)
	_, err = f.Write(data)
	CheckError(err)
	_, err = f.Write([]byte("\n"))
	CheckError(err)
}

// ReadUser reads a user struct from a json file
func ReadUser(username string) User {
	user := User{}
	fileData, _ := ioutil.ReadFile("users/" + username + ".json")
	json.Unmarshal([]byte(fileData), &user)
	return user
}

// ReadAllUsers reads all users contained in users/users.txt to the users map
func ReadAllUsers(k []byte) (users map[string]int) {
	users = make(map[string]int)
	file, err := os.Open("users/users.txt")
	CheckError(err)
	defer file.Close()

	scanner := bufio.NewScanner(file)

	id := 1
	for scanner.Scan() {
		users[scanner.Text()] = id
		id++
	}
	return
}

func WriteAllUsers() {
	f, err := os.OpenFile("users/tempusers.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	CheckError(err)
	defer f.Close()

	for k := range users {
		LogTrace("Writing user " + k + " as " + string(Encrypt([]byte(k), KEY)))
		data := Encrypt([]byte(k), KEY)
		_, err = f.Write(data)
		CheckError(err)
		_, err = f.Write([]byte("\n"))
		CheckError(err)
	}
}

const (
	infoColor    = "\033[1;34m%s\033[0m"
	warningColor = "\033[1;33m%s\033[0m"
	errorColor   = "\033[1;31m%s\033[0m"
	traceColor   = "\033[0;36m%s\033[0m"
)

func LogError(message string) {
	fmt.Printf(errorColor, "[ERROR]: "+message)
	fmt.Println()
}

func LogWarning(message string) {
	fmt.Printf(warningColor, "[WARN]: "+message)
	fmt.Println()
}

func LogInfo(message string) {
	fmt.Printf(infoColor, "[INFO]: "+message)
	fmt.Println()
}

func LogTrace(message string) {
	fmt.Printf(traceColor, "[TRACE]: "+message)
	fmt.Println()
}
