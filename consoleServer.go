package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/d0x1p2/DiscordBot-go/bot"
)

func serverInit() error {
	listener, err := net.Listen("tcp", "localhost:8411")
	if err != nil {
		errLog.Println("could not start listener", err)
		return err
	}

	// Close the listener when the application closes.
	defer listener.Close()

	for {
		// Wait for connection
		conn, err := listener.Accept()
		if err != nil {
			errLog.Println("[listener] accepting connection", err)
			continue
		}

		go requestHandler(conn)
	}
}

func requestHandler(conn net.Conn) {
	reader := bufio.NewReader(conn)
	state := bot.New(db, nil)
	state.Admin = false
	state.Sendmsg = false

	for {
		fmt.Printf("[%s] > ", time.Now().Format(time.Stamp))
		input, _ := reader.ReadString('\n')
		temp := strings.Fields(input)
		state.Cmd = inputText(strings.Join(temp, " "))
		if input == "quit" || input == "exit" {
			break
		}
		if state.Cmd.Command != "" {
			err := ioHandler(state)
			if err != nil {
				conn.Write([]byte(err.Error() + "\n"))
			}
		}
	}
	conn.Close()
}
