package main

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
)

func sqlCheckPerm(id string) bool {
	var val string
	err := db.QueryRow("SELECT id FROM permissions WHERE id=(?)", id).Scan(&val)
	if err != nil {
		errLog.Println(err)
		return false
	}

	return true
}

func sqlCMDGrant(info *inputInfo) string {

	input := info.dat
	who := info.user
	whoFull := fmt.Sprintf("%s#%s", who.Username, who.Discriminator)

	if input.length != 1 {
		discordLog.Println("Could not grant permissions.")
		return ""
	} else if sqlCheckPerm(who.ID) == false {
		discordLog.Println(whoFull + "(" + who.ID + ") attempted to grant permissions to " + input.args[0] + ".")
		return ""
	}

	if info.channelID == "" {
		info.channelID = info.channel.ID
	}

	// Find user, get ID
	addee := userFind(info.channelID, input.args[0])
	if addee == nil {
		discordLog.Println("Bad user")
		return ""
	}
	addeeID := addee.ID
	if addeeID == "" {
		discordLog.Printf("User [%s] not found. Missing discriminator (#000)?\n", input.args[0])
		return ""
	}
	addeeUsername := fmt.Sprintf("%s", input.args[0])

	_, err := db.Exec("INSERT INTO permissions (id, username, allow, date_added, accountable) VALUES (?, ?, false, Now(), ?)",
		addeeID, addeeUsername, whoFull)
	if err != nil {
		errLog.Println(err)
	}

	return fmt.Sprintf("%s granted permissions to use `,add` by %s", addeeUsername, whoFull)
}

func sqlCMDAdd(info *inputInfo) string {

	input := info.dat
	who := info.user
	whoFull := fmt.Sprintf("%s#%s", who.Username, who.Discriminator)

	if input.length < 2 {
		discordLog.Println("Bad add request")
		return ""
	} else if sqlCheckPerm(who.ID) == false {
		discordLog.Println(whoFull + "(" + who.ID + ") attempted to add a command.")
		return ""
	}

	// Check if it exists already
	existing := sqlCMDSearch(input, input.length-1)
	if existing != "" {
		discordLog.Println("Already exists in database")
		return ""
	}

	var err error
	var added string

	switch input.length {
	case 2:
		_, err = db.Exec("INSERT INTO commands (command, text, author, date_added, author_mod, date_mod) VALUES (?, ?, ?, Now(), ?, Now())", input.args[0], input.text, whoFull, whoFull)
		if len(input.text) > 40 {
			added = fmt.Sprintf("[Added: %s] %s -> %s...", whoFull, input.args[0], input.text[0:40])
		} else {
			added = fmt.Sprintf("[Added: %s] %s -> %s", whoFull, input.args[0], input.text)
		}
	case 3:
		if input.args[0] == "script" || input.args[0] == "event" {
			return sqlProxyAdd(info)
		}
		_, err = db.Exec("INSERT INTO commands (command, arg1, text, author, date_added, author_mod, date_mod) VALUES (?, ?, ?, ?, Now(), ?, Now())", input.args[0], input.args[1], input.text, whoFull, whoFull)
		if len(input.text) > 40 {
			added = fmt.Sprintf("[Added: %s] %s %s -> %s...", whoFull, input.args[0], input.args[1], input.text[0:40])
		} else {
			added = fmt.Sprintf("[Added: %s] %s %s -> %s", whoFull, input.args[0], input.args[1], input.text)
		}
	case 4:
		_, err = db.Exec("INSERT INTO commands (command, arg1, arg2, text, author, date_added, author_mod, date_mod) VALUES (?, ?, ?, ?, ?, Now(), ?, Now())", input.args[0], input.args[1], input.args[2], input.text, whoFull, whoFull)
		if len(input.text) > 40 {
			added = fmt.Sprintf("[Added: %s] %s %s %s -> %s...", whoFull, input.args[0], input.args[1], input.args[2], input.text[0:40])
		} else {
			added = fmt.Sprintf("[Added: %s] %s %s %s -> %s", whoFull, input.args[0], input.args[1], input.args[2], input.text)
		}
	default:
		discordLog.Println("too many inputs.")
		return ""
	}

	if err != nil {
		errLog.Println(err)
		return ""
	}

	return added

}

