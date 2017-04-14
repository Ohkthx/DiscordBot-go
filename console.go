package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

func core(session *discordgo.Session) {
	// Main loop for processing user input to console.
	coreInfo := &inputInfo{admin: false, user: nil, channel: nil}
	coreInfo.send = false
	coreInfo.session = session

	time.Sleep(150 * time.Millisecond)
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf("[%s] > ", time.Now().Format(time.Stamp))
		input, _ := reader.ReadString('\n')
		temp := strings.Fields(input)
		coreInfo.dat = inputText(strings.Join(temp, " "))
		if coreInfo.dat.command != "" {
			err := ioHandler(coreInfo)
			if err != nil {
				fmt.Println(err)
			}
		}
	}
}

func ioHandler(info *inputInfo) error {
	// Used to parse user input to console/cli
	input := info.dat
	user := info.user
	channel := info.channel
	session := info.session
	var err error

	switch input.command {
	case "ch-base":
		c, err := dbUpdate(info, false)
		if err != nil {
			return err
		}
		log.Printf("Messages found: %d\n", c)
	case "ch-update":
		c, err := dbUpdate(info, true)
		if err != nil {
			return err
		}
		log.Printf("Messages found: %d\n", c)
	case "user":
		fallthrough
	case "admin":
		if channel == nil {
			err = errors.New("set channel first")
			break
		}
		user, err = userFind(info, input.args[0])
		if err != nil {
			break
		}
		info.admin = true
		info.user = user

	case "channel":
		channel, err = channelFind(info, input.args[0])
		if err != nil {
			break
		}
		info.channel = channel

	case "msg":
		if channel == nil {
			err = errors.New("set channel first")
			break
		}
		_, _ = session.ChannelMessageSend(channel.ID, strings.Join(input.args, " "))
	case "privmsg":
		if input.length < 2 || input.text == "" {
			err = errors.New("bad privmsg")
			break
		}
		recep, err := userFind(info, input.args[0])
		if err != nil {
			break
		}
		c, _ := session.UserChannelCreate(recep.ID)
		_, _ = session.ChannelMessageSend(c.ID, strings.Join(input.args[1:], " "))

	default:
		if info.admin == true {
			text, _ := inputParser(info)
			if info.send {
				_, _ = session.ChannelMessageSend(info.channel.ID, text)
			} else {
				err = errors.New(text)
			}
			break
		} else if channel != nil && user != nil {
			break
		}
		err = errors.New("Set channel and/or admin")

	case "send":
		if info.send == true {
			info.send = false
		} else {
			info.send = true
		}

		tmp := fmt.Sprintf("send to channel: %s", strconv.FormatBool(info.send))
		err = errors.New(tmp)

	// kill-server will be used by remote consoles
	case "kill-server":
		fallthrough
	case "quit":
		fallthrough
	case "exit":
		info.session.Close()
		cleanup()
		os.Exit(0)
	}
	return err
}
