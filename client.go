package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha512"
	"crypto/tls"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"runtime"

	"github.com/google/uuid"
	"github.com/zserge/lorca"
)

// FormData stores the data collected from login/signup/add-password forms in the client
type FormData struct {
	Email    string
	Password string
	Name     string
	URL      string
	userUUID uuid.UUID
}

var currentUUID uuid.UUID

func startClientUI() {
	args := []string{}
	if runtime.GOOS == "linux" {
		args = append(args, "--class=Lorca")
	}
	ui, err := lorca.New("", "", 300, 500, args...)
	defer ui.Close()
	if err != nil {
		log.Fatal(err)
	}

	setUpLoginFunctions(ui)

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
func connect(command string, fd FormData) *Response {

	// Client that accepts self-signed certificates (for testing only)
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	// SHA512 hash of password
	clientKey := sha512.Sum512([]byte(fd.Password))
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
	data.Set("name", fd.Name)
	data.Set("email", fd.Email)
	data.Set("password", Encode64(loginKey))
	data.Set("pubKey", Encode64(Compress(JSONPub)))
	data.Set("privKey", Encode64(Encrypt(Compress(JSONkp), dataKey)))
	data.Set("url", fd.URL)
	data.Set("uuid", currentUUID.String())

	// Send data via POST
	r, err := client.PostForm("https://localhost:10443", data)
	CheckError(err)
	defer r.Body.Close()

	// Get response
	response := new(Response)
	json.NewDecoder(r.Body).Decode(response)
	return response
}

func readFormData(ui lorca.UI) (data FormData) {
	data.Email = ui.Eval("getEmail()").String()
	data.Password = ui.Eval("getPassword()").String()
	data.Name = ui.Eval("getUsername()").String()
	return
}

func readPasswordURL(ui lorca.UI) (data FormData) {
	data.URL = ui.Eval("getUrl()").String()
	return
}

func setUpLoginFunctions(ui lorca.UI) {
	ui.Bind("login", func() {
		data := readFormData(ui)
		resp := connect("login", data)
		LogTrace("Logged in as " + resp.UserData.Name)
		loadProfile(ui, resp.UserData)
	})
	ui.Bind("signup", func() {
		data := readFormData(ui)
		resp := connect("signup", data)
		loadProfile(ui, resp.UserData)
		LogTrace("Signed up as " + resp.UserData.Name)
	})
}

func setupProfileFunctions(ui lorca.UI) {
	ui.Bind("addPassword", func() {
		data := readPasswordURL(ui)
		resp := connect("add", data)
		loadProfile(ui, resp.UserData)
	})
}

func loadProfile(ui lorca.UI, user User) {
	setupProfileFunctions(ui)
	html, _ := ioutil.ReadFile("public/profile.html")
	ui.Load("data:text/html," + url.PathEscape(string(html)))
	replaceInDoc(ui, "username", user.Name)
	for k, v := range user.Passwords {
		ui.Eval("document.write('<p>Password for " + k + ": " + v + "</p>')")
	}
}

func replaceInDoc(ui lorca.UI, original string, new string) {
	ui.Eval("document.body.innerHTML = document.body.innerHTML.replace('{" + original + "}', '" + new + "');")
}