func sqlCMDDel(info *inputInfo) string {

	input := info.dat
	who := info.user
	whoFull := fmt.Sprintf("%s#%s", who.Username, who.Discriminator)

	if input.length < 1 {
		discordLog.Println("Bad delete request")
		return ""
	} else if sqlCheckPerm(who.ID) == false {
		discordLog.Println(whoFull + "(" + who.ID + ") attempted to delete a command.")
		return ""
	}

	// Check if it exists already
	existing := sqlCMDSearch(input, input.length)
	if existing == "" {
		discordLog.Println("Command doesn't exist")
		return ""
	}

	var deleted string
	var err error
	switch input.length {
	case 1:
		_, err = db.Exec("DELETE FROM commands WHERE command=(?) AND arg1 IS NULL AND author=(?)", input.args[0], whoFull)
		deleted = fmt.Sprintf("[%s deleted]: -> %s", whoFull, input.args[0])
	case 2:
		if input.args[0] == "script" || input.args[0] == "event" {
			return sqlProxyDel(info)
		}
		_, err = db.Exec("DELETE FROM commands WHERE command=(?) AND arg1=(?) AND arg2 IS NULL AND author=(?)", input.args[0], input.args[1], whoFull)
		deleted = fmt.Sprintf("[%s deleted]: -> %s %s", whoFull, input.args[0], input.args[1])
	case 3:
		_, err = db.Exec("DELETE FROM commands WHERE command=(?) AND arg1=(?) AND arg2=(?) AND author=(?)", input.args[0], input.args[1], input.args[0], whoFull)
		deleted = fmt.Sprintf("[%s deleted]: -> %s %s %s", whoFull, input.args[0], input.args[1], input.args[2])
	default:
		discordLog.Println("too many inputs.")
		return ""
	}

	if err != nil {
		errLog.Println(err)
		return ""
	}

	return deleted
}

func sqlCMDMod(info *inputInfo) string {

	input := info.dat
	who := info.user
	whoFull := fmt.Sprintf("%s#%s", who.Username, who.Discriminator)

	if input.length < 2 {
		discordLog.Println("Bad modify request")
		return ""
	} else if sqlCheckPerm(who.ID) == false {
		discordLog.Println(whoFull + "(" + who.ID + ") attempted to modify a command.")
		return ""
	}

	// Check if it exists already
	existing := sqlCMDSearch(input, input.length-1)
	if existing == "" {
		discordLog.Println("Doesn't exist in database")
		return ""
	}

	var modified string
	var err error
	switch input.length {
	case 2:
		_, err = db.Exec("UPDATE commands SET text=(?), author_mod=(?), date_mod=Now() WHERE command=(?) arg1 IS NULL AND author=(?)", input.text, whoFull, input.args[0], whoFull)
		modified = fmt.Sprintf("[%s updated]: -> %s", whoFull, input.args[0])
	case 3:
		if input.args[0] == "script" || input.args[0] == "event" {
			return sqlProxyMod(info)
		}
		_, err = db.Exec("UPDATE commands SET text=(?), author_mod=(?), date_mod=Now() WHERE command=(?) AND arg1=(?) AND arg2 IS NULL AND author=(?)", input.text, whoFull, input.args[0], input.args[1], whoFull)
		modified = fmt.Sprintf("[%s updated]: -> %s %s", whoFull, input.args[0], input.args[1])
	case 4:
		_, err = db.Exec("UPDATE commands SET text=(?), author_mod=(?), date_mod=Now() WHERE command=(?) AND arg1=(?) AND author=(?)", input.text, whoFull, input.args[0], input.args[1], input.args[2], whoFull)
		modified = fmt.Sprintf("[%s updated]: -> %s %s %s", whoFull, input.args[0], input.args[1], input.args[2])
	default:
		discordLog.Println("too many inputs.")
		return ""
	}

	if err != nil {
		errLog.Println(err)
		return ""
	}

	return modified
}

