package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/url"
	"os"
	"os/signal"
	"runtime"

	"github.com/zserge/lorca"
)

func startClient(width int, height int) {
	args := []string{}
	if runtime.GOOS == "linux" {
		args = append(args, "--class=Lorca")
	}
	ui, err := lorca.New("", "", width, height, args...)
	if err != nil {
		log.Fatal(err)
	}
	defer ui.Close()

	ui.Bind("login", func() {
		login(ui)
	})

	ui.Bind("signup", func() {
		signup(ui)
	})

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
func connect(email string, password string) {
	connection, err := net.Dial("tcp", address+":1337")
	checkError(err)
	defer connection.Close()

	fmt.Print("Connected to ", connection.RemoteAddr(), "\n")

	fmt.Printf("Client is running...\n")

	/*
		keyscan := bufio.NewScanner(os.Stdin)
		netscan := bufio.NewScanner(connection)
		for keyscan.Scan() {
			fmt.Fprintln(connection, keyscan.Text()) // Send input to server
			netscan.Scan()                           // Scan connection
			fmt.Printf("Server: " + netscan.Text())  // Show server messag
		}
	*/
}

func login(ui lorca.UI) {
	email := ui.Eval("getEmail()").String()
	password := ui.Eval("getPassword()").String()
	fmt.Print("Logged in\n")
	fmt.Printf("Email: %s", email)
	fmt.Printf("\nPassword: %s", password)
}

func signup(ui lorca.UI) {
	email := ui.Eval("getEmail()").String()
	password := ui.Eval("getPassword()").String()
	fmt.Print("Signed up\n")
	fmt.Printf("Email: %s", email)
	fmt.Printf("\nPassword: %s", password)
}
