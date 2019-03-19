package main

import (
	"fmt"
	"os"
	"strings"
)

var address = "localhost"

// TODO read server configuration from text file
func readServerConfig() {

}

func main() {
	if len(os.Args) > 1 {
		switch strings.ToLower(os.Args[1]) {
		case "server":
			startServer()
		case "client":
			startClientUI()
		}
	} else {
		fmt.Printf("Incorrect number of arguments.\nFirst argument should be either 'client' or 'server'\n")
	}
}
