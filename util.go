// Utility functions for encryption, compression and error checking
package main

import (
	"bytes"
	"compress/zlib"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"io"
)

// Login and register modes (login and sign up)
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
