package main

import (
	"fmt"
	"os"
	"strings"
)

// Error handling
func checkError(e error) {
	if e != nil {
		panic(e)
	}
}

func readServerConfig() {

}

func server() {
	readServerConfig()
	fmt.Printf("Server is running...\n")
}

func client() {
	fmt.Printf("Client is running...\n")
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
