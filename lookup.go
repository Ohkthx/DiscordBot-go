package main

import (
	"errors"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func userFind(info *inputInfo, user string) (*discordgo.User, error) {
	pos := "" // Position in searching (user ID)
	session := info.session
	var userFull []string
	var username string
	var err error

	userFull = strings.Split(user, "#")
	username = strings.ToLower(userFull[0])

	if len(userFull) != 2 {
		err = errors.New("bad user supplied. User format: username#1234 ")
		return nil, err
	}

	/*
		chanStruct, err := session.Channel(info.channel.ID)
		if err != nil {
			errLog.Println("could not get channel structure:", err)
			err = errors.New("could not obtain channels")
			return nil, err
		}
	*/
	for {
		members, err := session.GuildMembers(info.channel.GuildID, pos, 100)
		if err != nil {
			err = errors.New("could not obtain user list")
			return nil, err
		}
		for _, m := range members {
			if strings.ToLower(m.User.Username) == username && m.User.Discriminator == userFull[1] {
				return m.User, nil
				//} else if m.User.Username == user {
				//	return userString(m)
			}
		}

		pos = members[len(members)-1].User.ID
	}
}

func channelFind(info *inputInfo, name string) (*discordgo.Channel, error) {
	session := info.session
	guilds, err := session.UserGuilds()
	if err != nil {
		errLog.Println("error getting guilds", err)
		err = errors.New("could not get server info")
		return nil, err
	}

	for _, g := range guilds {
		channels, err := session.GuildChannels(g.ID)
		if err != nil {
			errLog.Println("error getting guild channels", err)
			err = errors.New("error getting rooms")
			return nil, err
		}
		for _, c := range channels {
			if strings.ToLower(c.Name) == strings.ToLower(name) {
				return c, nil
			}
		}
	}

	err = errors.New("no channels found")
	return nil, err
}

func userString(dat *discordgo.Member) string {
	var str string

	str = "```"
	str += "\nUsername: " + dat.User.Username
	str += "\nID: " + dat.User.ID
	str += "\nAvatar: " + discordgo.EndpointUserAvatar(dat.User.ID, dat.User.Avatar)
	str += "\nJoined: " + dat.JoinedAt
	str += "\n```"

	return str
}
