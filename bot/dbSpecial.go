package bot

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
)

// Compares current time to that which is stored and returns
// the result for all that are described in the table.
func (state *Instance) dbBattle() (res *Response) {

	switch strings.ToLower(state.Cmd.Command) {
	case "ctf":
	case "1v1":
	case "2v2":
	case "event":
	case "events":
	default:
		err := fmt.Errorf("bad command")
		res = makeResponse(err, err.Error(), "")
		return res
	}

	// Create initial message.
	c, err := state.ChannelFind("battlegrounds")
	if err != nil {
		res = makeResponse(err, "", "")
		return
	}

	res = makeResponse(nil, "", fmt.Sprintf("Auto-updates at: <#%s>", c.ID))
	return
}

func (state *Instance) dbEvent() (res *Response) {
	db := state.Database
	var weekday, hhmmFull, retText string
	var hh, mm, cnt int
	var err error

	now := time.Now()
	retText = "Events Coming Soon\n```"

	rows, err := db.Query("SELECT weekday, time FROM events")
	if err != nil {
		res = makeResponse(err, "Could not get events.", "")
		return
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&weekday, &hhmmFull)
		if err != nil {
			res = makeResponse(err, "", "")
			return
		}
		hhmm := strings.Split(hhmmFull, ":")
		hh, err = strconv.Atoi(hhmm[0])
		if err != nil {
			res = makeResponse(err, "", "")
			return
		}
		mm, err = strconv.Atoi(hhmm[1])
		if err != nil {
			res = makeResponse(err, "", "")
			return
		}

		var num, alt int
		cnt++
		switch strings.ToLower(weekday) {
		case "sunday":
			num = 0
			alt = 7
		case "monday":
			num = 1
		case "tuesday":
			num = 2
		case "wednesday":
			num = 3
		case "thursday":
			num = 4
		case "friday":
			num = 5
		case "saturday":
			num = 6
		}
		// BETWEEN HERE FOR ISSUE
		var dayAdd int
		if num < int(now.Weekday()) {
			dayAdd = alt - int(now.Weekday())
		} else if num > int(now.Weekday()) {
			dayAdd = num - int(now.Weekday())
		} else {
			dayAdd = 0
		}

		var dur time.Duration
		next := time.Date(now.Year(), now.Month(), now.Day()+dayAdd, hh, mm, 0, 0, now.Location())
		dur = next.Sub(now)
		// AND HERE

		hour := int(dur.Hours())
		if hour < -12 {
			next = time.Date(now.Year(), now.Month(), next.Day()+7, hh, mm, 0, 0, now.Location())
			dur = next.Sub(now)
			hour = int(dur.Hours())
		}

		min := int(dur.Minutes()) % 60

		hourText := "hours"
		if dur.Hours() < 2 || dur.Hours() > 0 {
			hourText = "hour"
		}

		minText := "minutes"
		if min < 2 || min > 0 {
			minText = "minute"
		} else if min < 0 {
			min = -min
		}

		if hour == 1 && min == 0 {
			if state.Event.Notified {
				state.Event.Notified = false
				// 1 hour notification
				state.notify(notifyEvent, "This is your 1 hour notification for an event!")
			}
		} else if hour == 0 && min == 30 {
			if state.Event.Notified {
				state.Event.Notified = false
				// 30min notification
				state.notify(notifyEvent, "This is your 30minute notification for an event!")
			}
		}

		event := fmt.Sprintf("%2d)  %3d %5s %2d %7s ->  %8s - %s CST\n", cnt, hour, hourText, min, minText, next.Weekday().String(), next.Format("15:04"))
		retText += event

		// END SOME STUFF
	}
	err = rows.Err()
	if err != nil {
		res = makeResponse(err, "", "")
		return
	}

	if cnt < 1 {
		res = makeResponse(nil, "", "`No events found.`")
		return
	}

	retText = strings.Trim(retText, " \n")
	var checktxt = "#events"
	if state.EventChan != nil {
		checktxt = "<#" + state.EventChan.ID + ">"
	}
	res = makeResponse(nil, "", retText+"```check "+checktxt)
	return
}

// Looks up a command returns all other commands with the same base
func (state *Instance) dbHelp(input []string) (res *Response) {
	db := state.Database
	var retStr string
	var arg0, arg1, arg2 sql.NullString
	var err error

	cnt := 0
	retStr = fmt.Sprintf("%s help: ```\n", input[0])

	rows, err := db.Query("SELECT command, arg1, arg2 FROM commands WHERE command=(?)", input[0])
	if err != nil {
		res = makeResponse(err, "could not retrieve information", "")
		return
	}
	defer rows.Close()

	for rows.Next() {
		var str0, str1, str2 string
		err := rows.Scan(&arg0, &arg1, &arg2)
		if err != nil {
			res = makeResponse(err, "issue processing information", "")
			return
		}
		cnt++
		if arg0.Valid {
			str0 = arg0.String
		}
		if arg1.Valid {
			str1 = arg1.String
		} else {
			break
		}
		if arg2.Valid {
			str2 = arg2.String
		}

		retStr += fmt.Sprintf("%2d: %s %s %s\n", cnt, str0, str1, str2)

	}
	err = rows.Err()
	if err != nil {
		res = makeResponse(err, "", "")
		return
	}

	// Prevent returning ugly text.
	if cnt < 1 {
		res = makeResponse(err, "", "")
		return
	}
	retText := strings.Trim(retStr, " \n")
	res = makeResponse(nil, "", retText+"```")

	return
}

