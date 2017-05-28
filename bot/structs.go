package bot

import (
	"database/sql"

	"github.com/bwmarrin/discordgo"
)

// Instance is a single instance of a message and it's processing.
type Instance struct {
	Admin     bool // If user and channel are set- true
	Sendmsg   bool // Send data thru channel or send to console.
	Streaming bool // If streaming...
	//NotifyB      int  // 1 = 15seconds, 2 = 30seconds, base timeout on ticks.
	//BattleNotify bool
	//NotifyE      int
	//EventNotify  bool
	Event     *Notifier
	Battle    *Notifier
	Cooldown  int // Ticks till send another message.
	User      *discordgo.User
	Message   *discordgo.MessageCreate
	Channel   *discordgo.Channel
	Guild     *discordgo.UserGuild
	Cmd       *Command
	Session   *discordgo.Session
	Database  *sql.DB
	BG        *Battleground
	EventChan *discordgo.Channel
	MainChan  *discordgo.Channel
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
	Battles []BattleID
}

// BattleID contains the ID, Message ID (to edit), and Name of the battle.
type BattleID struct {
	//ID    int64
	MsgID string
	Name  string
	Reset int // Counter until resend messages.
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
	Value  int // Used for storing numbers from some functions
}

// Notifier holds information as what/whom to notify
type Notifier struct {
	Type     int
	Tick     int
	Notified bool
	Msg      *discordgo.Message
}

// User data
type User struct {
	ID   string
	Name string
}