func sqlCMDSearch(input *inputDat, length int) string {

	var err error
	var text string
	i := input.args

	// Change format of the command structure.

	if input.command == "script" {
		text, _ := sqlProxyLinkGET(input.args[0], text)
		return text
	} else if input.command == "help" {
		return sqlCMDHelp(input.args)
	} else if input.modifier != true {
		i = cmdconv(input)
	}

	switch length {
	case 1:
		err = db.QueryRow("SELECT text FROM commands WHERE command=(?) AND arg1 IS NULL", i[0]).Scan(&text)
	case 2:
		err = db.QueryRow("SELECT text FROM commands WHERE command=(?) AND arg1=(?) AND arg2 IS NULL", i[0], i[1]).Scan(&text)
	case 3:
		err = db.QueryRow("SELECT text FROM commands WHERE command=(?) AND arg1=(?) AND arg2=(?)", i[0], i[1], i[2]).Scan(&text)
	default:
		return ""
	}
	if err != nil {
		errLog.Println(err)
		return ""
	}

	return text
}

func sqlCMDEvent() string {

	var weekday, hhmmFull, retText string
	var hh, mm, cnt int

	now := time.Now()
	retText = "Events Coming Soon\n```"

	rows, err := db.Query("SELECT weekday, time FROM events")
	if err != nil {
		errLog.Println("Event lookup (getting rows): ", err)
		return ""
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&weekday, &hhmmFull)
		if err != nil {
			errLog.Println("Event lookup (proc rows): ", err)
			return ""
		}
		hhmm := strings.Split(hhmmFull, ":")
		hh, err = strconv.Atoi(hhmm[0])
		if err != nil {
			errLog.Println("Event hour conv: ", err)
			return ""
		}
		mm, err = strconv.Atoi(hhmm[1])
		if err != nil {
			errLog.Println("Event min conv: ", err)
			return ""
		}
		// DO SOME STUFF WITH EVENT
		var num int
		cnt++
		switch strings.ToLower(weekday) {
		case "sunday":
			num = 0
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

		dayAdd := num - int(now.Weekday())
		var dur time.Duration
		next := time.Date(now.Year(), now.Month(), now.Day()+dayAdd, hh, mm, 0, 0, now.Location())
		dur = next.Sub(now)

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
		event := fmt.Sprintf("%2d)  %d %s %d %s ->  %8s - %s CST\n", cnt, hour, hourText, min, minText, next.Weekday().String(), next.Format("15:04"))
		retText += event

		// END SOME STUFF
	}
	err = rows.Err()
	if err != nil {
		errLog.Println("Event lookup (unknown): ", err)
		return "No events set."
	}
	retText = strings.Trim(retText, " \n")
	return retText + "```check #events"
}

func sqlCMDHelp(input []string) string {
	var retStr string
	var arg0, arg1, arg2 sql.NullString

	cnt := 1
	retStr = fmt.Sprintf("%s help: ```\n", input[0])

	//
	if input[0] != "vendor" {
		return ""
	}

	rows, err := db.Query("SELECT command, arg1, arg2 FROM commands WHERE command=(?)", input[0])
	if err != nil {
		errLog.Printf("%s lookup (getting rows): %s", input[0], err)
		return ""
	}
	defer rows.Close()

	for rows.Next() {
		var str0, str1, str2 string
		err := rows.Scan(&arg0, &arg1, &arg2)
		if err != nil {
			errLog.Printf("%s lookup (proc rows): %s", input[0], err)
			return ""
		}

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
		cnt++
	}
	err = rows.Err()
	if err != nil {
		errLog.Printf("%s lookup (unknown): %s", input[0], err)
		return "No help :'("
	}

	retText := strings.Trim(retStr, " \n")
	return retText + "```"
}

