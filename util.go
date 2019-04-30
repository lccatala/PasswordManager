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

	ctr := cipher.NewCTR(blk, out[:16]) // Cipher in CTR mode, requires IV
	ctr.XORKeyStream(out[16:], data)    // Encrypt the data
	return
}

// Decrypt decrypts AES-encrypted byte arrays in CTR mode
func Decrypt(data, key []byte) (out []byte) {
	out = make([]byte, len(data)-16)
	blk, err := aes.NewCipher(key)
	CheckError(err)
	ctr := cipher.NewCTR(blk, data[:16])
	ctr.XORKeyStream(out, data[16:])
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

// ReadUser reads a user struct from a json file
func ReadUser(username string) User {
	user := User{}
	fileData, _ := ioutil.ReadFile("users/" + username + ".json")
	json.Unmarshal([]byte(fileData), &user)
	return user
}

// ReadAllUsers reads all users contained in users/users.txt to the users map
func ReadAllUsers(k []byte) (users map[string]bool) {
	users = make(map[string]bool)
	file, err := os.Open("users/users.txt")
	CheckError(err)
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		name := Decrypt([]byte(scanner.Text()), KEY)
		users[string(name)] = true
	}
	return
}
