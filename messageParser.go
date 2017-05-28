package main

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/d0x1p2/DiscordBot-go/bot"
)

func messageHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	var err error

	u, _ := s.User("@me")
	c, err := s.Channel(m.ChannelID)
	if err != nil {
		return
	} else if c.IsPrivate && m.Author.ID != u.ID {
		dmLog.Printf("--> %s#%s: %s \n", m.Author.Username, m.Author.Discriminator, m.Content)
	}
	if m.Content == "" || m.Content[0] != ',' {
		return
	} else if m.Author.ID == u.ID || m.Author.Bot {
		return
	}

	state := bot.New(db, s)
	if err != nil {
		return
	}
	state.User = m.Author
	state.Channel = c
	state.Message = m
	state.Cmd = inputText(m.Content)
	err = state.SetChannels(3)
	if err != nil {
		errLog.Println(err)
	}

	if state.BlacklistWrapper(fmt.Sprintf("%s#%s", m.Author.Username, m.Author.Discriminator)) == true {
		return
	}

	res := inputParser(state)
	if res.Err != nil {
		errLog.Println(res.Err)
		sndmsg := res.Errmsg
		channel, err := s.UserChannelCreate(state.User.ID)
		if err != nil {
			errLog.Println("could not create PM", err)
			// Fail silently.
			return
		}

		msg, err := s.ChannelMessageSend(channel.ID, sndmsg)
		if err != nil {
			return
		}
		dmLog.Printf("<-- %s#%s [%s#%s]: %s \n", msg.Author.Username, msg.Author.Discriminator, m.Author.Username, m.Author.Discriminator, msg.Content)
		//fmt.Printf("%s#%s: %s \n", msg.Author.Username, msg.Author.Discriminator, msg.Content)
	} else {
		if res.Sndmsg != "" {
			_, _ = s.ChannelMessageSend(m.ChannelID, res.Sndmsg)
		}
	}

}
