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
	coreInfo := &inputInfo{admin: false, user: nil, channel: nil}
	coreInfo.send = false

	time.Sleep(150 * time.Millisecond)
	reader := bufio.NewReader(os.Stdin)

	for _running {
		fmt.Printf("[%s] > ", time.Now().Format(time.Stamp))
		input, _ := reader.ReadString('\n')
		temp := strings.Fields(input)
		coreInfo.dat = inputText(strings.Join(temp, " "))
		if coreInfo.dat.command != "" {
			ioHandler(coreInfo)
		}
	}
}

func ioHandler(info *inputInfo) {
	// Used to parse user input to console/cli
	input := info.dat
	user := info.user
	channel := info.channel

	switch input.command {
	case "user":
		fallthrough
	case "admin":
		if channel == nil {
			log.Println("Set channel first.")
			break
		}
		user = userFind(channel.ID, input.args[0])
		if user == nil {
			log.Println("Bad user  => ", input.args[0])
			break
		}
		info.admin = true
		info.user = user

	case "channel":
		channel = channelFind(input.args[0])
		if channel == nil {
			log.Println("Bad channel  => ", input.args[0])
			break
		}
		info.channel = channel

	case "msg":
		if channel == nil {
			log.Println("Set channel first.")
			break
		}
		_, _ = dSession.ChannelMessageSend(channel.ID, strings.Join(input.args, " "))

	default:
		if info.admin == true {
			if info.send {
				_, _ = dSession.ChannelMessageSend(info.channel.ID, inputParser(info))
			} else {
				log.Println(inputParser(info))
			}
			break
		} else if channel != nil && user != nil {
			break
		}
		log.Println("Set channel and/or admin")
		fallthrough

	case "status":
		if channel == nil {
			log.Printf("Status: %10s [%s]\n", "Channel", "----")
			log.Printf("Status: %10s [%s]\n", "Admin", "----")
			break
		}
		if user == nil {
			log.Printf("Status: %10s [%s]\n", "Channel", channel.Name)
			log.Printf("Status: %10s [%s]\n", "Admin", "----")
			break
		}
		log.Printf("Status: %10s [%s]\n", "Channel", channel.Name)
		log.Printf("Status: %10s [%s]\n", "Admin", user.Username)
		break

	case "send":
		if info.send == true {
			info.send = false
		} else {
			info.send = true
		}
		log.Println("Send to channel: ", info.send)

	case "quit":
		fallthrough
	case "exit":
		cleanup()
		os.Exit(0)
	}
}
