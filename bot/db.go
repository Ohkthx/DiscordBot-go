package bot

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

// DBCore will handle all SQL DB interactions.
func (state *Instance) DBCore() (res *Response) {

	switch state.Cmd.Command {
	case "channel-base":
		state.procChannel(state.Channel, false)
		res = makeResponse(nil, "", "")
	case "channel-update":
		state.procChannel(state.Channel, true)
		res = makeResponse(nil, "", "")
	case "grant":
		res = state.dbUserPermissionsAdd()
	case "add":
		res = state.dbCommandAdd()
	case "del":
		res = state.dbCommandDelete()
	case "mod":
		res = state.dbCommandModify()
	case "blacklist":
		res = state.dbUserBlacklist()
	case "report":
		res = state.dbUserReport()
	case "ban":
		res = state.dbUserBan()
	case "rolls":
		fallthrough
	case "roll":
		res = state.dbRoll()
	default:
		res = state.dbSearch(state.Cmd.Length + 1)
	}

	return res
}

func tsConvert(ts discordgo.Timestamp) string {
	a := strings.FieldsFunc(fmt.Sprintf("%s", ts), tsSplit)
	return fmt.Sprintf("%s %s", a[0], a[1])
}

func tsSplit(r rune) bool {
	return r == 'T' || r == '.' || r == '+'
}
