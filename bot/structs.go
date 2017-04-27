package bot

import (
	"database/sql"

	"github.com/bwmarrin/discordgo"
)

// Instance is a single instance of a message and it's processing.
type Instance struct {
	Admin     bool // if user and channel are set- true
	Sendmsg   bool // Send data thru channel or send to console.
	Streaming bool // If streaming...
	User      *discordgo.User
	Channel   *discordgo.Channel
	Guild     *discordgo.UserGuild
	Cmd       *Command
	Session   *discordgo.Session
	Database  *sql.DB
	BG        *Battleground
}

// Command contains information sent by user
type Command struct {
	Text    string
	Command string
	Args    []string
	Length  int
	Attr    int
}

// Battleground is a single instance of the #battlegrounds channel
type Battleground struct {
	// ChannelID is the ID of #battlegrounds
	ChannelID string
	// Battles is an array containing all current battles running.
	Battles []Battle
}

// Battle contains the ID, Message ID (to edit), and Name of the battle.
type Battle struct {
	//ID    int64
	MsgID string
	Name  string
}

// DBError is a wrapper for Error.
// Levels to indicate if should continue processing or return early.
// -1 - Fatal, stop processing.
// 0 - Continue processing, bad input.
type DBError struct {
	Err   error
	Level int
}

// Response will be returned for Database commands to inform the user.
type Response struct {
	Err    error
	Sndmsg string
	Errmsg string
}
