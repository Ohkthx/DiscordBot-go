package main

import (
	"fmt"
	"strings"
)

func inputParser(info *inputInfo) string {

	var sndmsg string

	switch info.dat.command {
	case "event":
		sndmsg = sqlCMDEvent()
	case "grant":
		sndmsg = sqlCMDGrant(info)
	case "add":
		sndmsg = sqlCMDAdd(info)
	case "del":
		sndmsg = sqlCMDDel(info)
	case "mod":
		sndmsg = sqlCMDMod(info)
	case "blacklist":
		sndmsg = sqlCMDBlacklist(info)
	case "report":
		sndmsg = sqlCMDReport(info)
	case "version":
		sndmsg = fmt.Sprintf("version: `%s`", _version)
	default:
		sndmsg = sqlCMDSearch(info)
	}

	return sndmsg
}

func inputText(input string) *inputDat {

	s := strings.Split(input, " ")

	_proc := false // test for stripping/fixing quotations
	var text string
	var inputLen int
	var iodat inputDat

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

	iodat.args = s[1:]
	iodat.length = inputLen
	iodat.text = text
	if strings.HasPrefix(s[0], ",") {
		iodat.command = s[0][1:]
	} else {
		iodat.command = s[0]
	}

	return &iodat

}
