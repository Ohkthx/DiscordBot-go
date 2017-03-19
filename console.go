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
	var err error

	switch input.command {
	case "user":
		fallthrough
	case "admin":
		if channel == nil {
			log.Println("Set channel first.")
			break
		}
		user, err = userFind(channel.ID, input.args[0])
		if err != nil {
			log.Println(err)
			break
		}
		info.admin = true
		info.user = user

	case "channel":
		channel, err = channelFind(input.args[0])
		if err != nil {
			log.Println(err)
			break
		}
		info.channel = channel

	case "msg":
		if channel == nil {
			log.Println("Set channel first.")
			break
		}
		_, _ = dSession.ChannelMessageSend(channel.ID, strings.Join(input.args, " "))
	case "privmsg":
		if input.length < 2 || input.text == "" {
			log.Println("Bad privmsg")
			break
		}
		recep, err := userFind(channel.ID, input.args[0])
		if err != nil {
			log.Println(err)
			break
		}
		c, _ := dSession.UserChannelCreate(recep.ID)
		_, _ = dSession.ChannelMessageSend(c.ID, strings.Join(input.args[1:], " "))
	case "pmc":
		c, _ := dSession.UserChannels()
		for k, p := range c {
			fmt.Printf("%d) %s\n", k, p.Name)
		}

	default:
		if info.admin == true {
			text, _ := inputParser(info)
			if info.send {

				_, _ = dSession.ChannelMessageSend(info.channel.ID, text)
			} else {
				log.Println(text)
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
