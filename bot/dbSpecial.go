package bot

import (
	"database/sql"
	"fmt"
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

		hourText := "hour"
		if dur.Hours() > 1 || dur.Hours() < -1 {
			hourText = "hours"
		}
		minText := "minute"
		if min > 1 || min < -1 {
			minText = "minutes"
		}
		event := fmt.Sprintf("%2d)  %2d %5s %2d %7s ->  %8s - %s CST\n", cnt, hour, hourText, min, minText, next.Weekday().String(), next.Format("15:04"))
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
	res = makeResponse(nil, "", retText+"```check #events")
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
