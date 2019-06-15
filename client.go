package main

import (
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

	"github.com/sethvargo/go-password/password"

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
	Note     SecureNote
}

var currentUser User

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

	// Prepare data to be sent to server
	data := url.Values{}
	data.Set("command", command)
	data.Set("name", fd.Name)
	data.Set("email", fd.Email)
	data.Set("uuid", currentUser.UUID.String())
	data.Set("dataKey", Encode64(dataKey))

	if command == LOGIN || command == SIGNUP {
		data.Set("loginKey", Encode64(loginKey))
	} else if command == ADDPASS {
		data.Set("url", fd.URL)
		data.Set("password", fd.Password)
	} else if command == ADDNOTE {
		data.Set("notetitle", fd.Note.Title)
		data.Set("notecontent", fd.Note.Content)
	}
	// Send data via POST
	r, err := client.PostForm("https://localhost:10443", data)
	CheckError(err)
	defer r.Body.Close()

	// Get response
	response := new(Response)
	json.NewDecoder(r.Body).Decode(response)
	currentUser.UUID = response.UserData.UUID
	return response
}

func readFormData(ui lorca.UI) (data FormData) {
	data.Email = ui.Eval("getEmail()").String()
	data.Password = ui.Eval("getPassword()").String()
	data.Name = ui.Eval("getUsername()").String()
	return
}

func readProfileForm(ui lorca.UI) (data FormData) {
	data.URL = ui.Eval("getUrl()").String()
	useUppercase := true //ui.Eval("getUseUpperCase()").Bool()
	useRepeating := true //ui.Eval("getUseRepeating()").Bool()

	pass, err := password.Generate(12, 4, 4, !useUppercase, useRepeating)
	CheckError(err)

	data.Password = pass
	return
}

func setUpLoginFunctions(ui lorca.UI) {
	ui.Bind("login", func() {
		data := readFormData(ui)
		resp := connect("login", data)
		currentUser.UUID = resp.UserData.UUID
		if resp.Ok {
			loadProfile(ui, resp.UserData)
		}
	})
	ui.Bind("signup", func() {
		data := readFormData(ui)
		resp := connect("signup", data)
		currentUser.UUID = resp.UserData.UUID
		if resp.Ok {
			loadProfile(ui, resp.UserData)
		}
	})
}

func setupProfileFunctions(ui lorca.UI, user User) {
	ui.Bind("addPassword", func() {
		//LogTrace("Pressed")
		data := readProfileForm(ui)
		data.userUUID = currentUser.UUID
		connect("add", data)
		loadProfile(ui, user)
	})
}

func loadProfile(ui lorca.UI, user User) {
	setupProfileFunctions(ui, user)
	html, _ := ioutil.ReadFile("public/profile.html")
	ui.Load("data:text/html," + url.PathEscape(string(html)))
	replaceInDoc(ui, "username", user.Name)
	/*
		i := 1
		for k, v := range user.Passwords {
			replaceInDoc(ui, "url"+string(i), k)
			replaceInDoc(ui, "password"+string(i), v)
			i++
		}
	*/
}

func replaceInDoc(ui lorca.UI, original string, new string) {
	ui.Eval("document.body.innerHTML = document.body.innerHTML.replace('{" + original + "}', '" + new + "');")
}
