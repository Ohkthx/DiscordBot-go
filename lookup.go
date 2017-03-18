package main

import (
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func userFind(channel, user string) *discordgo.User {

	pos := "" // Position in searching (user ID)
	var userFull []string
	var username string

	userFull = strings.Split(user, "#")
	username = strings.ToLower(userFull[0])

	if len(userFull) != 2 {
		return nil
	}

	chanStruct, err := dSession.Channel(channel)
	if err != nil {
		errLog.Println("Could not get Channel Structure:", err)
		return nil
	}

	for {
		members, err := dSession.GuildMembers(chanStruct.GuildID, pos, 100)
		if err != nil {
			discordLog.Println("Could not find user:", err)
			return nil
		}
		for _, m := range members {
			if strings.ToLower(m.User.Username) == username && m.User.Discriminator == userFull[1] {
				return m.User
				//} else if m.User.Username == user {
				//	return userString(m)
			}
		}

		pos = members[len(members)-1].User.ID
	}

}

func channelFind(name string) *discordgo.Channel {
	guilds, err := dSession.UserGuilds()
	if err != nil {
		discordLog.Println(err)
		log.Println("Error getting guilds", err)
		return nil
	}

	for _, g := range guilds {
		channels, err := dSession.GuildChannels(g.ID)
		if err != nil {
			discordLog.Println(err)
			return nil
		}
		for _, c := range channels {
			if strings.ToLower(c.Name) == strings.ToLower(name) {
				return c
			}
		}
	}

	return nil
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
