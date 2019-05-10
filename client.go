package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha512"
	"crypto/tls"
	"encoding/json"
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
		sendFormData("login", ui)
	})
	ui.Bind("signup", func() {
		sendFormData("signup", ui)
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
func connect(command string, name string, email string, password string) {

	// Client that accepts self-signed certificates (for testing only)
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	// SHA512 hash of password
	clientKey := sha512.Sum512([]byte(password))
	loginKey := clientKey[:32]  // first half for login (256 bits)
	dataKey := clientKey[32:64] // second half for data (256 bits)

	// Generate public/private key pair for server
	clientKP, err := rsa.GenerateKey(rand.Reader, 1024)
	CheckError(err)
	clientKP.Precompute()

	// Format key pair as JSON
	JSONkp, err := json.Marshal(&clientKP)
	CheckError(err)

	// Format public key as JSON
	pubKey := clientKP.Public()
	JSONPub, err := json.Marshal(&pubKey)
	CheckError(err)

	// Prepare data to be sent to server
	data := url.Values{}
	data.Set("command", command)
	data.Set("name", name)
	data.Set("email", email)
	data.Set("password", Encode64(loginKey))
	data.Set("pubKey", Encode64(Compress(JSONPub)))
	data.Set("privKey", Encode64(Encrypt(Compress(JSONkp), dataKey)))

	// Send data via POST
	r, err := client.PostForm("https://localhost:10443", data)
	io.Copy(os.Stdout, r.Body)
	// response := Response{}
	// json.NewDecoder(r.Body).Decode(response)

	CheckError(err)
}

func sendFormData(command string, ui lorca.UI) {
	email := ui.Eval("getEmail()").String()
	password := ui.Eval("getPassword()").String()
	name := ui.Eval("getUsername()").String()
	connect(command, name, email, password)
}
