package main

import (
	"io/ioutil"
	"log"
	"net/url"
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

	// A simple way to know when UI is ready (uses body.onload event in JS)
	ui.Bind("start", func() {
		log.Println("UI is ready")
	})

	// Load HTML after Go functions are bound to JS
	html, _ := ioutil.ReadFile("public/login.html")
	ui.Load("data:text/html," + url.PathEscape(string(html)))
	<-ui.Done()
	/*

		fmt.Printf("Client is running...\n")

		// Establish connection with server
		connection, err := net.Dial("tcp", address+":1337")
		checkError(err)
		defer connection.Close()

		fmt.Print("Connected to ", connection.RemoteAddr(), "\n")

		keyscan := bufio.NewScanner(os.Stdin)
		netscan := bufio.NewScanner(connection)

		// Input scan
		for keyscan.Scan() {
			fmt.Fprintln(connection, keyscan.Text()) // Send input to server
			netscan.Scan()                           // Scan connection
			fmt.Printf("Server: " + netscan.Text())  // Show server messag
		}

		// Data model: number of ticks
		ticks := uint32(0)
		// Channel to connect UI events with the background ticking goroutine
		togglec := make(chan bool)
		// Bind Go functions to JS
		ui.Bind("toggle", func() { togglec <- true })
		ui.Bind("reset", func() {
			atomic.StoreUint32(&ticks, 0)
			ui.Eval(`document.querySelector('.timer').innerText = '0'`)
		})



		// Start ticker goroutine
		go func() {
			t := time.NewTicker(100 * time.Millisecond)
			for {
				select {
				case <-t.C: // Every 100ms increate number of ticks and update UI
					ui.Eval(fmt.Sprintf(`document.querySelector('.timer').innerText = 0.1*%d`,
						atomic.AddUint32(&ticks, 1)))
				case <-togglec: // If paused - wait for another toggle event to unpause
					<-togglec
				}
			}
		}()
		<-ui.Done()
	*/
}
