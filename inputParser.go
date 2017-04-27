package main

import (
	"fmt"
	"strings"

	"github.com/d0x1p2/DiscordBot-go/bot"
)

func inputParser(state *bot.Instance) (res *bot.Response) {

	switch state.Cmd.Command {
	case "events":
		fallthrough
	case "event":
		fallthrough
	case "grant":
		fallthrough
	case "add":
		fallthrough
	case "del":
		fallthrough
	case "mod":
		fallthrough
	case "blacklist":
		fallthrough
	case "report":
		res = state.DBCore()
	case "ctf":
		fallthrough
	case "online":
		fallthrough
	case "player":
		fallthrough
	case "item":
		res = state.VNCCore()
	case "version":
		res = &bot.Response{Err: nil, Errmsg: "", Sndmsg: fmt.Sprintf("version: `%s`", _version)}
	default:
		res = state.DBCore()
	}

	return
}

func inputText(input string) (command *bot.Command) {
	_proc := false // test for stripping/fixing quotations
	var text string
	var inputLen int
	var iodat bot.Command

	s := strings.Split(input, " ")

	for k, p := range s[1:] {
		if strings.HasPrefix(p, "\"") {
			text = strings.Join(s[k+1:], " ")
			text = strings.TrimPrefix(text, "\"")
			if strings.HasSuffix(text, "\"") {
				text = strings.TrimSuffix(text, "\"")
			}
			_proc = true
			inputLen = k + 1
			break
		}
	}

	if _proc == false {
		text = s[len(s)-1]
		inputLen = len(s) - 1
	}

	iodat.Attr = bot.SetAttr(s)
	iodat.Args = s[1:]
	iodat.Length = inputLen
	iodat.Text = text
	if strings.HasPrefix(s[0], ",") {
		iodat.Command = s[0][1:]
	} else {
		iodat.Command = s[0]
	}

	command = &iodat
	return

}
