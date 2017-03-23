package main

import (
	"fmt"
	"strings"
)

func inputParser(info *inputInfo) (string, error) {

	var sndmsg string
	var err error

	switch info.dat.command {
	case "yay":
		fallthrough
	case "nay":
	case "event":
		sndmsg, err = sqlCMDEvent()
	case "grant":
		sndmsg, err = sqlCMDGrant(info)
	case "add":
		sndmsg, err = sqlCMDAdd(info)
	case "del":
		sndmsg, err = sqlCMDDel(info)
	case "mod":
		sndmsg, err = sqlCMDMod(info)
	case "blacklist":
		sndmsg, err = sqlCMDBlacklist(info)
	case "report":
		sndmsg, err = sqlCMDReport(info)
	case "version":
		sndmsg = fmt.Sprintf("version: `%s`", _version)
	default:
		sndmsg, err = sqlCMDSearch(info.dat, info.dat.length+1)
	}

	return sndmsg, err
}

func inputText(input string) *inputDat {
	_proc := false // test for stripping/fixing quotations
	var text string
	var inputLen int
	var iodat inputDat

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

	iodat.attr = setAttr(s)
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

func cmdconv(info *inputDat) []string {

	str := fmt.Sprintf("%s %s", info.command, strings.Join(info.args, " "))
	return strings.Split(str, " ")

}

func setAttr(input []string) int {
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

func modifierSet(info *inputDat) bool {
	m := info.attr
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
