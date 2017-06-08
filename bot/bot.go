package bot

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

// Codes for types of requests.
const (
	Add = 1 << iota
	Modify
	Delete
)

const (
	cmdADD = 1 << iota
	cmdMODIFY
	cmdDELETE
	cmdEVENT
	cmdSCRIPT
	cmdVENDOR
)

// Variables to be used for messages and channels.
var (
	UpdateChannel = "events"
	UNXRes        = "Unexpected issue occured. It has been documented. :)"
	MSGDIV        = "~~**>----------------------------------------<**~~"
)

// New creates a new instance to be used. Return pointer.
func New(db *sql.DB, dg *discordgo.Session) (s *Instance) {
	s = &Instance{
		Database: db,
		Session:  dg,
		Cooldown: 40,
	}
	return
}

func makeResponse(err error, emsg, smsg string) *Response {
	if emsg == "" {
		emsg = UNXRes
	}
	return &Response{Err: err, Errmsg: emsg, Sndmsg: smsg}
}

func (state *Instance) cmdconv() []string {
	input := state.Cmd
	str := fmt.Sprintf("%s %s", input.Command, strings.Join(input.Args, " "))
	return strings.Split(str, " ")
}

// SetAttr returns attributes
func SetAttr(input []string) int {
	or := 0
	if len(input) > 1 {
		for i := 0; i < 2; i++ {
			switch strings.ToLower(input[i]) {
			case ",add":
				or = or | cmdADD
			case ",mod":
				or = or | cmdMODIFY
			case ",del":
				or = or | cmdDELETE
			case "event":
				or = or | cmdEVENT
			case "script":
				or = or | cmdSCRIPT
			case "vendor":
				or = or | cmdVENDOR
			}
		}
	}
	return or
}

func (state *Instance) modifierSet() bool {
	m := state.Cmd.Attr
	switch {
	case m&cmdADD == cmdADD:
		return true
	case m&cmdMODIFY == cmdMODIFY:
		return true
	case m&cmdDELETE == cmdDELETE:
		return true
	default:
		return false
	}
}
