package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha512"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"runtime"

	"github.com/zserge/lorca"
)

func startClientUI() {
	args := []string{}
	if runtime.GOOS == "linux" {
		args = append(args, "--class=Lorca")
	}
	ui, err := lorca.New("", "", 300, 500, args...)
	defer ui.Close()
	ui.Bind("login", func() {
		getUserData("login", ui)
	})
	ui.Bind("signup", func() {
		getUserData("signup", ui)
	})

	if err != nil {
		log.Fatal(err)
	}

	// Load HTML after Go functions are bound to JS
	html, _ := ioutil.ReadFile("public/login.html")
	ui.Load("data:text/html," + url.PathEscape(string(html)))

	// Wait until the interrupt signal arrives or browser window is closed
	sigc := make(chan os.Signal)
	signal.Notify(sigc, os.Interrupt)
	select {
	case <-sigc:
	case <-ui.Done():
	}

}

// Establish connection with server
func connect(mode string, email string, password string) {

	// Client that accepts self-signed certificates (for testing only)
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	// SHA512 hash of password
	clientKey := sha512.Sum512([]byte(password))
	loginKey := clientKey[:32]  // one half for login (256 bits)
	dataKey := clientKey[32:64] // the other for data (256 bits)

	// Generate public/private key pair for server
	clientKP, err := rsa.GenerateKey(rand.Reader, 1024)
	checkError(err)
	clientKP.Precompute() // Speed up it's use with a precomputation

	JSONkp, err := json.Marshal(&clientKP) // codificamos con JSON
	checkError(err)

	pubKey := clientKP.Public()           // extraemos la clave pública por separado
	JSONPub, err := json.Marshal(&pubKey) // y codificamos con JSON
	checkError(err)

	// Prepare data to be sent to server
	data := url.Values{}                     // struct to store the values
	data.Set("command", string(mode))        // command (string)
	data.Set("email", email)                 // email (string)
	data.Set("password", encode64(loginKey)) // password in base64

	// Compress and code the private key
	data.Set("pubKey", encode64(compress(JSONPub)))

	// Compress, cypher and code the private key
	data.Set("privKey", encode64(encrypt(compress(JSONkp), dataKey)))

	response, err := client.PostForm("https://localhost:10443", data) // Send data via POST
	checkError(err)
	io.Copy(os.Stdout, response.Body) // Show response body
	fmt.Println()
}

func getUserData(mode string, ui lorca.UI) {
	email := ui.Eval("getEmail()").String()
	password := ui.Eval("getPassword()").String()
	connect(mode, email, password)
}