func sqlProxyAdd(info *inputInfo) string {
	input := info.dat
	who := info.user
	whoFull := fmt.Sprintf("%s#%s", who.Username, who.Discriminator)

	id := sqlProxyLinkSET(input.args, input.text)
	if id == "" {
		errLog.Println("Bad ID returned.")
		return ""
	}

	_, err := db.Exec("INSERT INTO commands (command, arg1, text, author, date_added, author_mod, date_mod) VALUES (?, ?, ?, ?, Now(), ?, Now())", input.args[0], input.args[1], id, whoFull, whoFull)
	if err != nil {
		errLog.Println("Unable to add parent db", err)
		return ""
	}

	if len(input.text) > 40 {
		return fmt.Sprintf("[Added: %s] %s %s -> %s...", whoFull, input.args[0], input.args[1], input.text[0:40])
	}
	return fmt.Sprintf("[Added: %s] %s %s -> %s", whoFull, input.args[0], input.args[1], input.text)
}

func sqlProxyMod(info *inputInfo) string {
	input := info.dat
	who := info.user
	whoFull := fmt.Sprintf("%s#%s", who.Username, who.Discriminator)

	success := sqlProxyLinkMOD(input.args)
	if success != true {
		return ""
	}

	_, err := db.Exec("UPDATE commands SET text=(?), author_mod=(?), date_mod=Now() WHERE command=(?) AND arg1=(?) AND arg2 IS NULL AND author=(?)", input.text, whoFull, input.args[0], input.args[1], whoFull)
	if err != nil {
		return ""
	}

	return fmt.Sprintf("[%s updated]: -> %s %s", whoFull, input.args[0], input.args[1])
}

func sqlProxyDel(info *inputInfo) string {
	input := info.dat
	who := info.user
	whoFull := fmt.Sprintf("%s#%s", who.Username, who.Discriminator)

	success := sqlProxyLinkDEL(input.args)
	if success != true {
		return ""
	}

	_, err := db.Exec("DELETE FROM commands WHERE command=(?) AND arg1=(?) AND arg2 IS NULL AND author=(?)", input.args[0], input.args[1], whoFull)
	if err != nil {
		return ""
	}

	return fmt.Sprintf("[%s deleted]: -> %s %s", whoFull, input.args[0], input.args[1])
}

func sqlCMDBlacklist(info *inputInfo) string {
	// Check permisions for Blacklist
	user := info.user
	input := info.dat
	userFull := fmt.Sprintf("%s#%s", user.Username, user.Discriminator)
	if input.length != 1 {
		discordLog.Println("Not enough arguments")
		return ""
	} else if sqlCheckPerm(user.ID) == false {
		discordLog.Println(userFull + "(" + user.ID + ") attempted to add a command.")
		return ""
	}

	reportUser := userFind(info.channelID, input.args[0])
	if reportUser == nil {
		discordLog.Println("Bad user.")
		return ""
	}

	criminal := fmt.Sprintf("%s#%s", reportUser.Username, reportUser.Discriminator)
	// Check if it exists already
	var status bool
	err := db.QueryRow("SELECT status FROM blacklist WHERE name=(?)", criminal).Scan(&status)
	if err != nil && status == false {
		// Add to table
		_, err := db.Exec("INSERT INTO blacklist (name, status, times, start_date, who) VALUES (?, true, 1, Now(), ?)", criminal, userFull)
		if err != nil {
			discordLog.Println("Issue with init blacklisting user")
			return ""
		}
	} else {
		_, err := db.Exec("UPDATE blacklist SET times = times+1, start_date = Now(), who = (?), status = true WHERE name = (?)", userFull, criminal)
		if err != nil {
			discordLog.Println("Error updating blacklist")
			return ""
		}
	}
	return fmt.Sprintf("%s has blacklisted %s. Sucks to suck.", userFull, criminal)
}

