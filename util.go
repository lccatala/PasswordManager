// Utility functions for encryption, compression and error checking
package main

import (
	"bytes"
	"compress/zlib"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"io"
	"io/ioutil"
)

// User data type
type User struct {
	Email string
	Name  string
	Hash  []byte            // Password hash
	Salt  []byte            // Password salt
	Data  map[string]string // Additional data
}

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

// Error handling
func checkError(e error) {
	if e != nil {
		panic(e)
	}
}

// Encrypt with AES, adding IV at the beggining
func encrypt(data, key []byte) (out []byte) {
	out = make([]byte, len(data)+16) // Allocate space for IV + data
	rand.Read(out[:16])              // generate IV
	blk, err := aes.NewCipher(key)   // AES block cipher, requires a key
	checkError(err)

	ctr := cipher.NewCTR(blk, out[:16]) // Flow (stream?) cipher in CTR mode, requires IV
	ctr.XORKeyStream(out[16:], data)    // Encrypt the data
	return
}

// Compress
func compress(data []byte) []byte {
	var b bytes.Buffer

	// Use a writer to compress over b
	w := zlib.NewWriter(&b)
	w.Write(data)
	w.Close()

	return b.Bytes()
}

// Decompress
func decompress(data []byte) []byte {
	var b bytes.Buffer

	// The reader decompresses while reading
	r, err := zlib.NewReader(bytes.NewReader(data))
	checkError(err)

	// Copy from decompressor to buffer
	io.Copy(&b, r)
	r.Close()

	return b.Bytes()
}

// []byte (base64) to string
func encode64(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

// String to []byte (base64)
func decode64(s string) []byte {
	b, err := base64.StdEncoding.DecodeString(s)
	checkError(err)
	return b
}

// Write user struct to json file
func writeUser(user User) {
	// TODO encrypt user data before saving it to file
	fileData, err := json.MarshalIndent(user, "", "  ")
	checkError(err)
	err = ioutil.WriteFile("users/"+user.Name+".json", fileData, 0644)
	checkError(err)
}

// Read user struct from json file
func readUser(username string) (User, bool) {
	user := User{}
	fileData, err := ioutil.ReadFile(username + ".json")
	json.Unmarshal([]byte(fileData), &user)
	return user, err != nil
}
