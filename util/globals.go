package util

import (
	"bytes"
	"compress/zlib"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"
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

// Users stores all signed up usernames and user ID's
var Users map[string]int

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
