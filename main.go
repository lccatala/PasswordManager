package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

var address = "localhost"

// Error handling
func checkError(e error) {
	if e != nil {
		panic(e)
	}
}

// TODO read server configuration from text file
func readServerConfig() {

}

// Server logic
func server() {
	readServerConfig()

	// Listen on port 1337 with tcp protocol
	line, err := net.Listen("tcp", "localhost:1337")
	checkError(err)
	defer line.Close()
	fmt.Printf("Server is running...\n")
	for {

		// Accept incomming connections
		connection, err := line.Accept()
		checkError(err)
		connection.Close()

		go func() { // Concurrent lambda function to handle incomming connections
			// Get remote port
			_, port, err := net.SplitHostPort(connection.RemoteAddr().String())
			checkError(err)

			fmt.Print("Connection: ", connection.LocalAddr(), "<-->", connection.RemoteAddr())
			scanner := bufio.NewScanner(connection)

			// Scan
			for scanner.Scan() {
				fmt.Print("client[", port, "]:", scanner.Text()) // Show client message
				fmt.Fprint(connection, "ack: ", scanner.Text())  // Send ACK to client
			}

			connection.Close()
			fmt.Print("close[", port, "]")
		}()
	}
}

// Client logic
func client() {
	fmt.Printf("Client is running...\n")

	// Establish connection with server
	connection, err := net.Dial("tcp", address+":1337")
	checkError(err)
	defer connection.Close()

	fmt.Print("Connected to ", connection.RemoteAddr())

	keyscan := bufio.NewScanner(os.Stdin)
	netscan := bufio.NewScanner(connection)

	// Input scan
	for keyscan.Scan() {
		fmt.Fprintln(connection, keyscan.Text()) // Send input to server
		netscan.Scan()                           // Scan connection
		fmt.Printf("Server: " + netscan.Text())  // Show server messag
	}
}

func main() {
	if len(os.Args) > 1 {
		switch strings.ToLower(os.Args[1]) {
		case "server":
			server()
		case "client":
			client()
		}
	} else {
		fmt.Printf("Incorrect number of arguments.\nFirst argument should be either 'client' or 'server'\n")
	}
}
