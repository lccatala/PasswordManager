package main

import "fmt"

const (
	infoColor    = "\033[1;34m%s\033[0m"
	warningColor = "\033[1;33m%s\033[0m"
	errorColor   = "\033[1;31m%s\033[0m"
	traceColor   = "\033[0;36m%s\033[0m"
)

func LogError(message string) {
	fmt.Printf(errorColor, "[ERROR]: "+message)
	fmt.Println()
}

func LogWarning(message string) {
	fmt.Printf(warningColor, "[WARN]: "+message)
	fmt.Println()
}

func LogInfo(message string) {
	fmt.Printf(infoColor, "[INFO]: "+message)
	fmt.Println()
}

func LogTrace(message string) {
	fmt.Printf(traceColor, "[TRACE]: "+message)
	fmt.Println()
}
