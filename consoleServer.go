package main

import (
	"bufio"
	"net"
	"strings"
)

func serverInit() error {
	listener, err := net.Listen("tcp", "localhost:8411")
	if err != nil {
		errLog.Println("could not start lsitener", err)
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
	coreInfo := &inputInfo{admin: false, user: nil, channel: nil}
	coreInfo.send = false

	for {
		input, _ := reader.ReadString('\n')
		temp := strings.Fields(input)
		coreInfo.dat = inputText(strings.Join(temp, " "))
		if input == "quit" || input == "exit" {
			break
		}
		if coreInfo.dat.command != "" {
			err := ioHandler(coreInfo)
			if err != nil {
				conn.Write([]byte(err.Error() + "\n"))
			}
		}
	}
	conn.Close()
}
