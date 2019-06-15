package main

import (
	"os"
	"strings"
)

func main() {
	if len(os.Args) > 1 {
		switch strings.ToLower(os.Args[1]) {
		case "server":
			startServer()
		case "client":
			startClientUI()
		}
	} else {
		LogError("Incorrect number of arguments. Possible executions are: 'go run *.client' or 'go run *.server'")
	}
}
