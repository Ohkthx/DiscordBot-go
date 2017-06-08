package bot

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/d0x1p2/vncgo"
)

// Constants for the different types of battles
const (
	EventBattle = 1 << iota
	CTFBattle
	FFABattle
	TvTBattle1v1
	TvTBattle2v2
	TvTBattle3v3
	TvTBattle4v4
	TvTBattle5v5
)

// Error Constants
var (
	ErrNotFound   = errors.New("not found")
	ErrBadRequest = errors.New("bad request")
	ErrExists     = errors.New("already exists")
)

// VNCCore handles all things that are for Vita-Nex: Core API
func (state *Instance) VNCCore() (res *Response) {
	switch state.Cmd.Command {
	case "event":
		fallthrough
	case "events":
		state.dbNotifyAdd(state.User.ID, state.User.Username)
		fallthrough
	case "ctf":
		res = state.dbBattle()
	case "online":
		res = getMobile("0x19c03e", "online")
	case "player":
		res = getMobile(state.Cmd.Args[0], "status")
	case "item":
		res = getItem(state.Cmd.Args[0], false)
	}
	return
}

func getMobile(mobileID, request string) (res *Response) {

	var sndmsg string

	s, err := vncgo.New()
	if err != nil {
		res = makeResponse(err, "", "")
		return
	}

	mobile, err := s.GetMobile(mobileID, false)
	if err != nil {
		res = makeResponse(err, "", "")
		return
	}

	switch strings.ToLower(request) {
	case "online":
		if mobile.Online == true {
			sndmsg = fmt.Sprintf("%s\n```\nStatus: Online\n```", mobile.FullName)
		} else {
			sndmsg = fmt.Sprintf("%s\n```\nStatus: Offline\n```", mobile.FullName)
		}
	case "status":
		sndmsg = fmt.Sprintf("%s\n```\nOnline: %v\nHealth: %d / %d\nDeaths: %d\nFame: %d, Karma: %d```",
			mobile.FullName, mobile.Online, mobile.Stats.Hits, mobile.Stats.HitsMax, mobile.Deaths, mobile.Fame, mobile.Karma)
	default:
		sndmsg = fmt.Sprintf("%s found.", mobile.FullName)
	}

	res = makeResponse(nil, "", sndmsg)
	return
}

func getItem(itemID string, internal bool) (res *Response) {
	const errStr = "unexpected error :)"

	s, err := vncgo.New()
	if err != nil {
		res = makeResponse(err, "", "")
		return
	}

	item, err := s.GetItem(itemID, false)
	if err != nil {
		res = makeResponse(err, "", "")
		return
	}

	itemDesc := vncgo.CleanHTML(item.Opl)
	res = makeResponse(nil, "", fmt.Sprintf("```\n%s```", itemDesc))

	return
}

/*

	BG editing of messages in #battlegrounds below

*/

//dbCheckBattle
func (state *Instance) dbCheckBattle(class string) (string, error) {
	var sqlMsgID sql.NullString
	db := state.Database

	// Load rows from DB
	err := db.QueryRow("SELECT msgid FROM events WHERE type=(?)", class).Scan(&sqlMsgID)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", ErrNotFound
		}
		return "", err
	}

	if sqlMsgID.Valid {
		return sqlMsgID.String, nil
	}

	return "", errors.New("invalid message id in db")

}

// dbLoadBattle attempts to find the battle information from SQL and assign it to the structure.
func (state *Instance) dbLoadBattle(class string) error {
	var sqlMsgID, sqlName sql.NullString
	db := state.Database

	// Load rows from DB
	err := db.QueryRow("SELECT msgid, name FROM events WHERE type=(?)", class).Scan(&sqlMsgID, &sqlName)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrNotFound
		}
		return err
	}

	if sqlMsgID.Valid && sqlName.Valid {
		if strings.ToLower(sqlName.String) == class {
			for _, p := range state.Events.Battles {
				if strings.ToLower(p.Name) == class {
					return ErrExists
				}
			}
			state.Events.Battles = append(state.Events.Battles, BattleID{Type: class, MsgID: sqlMsgID.String, Name: sqlName.String})
			return nil
		}
	}
	return ErrNotFound
}

func (state *Instance) bgSaveDB(class, msgid string) (err error) {
	db := state.Database

	_, err = db.Exec("INSERT INTO events (type, msgid) VALUES (?, ?)", class, msgid)
	if err != nil {
		return
	}

	return nil
}

// Perform checks on all BGs currently assigned in memory.
func (state *Instance) bgChecks() (err error) {
	//
	s := state.Session
	c := state.Events.ChannelID
	var r *Response
	var msg string

	if state.Events.Battles == nil {
		return ErrNilMem
	}

	for _, e := range state.Events.Battles {
		if e.Type == "Event" {
			r = state.dbEvent()
			//msg = fmt.Sprintf("%s%s", r.Sndmsg, MSGDIV)
			msg = r.Sndmsg + MSGDIV
		} else {
			// If it isn't an event type, but an actual battle, pull string here for message.
			msg, err = state.getBattleMSG(e.Type)
			if err != nil {
				return
			}
		}
		if msg != "" {
			// Edit message here with info obtained.
			_, err = s.ChannelMessageEdit(c, e.MsgID, msg)
			if err != nil {
				fmt.Printf("DEBUG/FAIL: [%s] %s\n", e.Type, err.Error())
			}
		}
	}

	return
}

func convBGName(in string) (out string, err error) {
	switch strings.ToLower(in) {
	case "capture the flag":
		out = "ctf"
	case "1 vs 1":
		out = "1v1"
	case "2 vs 2":
		out = "2v2"
	case "gladiator training":
		out = "gt"
	case "colosseum":
		out = "colosseum" // Free For All became Colosseum in the API
	case "ctf duality":
		out = "ctfd"
	case "ctf apocalypse":
		out = "ctfa"
	case "event":
		out = "event"
	default:
		err = fmt.Errorf("bad BG name")
	}
	return
}
func convBGNameReverse(in string) (out string, err error) {
	switch strings.ToLower(in) {
	case "ctf":
		out = "Capture the Flag"
	case "1v1":
		out = "1 vs 1"
	case "2v2":
		out = "2 v 2"
	case "gt":
		out = "Gladiator Training"
	case "colosseum":
		out = "Colosseum"
	case "ctfd":
		out = "CTF Duality"
	case "ctfa":
		out = "CTF Apocalypse"
	case "event":
		out = "Event"
	default:
		err = fmt.Errorf("bad BG name")
	}
	return
}
