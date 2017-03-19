package main

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

func messageHandler(s *discordgo.Session, m *discordgo.MessageCreate) {

	var err error

	c, err := s.Channel(m.ChannelID)
	if err != nil {
		return
	} else if c.IsPrivate {
		dmLog.Printf("--> [%s] %s#%s: %s \n", m.Timestamp, m.Author.Username, m.Author.Discriminator, m.Content)
		//fmt.Printf("%s#%s: %s \n", m.Author.Username, m.Author.Discriminator, m.Content)
	}
	if m.Content == "" || m.Content[0] != ',' {
		return
	} else if m.Author.ID == dUser.ID || m.Author.Bot {
		return
	}

	if sqlBlacklistGET(fmt.Sprintf("%s#%s", m.Author.Username, m.Author.Discriminator)) == true {
		return
	}

	msgInfo := &inputInfo{user: m.Author, channel: c}

	msgInfo.dat = inputText(m.Content)
	sndmsg, err := inputParser(msgInfo)
	if err != nil {
		sndmsg := err.Error()
		channel, err := s.UserChannelCreate(msgInfo.user.ID)
		if err != nil {
			errLog.Println("could not create PM", err)
			// Fail silently.
			return
		}

		msg, err := s.ChannelMessageSend(channel.ID, sndmsg)
		if err != nil {
			return
		}
		dmLog.Printf("<-- [%s] %s#%s: %s \n", msg.Timestamp, msg.Author.Username, msg.Author.Discriminator, msg.Content)
		//fmt.Printf("%s#%s: %s \n", msg.Author.Username, msg.Author.Discriminator, msg.Content)
	} else {
		if sndmsg != "" {
			_, _ = s.ChannelMessageSend(m.ChannelID, sndmsg)
		}
	}

}