func sqlCMDReport(info *inputInfo) string {

	user := info.user
	input := info.dat
	userFull := fmt.Sprintf("%s#%s", user.Username, user.Discriminator)
	var amount int

	reportUser := userFind(info.channelID, input.args[0])
	if reportUser == nil {
		discordLog.Println("Bad user.")
		return ""
	}
	criminal := fmt.Sprintf("%s#%s", reportUser.Username, reportUser.Discriminator)

	if input.length != 1 {
		discordLog.Println("Not enough arguments")
		return ""
	}

	err := db.QueryRow("SELECT times FROM blacklist WHERE name=(?)", criminal).Scan(&amount)
	if err != nil && amount == 0 {
		// Add to table
		_, err := db.Exec("INSERT INTO blacklist (name, status, reports) VALUES (?, false, 1)", criminal)
		if err != nil {
			discordLog.Println("Issue with init blacklisting user")
			return ""
		}
	} else {
		_, err := db.Exec("UPDATE blacklist SET reports = reports+1 WHERE name = (?)", criminal)
		if err != nil {
			discordLog.Println("Error updating blacklist")
			return ""
		}
	}

	return fmt.Sprintf("%s has reported %s.", userFull, criminal)
}

func sqlProxyLinkSET(args []string, text string) string {

	var err error
	var res sql.Result

	switch args[0] {
	case "script":
		res, err = db.Exec("INSERT INTO library (name, script) VALUES (?, ?)", args[1], text)
	case "event":
		res, err = db.Exec("INSERT INTO events (weekday, time) VALUES (?, ?)", args[1], text)
	default:
		errLog.Println("Could not determine type for addition")
		return ""
	}
	if err != nil {
		errLog.Println("Could not add to table", err)
		return ""
	}
	lastID, err := res.LastInsertId()
	if err != nil {
		log.Println("Could not retrieve ID for insert", err)
	}

	return fmt.Sprintf("%d", lastID)
}

func sqlProxyLinkGET(command, id string) (string, string) {
	var info1, info2 string
	var err error
	switch command {
	case "script":
		err = db.QueryRow("SELECT script FROM library WHERE id=(?)", id).Scan(&info1)
	case "event":
		err = db.QueryRow("SELECT weekday, time FROM events WHERE id=(?)", id).Scan(&info1, &info2)
	default:
		return "", ""
	}
	if err != nil {
		errLog.Println(err)
		return "", ""
	}
	return info1, info2
}

func sqlProxyLinkMOD(info []string) bool {

	var err error
	switch info[0] {
	case "script":
		_, err = db.Exec("UPDATE library SET script=(?) WHERE name=(?)", info[2], info[1])
	case "event":
		_, err = db.Exec("UPDATE events SET time=(?) WHERE weekday=(?)", info[2], info[1])
	default:
		return false
	}

	if err != nil {
		errLog.Println(err)
		return false
	}

	return true
}

func sqlProxyLinkDEL(info []string) bool {
	var err error
	switch info[0] {
	case "script":
		_, err = db.Exec("DELETE FROM library WHERE name=(?)", info[1])
	case "event":
		_, err = db.Exec("DELETE FROM events WHERE weekday=(?)", info[1])
	default:
		return false
	}

	if err != nil {
		errLog.Println(err)
		return false
	}

	return true
}

func sqlBlacklistGET(input string) bool {
	var status bool
	err := db.QueryRow("SELECT status FROM blacklist WHERE name=(?)", input).Scan(&status)
	if err != nil {
		errLog.Println(err)
		return false
	} else if status {
		return true
	}
	return false
}
