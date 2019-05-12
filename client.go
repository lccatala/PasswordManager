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

	"github.com/zserge/lorca"
)

// FormData stores the data collected from login/signup forms in the client
type FormData struct {
	Email    string
	Password string
	Name     string
}

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

	setUpFunctions(ui)

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

func setUpFunctions(ui lorca.UI) {
	ui.Bind("login", func() {
		data := readFormData(ui)
		resp := connect("login", data)
		LogTrace("Logged in as " + resp.UserData.Name)
		//loadProfile(ui, resp.UserData)
	})
	ui.Bind("signup", func() {
		data := readFormData(ui)
		resp := connect("signup", data)
		loadProfile(ui, resp.UserData)
		LogTrace("Signed up as " + resp.UserData.Name)
	})
}

func loadProfile(ui lorca.UI, user User) {
	html, _ := ioutil.ReadFile("public/profile.html")
	ui.Load("data:text/html," + url.PathEscape(string(html)))
	ui.Eval("document.write('<html>')")
	ui.Eval("document.write('<div class=\"container\">')")
	ui.Eval("document.write('<html><h1>Welcome, " + user.Name + "</h1></html>')")
	ui.Eval("document.write('<p>Password 1: </p>')")
	ui.Eval("document.write('<p>Password 2: </p>')")
	ui.Eval("document.write('<p>Password 2: </p>')")
	ui.Eval("document.write('<button id=\"login-button\" class=\"btn btn-lg btn-primary btn-block\" type=\"submit\">Log in</button>')")
	ui.Eval("document.write('</div>')")
	ui.Eval("document.write('</html>')")
}