// BlacklistWrapper just wraps dbBlackListGet, this is a temporary measure.
func (state *Instance) BlacklistWrapper(input string) bool {
	return state.dbBlacklistGet(input)
}

// Check if user is blacklisted from using the bot
func (state *Instance) dbBlacklistGet(input string) bool {
	db := state.Database
	var status sql.NullBool
	err := db.QueryRow("SELECT status FROM blacklist WHERE name=(?)", input).Scan(&status)
	if err != nil {
		// If row doesn't exist, return false
		if err == sql.ErrNoRows {
			return false
		}
		return false
	} else if status.Valid && status.Bool {
		return true
	}
	return false
}

// Blacklists a player and prevents them from using.
// This could be for a number of things such as,
// Bot abuse, in-service abuse, etc
func (state *Instance) dbUserBlacklist() (res *Response) {
	var err error
	user := state.User
	input := state.Cmd
	db := state.Database
	userFull := fmt.Sprintf("%s#%s", user.Username, user.Discriminator)

	if input.Length != 1 {
		err = fmt.Errorf("not enough arguments")
		res = makeResponse(err, err.Error(), "")
		return
	} else if state.dbUserPermissions(user.ID) == false {
		err = fmt.Errorf("you do not have permissions to do that")
		res = makeResponse(err, err.Error(), "")
		return
	}

	// Find user, get ID. Return on bad ID
	reported, err := state.UserFind(input.Args[0])
	if err != nil {
		res = makeResponse(err, "Could not find user specified.", "")
		return
	}

	criminal := fmt.Sprintf("%s#%s", reported.Username, reported.Discriminator)
	// Check if it exists already
	var status sql.NullBool
	err1 := db.QueryRow("SELECT status FROM blacklist WHERE name=(?)", criminal).Scan(&status)
	if err1 != nil {
		if err1 == sql.ErrNoRows {
			// Add to table
			_, err = db.Exec("INSERT INTO blacklist (name, status, times, start_date, who) VALUES (?, true, 1, Now(), ?)", criminal, userFull)
			if err != nil {
				res = makeResponse(err, "could not add new blacklistee", "")
				return
			}
			// Rows were found, another issue occured.
			res = makeResponse(nil, "", fmt.Sprintf("<@%s> has blacklisted <@%s>. Sucks to suck.", user.ID, reported.ID))
			return
		}
	}
	if status.Valid && status.Bool == false {
		_, err = db.Exec("UPDATE blacklist SET times = times+1, start_date = Now(), who = (?), status = true WHERE name = (?)", userFull, criminal)
		if err != nil {
			res = makeResponse(err, "could not blacklist user", "")
			return
		}
	}
	res = makeResponse(nil, "", fmt.Sprintf("<@%s> has blacklisted <@%s>. Sucks to suck.", user.ID, reported.ID))
	return
}

// Updates a counter for amount of reports on a player.
// Adds a new entry if it is not existant.
func (state *Instance) dbUserReport() (res *Response) {
	user := state.User
	input := state.Cmd
	db := state.Database
	userFull := fmt.Sprintf("%s#%s", user.Username, user.Discriminator)
	var amount sql.NullInt64
	var err error

	reportUser, err := state.UserFind(input.Args[0])
	if err != nil {
		res = makeResponse(err, err.Error(), "")
		return
	}

	criminal := fmt.Sprintf("%s#%s", reportUser.Username, reportUser.Discriminator)

	if input.Length != 1 {
		err = fmt.Errorf("not enough arguments. want: [report] [username#1234]")
		res = makeResponse(err, err.Error(), "")
		return
	}

	// Get info for amount of times previously reported. If 0, add.
	err = db.QueryRow("SELECT times FROM blacklist WHERE name=(?)", criminal).Scan(&amount)
	if err != nil {
		// If rows are not found. Insert into table.
		if err == sql.ErrNoRows {
			_, err := db.Exec("INSERT INTO blacklist (name, status, reports) VALUES (?, false, 1)", criminal)
			if err != nil {
				res = makeResponse(err, "could not report user", "")
				return
			}
		}
		// Rows were found, another issue occured.
		res = makeResponse(err, "could not report user", "")
		return
	}
	// Update the tables counter for reporting.
	if amount.Valid && amount.Int64 > 0 {
		_, err := db.Exec("UPDATE blacklist SET reports = reports+1 WHERE name = (?)", criminal)
		if err != nil {
			res = makeResponse(err, "could not report user", "")
			return
		}
	}

	res = makeResponse(nil, "", fmt.Sprintf("%s has reported %s.", userFull, criminal))
	return
}

