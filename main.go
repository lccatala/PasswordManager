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
	line, err := net.Listen("tcp", address+":1337")
	checkError(err)
	defer line.Close()
	fmt.Printf("Server is running...\n")
	for {

		// Accept incomming connections
		connection, err := line.Accept()
		checkError(err)

		go func() { // Concurrent lambda function to handle incomming connections
			// Get remote port
			_, port, err := net.SplitHostPort(connection.RemoteAddr().String())
			checkError(err)

			fmt.Print("Connection: ", connection.LocalAddr(), "<-->", connection.RemoteAddr(), "\n")

			scanner := bufio.NewScanner(connection)

			// Scan
			for scanner.Scan() {
				fmt.Println("client[", port, "]:", scanner.Text()) // Show client message
				fmt.Fprintln(connection, "ack: ", scanner.Text())  // Send ACK to client
			}

			connection.Close()
			fmt.Print("closed[", port, "]")
		}()

	}
}

func main() {
	if len(os.Args) > 1 {
		switch strings.ToLower(os.Args[1]) {
		case "server":
			server()
		case "client":
			startClient(300, 500)
		}
	} else {
		fmt.Printf("Incorrect number of arguments.\nFirst argument should be either 'client' or 'server'\n")
	}
}
