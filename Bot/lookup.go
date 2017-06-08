package bot

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

// UserFind searches a session's users and if the ID is found, returns the object.
func (state *Instance) UserFind(userbase string) (user *discordgo.User, err error) {
	s := state.Session
	pos := "" // Position in searching (user ID)
	var userFull []string
	var username string

	userFull = strings.Split(userbase, "#")
	username = strings.ToLower(userFull[0])

	if len(userFull) != 2 {
		err = fmt.Errorf("bad user supplied. User format: username#1234")
		return
	}

	for {
		members, err1 := s.GuildMembers(state.Channel.GuildID, pos, 100)
		if err1 != nil {
			err = fmt.Errorf("could not obtain user list: %s", err1.Error)
			return
		}
		for _, m := range members {
			if strings.ToLower(m.User.Username) == username && m.User.Discriminator == userFull[1] {
				user = m.User
				return
			}
		}

		pos = members[len(members)-1].User.ID
	}
}

// ChannelFind searches the guilds accessible for a particular channel and returns the object if found.
func (state *Instance) ChannelFind(name string) (channel *discordgo.Channel, err error) {
	session := state.Session
	var channels []*discordgo.Channel
	guilds, err := session.UserGuilds(100, "", "")
	if err != nil {
		err = fmt.Errorf("could not get server info")
		return
	}

	for _, g := range guilds {
		state.Guild = g
		channels, err = session.GuildChannels(g.ID)
		if err != nil {

			return
		}
		for _, c := range channels {
			if strings.ToLower(c.Name) == strings.ToLower(name) {
				channel = c
				return
			}
		}
	}

	err = fmt.Errorf("no channels found")
	return
}

// ChannelExist attempts to discover a channel, if it doesn't find it... creates it and return object.
func (state *Instance) ChannelExist(name string) (channel *discordgo.Channel, err error) {
	s := state.Session
	channel, err = state.ChannelFind(name)
	if err != nil {
		if err.Error() != "no channels found" {
			return
		}
		channel, err = s.GuildChannelCreate(state.Guild.ID, "events", "text")
		if err != nil {
			return
		}
		_, err = state.Session.ChannelMessageSend(channel.ID, fmt.Sprintf("Messages will be updated within `15 seconds`.\nServer updates `once every minute`.\n%s", MSGDIV))
		if err != nil {
			return
		}
	}
	return
}
