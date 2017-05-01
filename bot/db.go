package bot

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

// DBCore will handle all SQL DB interactions.
func (state *Instance) DBCore() (res *Response) {

	switch state.Cmd.Command {
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
