package main

import (
	"strings"

	"github.com/bwmarrin/discordgo"
)

func userFind(channel, user string, idLookup bool) string {

	pos := "" // Position in searching (user ID)
	var userFull []string

	if idLookup == true {
		userFull = strings.Split(user, "#")
		user = userFull[0]
	}

	chanStruct, err := dSession.Channel(channel)
	if err != nil {
		errLog.Println("Could not get Channel Structure:", err)
		return ""
	}

	for {
		members, err := dSession.GuildMembers(chanStruct.GuildID, pos, 100)
		if err != nil {
			discordLog.Println("Could not find user:", err)
			return ""
		}
		for _, m := range members {
			if m.User.Username == user && m.User.Discriminator == userFull[1] {
				return m.User.ID
			} else if m.User.Username == user {
				return userString(m)
			}
		}

		pos = members[len(members)-1].User.ID
	}

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
