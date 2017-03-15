package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

func core() {
	// Main loop for processing user input to console.
	_running := true
	time.Sleep(150 * time.Millisecond)
	reader := bufio.NewReader(os.Stdin)

	for _running {
		fmt.Printf("[%s] > ", time.Now().Format(time.Stamp))
		input, _ := reader.ReadString('\n')
		if len(input) > 1 {
			ret := ioHandler(strings.Fields(input))
			if ret == "" {
				fmt.Printf("")
			}
		}

	}
	os.Exit(1)
}

func ioHandler(input []string) string {
	// Used to parse user input to console/cli
	switch input[0] {
	case "log":
	case "quit":
		fallthrough
	case "exit":
		cleanup()
	default:
		log.Println("No options found. Try again.")
		break

	}

	return ""

}