func (state *Instance) loadBGEvent() (res *Response) {
	c, err := state.ChannelFind("battlegrounds")
	if err != nil {
		res = makeResponse(err, err.Error(), "")
		return
	}

	m, err := state.Session.ChannelMessageSend(c.ID, "``` [PlaceHolder] ```")
	if err != nil {
		res = makeResponse(err, err.Error(), "")
		return
	}

	// Save to SQL to prevent lookups.
	db := state.Database

	_, err = db.Exec("INSERT INTO battlegrounds (msgid, name) VALUES (?, ?)", m.ID, "event")
	if err != nil {
		res = makeResponse(err, err.Error(), "")
		return
	}

	var found bool
	if len(state.BG.Battles) > 0 {
		for _, p := range state.BG.Battles {
			if p.Name == "event" {
				found = true
			}
		}
	}

	if found == false {
		state.BG.Battles = append(state.BG.Battles, BattleID{MsgID: m.ID, Name: "event"})
	}

	res = makeResponse(nil, "", "Loaded and saved.")
	return
}

func (state *Instance) dbGetChannel(name string) (res *Response) {
	db := state.Database
	session := state.Session
	var cid sql.NullString
	var err error

	err = db.QueryRow("SELECT id FROM config WHERE name=(?)", name).Scan(&cid)
	if err != nil {
		res = makeResponse(err, "", cid.String)
		return
	}

	switch name {
	case "main":
		state.MainChan, err = session.Channel(cid.String)
	case "event":
		state.EventChan, err = session.Channel(cid.String)
	default:
		err := fmt.Errorf("bad channel to load")
		res = makeResponse(err, err.Error(), "")
		return
	}

	res = makeResponse(err, "", "")
	return
}

// LoadChannel gets a channel from the Database
func (state *Instance) LoadChannel(chanType int) (res *Response, bad int) {
	for chanType > 0 {
		switch {
		case chanType&1 == 1:
			chanType ^= 1
			res = state.dbGetChannel("main")
			if res.Err != nil {
				bad |= 1
			}
		case chanType&2 == 2:
			chanType ^= 2
			res = state.dbGetChannel("event")
			if res.Err != nil {
				bad |= 2
			}
		default:
			err := fmt.Errorf("bad channel provided")
			res = makeResponse(err, err.Error(), "")
			return
		}
	}
	return
}

// SaveChannel stores channel info into Database
func (state *Instance) SaveChannel(name string) (res *Response) {
	db := state.Database
	res = state.dbGetChannel(name)
	var err error
	var id string

	switch name {
	case "main":
		if state.MainChan != nil {
			id = state.MainChan.ID
			break
		}
		res = makeResponse(fmt.Errorf("channel not set"), "channel not set", "")
	case "event":
		if state.EventChan != nil {
			id = state.EventChan.ID
			break
		}
		res = makeResponse(fmt.Errorf("channel not set"), "channel not set", "")
	default:
		res = makeResponse(fmt.Errorf("bad channel to set"), "bad channel to set", "")
	}

	if res.Err != nil {
		// May not exist.
		if res.Err == sql.ErrNoRows {
			// Doesn't exist, INSERT
			_, err = db.Query("INSERT INTO config (id, name) VALUES (?, ?)", id, name)
			if err != nil {
				res = makeResponse(err, err.Error(), "")
			}
			return
		}
		// Other issue, return
		return
	}
	// Exists, update.
	_, err = db.Query("UPDATE config SET id=(?) WHERE name=(?)", id, name)
	if err != nil {
		res = makeResponse(err, err.Error(), "")
	}

	return
}

// SetChannels is just a wrapper that will likely be removed in the future.
// Sets main and Event channels.
func (state *Instance) SetChannels(chann int) (err error) {
	r, b := state.LoadChannel(chann)
	if r.Err != nil && b&1 == 1 {
		state.MainChan, err = state.ChannelFind("multiverse")
		if err != nil {
			log.Println("Error loading channel:", err.Error())
			return
		}
		r = state.SaveChannel("main")
		if err != nil {
			log.Println("Error saving channel:", err.Error())
			return
		}
	}
	//
	if r.Err != nil && b&2 == 2 {
		// EVENTS CHANNEL HERE
		state.EventChan, err = state.ChannelFind("events")
		if err != nil {
			log.Println("Error loading channel:", err.Error())
			return
		}
		r = state.SaveChannel("event")
		if err != nil {
			log.Println("Error saving channel:", err.Error())
			return
		}
	}

	return
}
