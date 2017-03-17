package main

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

func messageHandler(s *discordgo.Session, m *discordgo.MessageCreate) {

	if m.Content == "" || m.Content[0] != ',' {
		return
	} else if m.Author.ID == dUser.ID || m.Author.Bot {
		return
	}

	if sqlBlacklistGET(fmt.Sprintf("%s#%s", m.Author.Username, m.Author.Discriminator)) == true {
		return
	}

	var info *inputInfo
	msgInfo := inputInfo{user: m.Author, channelID: m.ChannelID}
	info = &msgInfo

	info.dat = inputText(m.Content)
	sndmsg := inputParser(info)

	if sndmsg != "" {
		_, _ = s.ChannelMessageSend(m.ChannelID, sndmsg)
	}

}
