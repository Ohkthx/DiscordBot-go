package main

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func messageHandler(s *discordgo.Session, m *discordgo.MessageCreate) {

	u, err := s.User("@me")
	if err != nil {
		discordLog.Println("Error obtaining account details:", err)
	}

	if m.Author.ID == u.ID || m.Author.Bot {
		return
	}

	//input := strings.Fields(m.Content)
	input := strings.Split(m.Content, " ")
	if input[0][0] != ',' {
		return
	}

	_proc := false
	var text string
	var inputLen int
	for k, p := range input {
		if strings.HasPrefix(p, "\"") {
			text = strings.Join(input[k:], " ")
			text = strings.TrimPrefix(text, "\"")
			if strings.HasSuffix(text, "\"") {
				text = strings.TrimSuffix(text, "\"")
			}
			_proc = true
			inputLen = k
			break
		}
	}

	if _proc == false {
		text = input[len(input)-1]
		inputLen = len(input) - 1
	}

	input[0] = input[0][1:]

	var sndmsg string

	switch input[0] {
	case "event":
		sndmsg = sqlCMDEvent()
	case "grant":
		sndmsg = sqlCMDGrant(m.Author, m.ChannelID, input)
	case "add":
		sndmsg = sqlCMDAdd(m.Author, input, text, inputLen)
	case "version":
		sndmsg = fmt.Sprintf("version: `%s`", _version)
	default:
		sndmsg = sqlCMDSearch(input)
	}

	if sndmsg != "" {
		_, _ = s.ChannelMessageSend(m.ChannelID, sndmsg)
	}

}
