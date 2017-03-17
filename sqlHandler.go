package main

import (
	"fmt"
	"log"
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
	addee := userFind(info.channelID, input.args[0], true)
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
	existing := sqlCMDSearch(info)
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
		if input.args[0] == "script" {
			id := sqlScriptSET(input.args[1], input.text)
			_, err = db.Exec("INSERT INTO commands (command, arg1, text, author, date_added, author_mod, date_mod) VALUES (?, ?, ?, ?, Now(), ?, Now())", input.args[0], input.args[1], id, whoFull, whoFull)
		} else {
			_, err = db.Exec("INSERT INTO commands (command, arg1, text, author, date_added, author_mod, date_mod) VALUES (?, ?, ?, ?, Now(), ?, Now())", input.args[0], input.args[1], input.text, whoFull, whoFull)
		}
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

	var deleted string
	var err error
	switch input.length {
	case 1:
		_, err = db.Exec("DELETE FROM commands WHERE command=(?) AND author=(?)", input.args[0], whoFull)
		deleted = fmt.Sprintf("[%s deleted]: -> %s", whoFull, input.args[0])
	case 2:
		_, err = db.Exec("DELETE FROM commands WHERE command=(?) AND arg1=(?) AND author=(?)", input.args[0], input.args[1], whoFull)
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

	var modified string
	var err error
	switch input.length {
	case 2:
		_, err = db.Exec("UPDATE commands SET text=(?), author_mod=(?), date_mod=Now() WHERE command=(?) AND author=(?)", input.text, whoFull, input.args[0], whoFull)
		modified = fmt.Sprintf("[%s updated]: -> %s", whoFull, input.args[0])
	case 3:
		if input.args[0] == "script" {
			res := sqlScriptMOD(input.args[1], input.text)
			if res == false {
				discordLog.Println("Error updating script.")
				return ""
			}
			_, err = db.Exec("UPDATE commands SET author_mod=(?), date_mod=Now() WHERE command=(?) AND arg1=(?) AND author=(?)", whoFull, input.args[0], input.args[1], whoFull)
		} else {
			_, err = db.Exec("UPDATE commands SET text=(?), author_mod=(?), date_mod=Now() WHERE command=(?) AND arg1=(?) AND author=(?)", input.text, whoFull, input.args[0], input.args[1], whoFull)
		}
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

func sqlCMDSearch(info *inputInfo) string {

	input := info.dat
	var text, author, date, authorMod, dateMod string
	var request string
	var err error

	// Change format of the command structure.
	if input.command == "info" {
		request = "author, date_added, author_mod, date_mod"
		switch input.length {
		case 1:
			err = db.QueryRow("SELECT (?) FROM commands WHERE command=(?)", request, input.args[1]).Scan(&author, &date, &authorMod, &dateMod)
		case 2:
			err = db.QueryRow("SELECT (?) FROM commands WHERE command=(?) AND arg1=(?)", request, input.args[1], input.args[2]).Scan(&author, &date, &authorMod, &dateMod)
		case 3:
			err = db.QueryRow("SELECT (?) FROM commands WHERE command=(?) AND arg1=(?) AND arg2=(?)", request, input.args[1], input.args[2], input.args[3]).Scan(&author, &date, &authorMod, &dateMod)
		default:
			return ""
		}

		if err != nil {
			errLog.Println(err)
			return ""
		}

		return fmt.Sprintf("Author: %s [%s], Last Modified: %s [%s]", author, date, authorMod, dateMod)
	}

	switch input.length {
	case 0:
		err = db.QueryRow("SELECT text FROM commands WHERE command=(?)", input.command).Scan(&text)
	case 1:
		err = db.QueryRow("SELECT text FROM commands WHERE command=(?) AND arg1=(?)", input.command, input.args[0]).Scan(&text)
	case 2:
		err = db.QueryRow("SELECT text FROM commands WHERE command=(?) AND arg1=(?) AND arg2=(?)", input.command, input.args[0], input.args[1]).Scan(&text)
	default:
		return ""
	}
	if err != nil {
		errLog.Println(err)
		return ""
	}

	if input.command == "script" {
		return sqlScriptGET(text)
	}

	return text
}

func sqlCMDEvent() string {
	now := time.Now()

	var dayAdd int
	switch now.Weekday().String() {
	case "Sunday":
		if now.Hour() < 12 {
			dayAdd = 0
		}
		dayAdd = 7
	case "Monday":
		dayAdd = 6
	case "Tuesday":
		dayAdd = 5
	case "Wednesday":
		dayAdd = 4
	case "Thursday":
		dayAdd = 3
	case "Friday":
		dayAdd = 2
	case "Saturday":
		dayAdd = 1
	}

	var dur time.Duration
	next := time.Date(now.Year(), now.Month(), now.Day()+dayAdd, 12, 0, 0, 0, now.Location())
	if dayAdd == 7 {
		dur = now.Sub(next)
	} else {
		dur = next.Sub(now)
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
	return fmt.Sprintf("`%0.f %s` and `%d %s` until `Sunday, 12:00 CST`", dur.Hours(), hourText, min, minText)
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

	// Check if it exists already
	var status bool
	fuckboy := input.args[0]
	err := db.QueryRow("SELECT status FROM blacklist WHERE name=(?)", fuckboy).Scan(&status)
	if err != nil && status == false {
		// Add to table
		_, err := db.Exec("INSERT INTO blacklist (name, status, times, start_date, who) VALUES (?, true, 1, Now(), ?)", fuckboy, userFull)
		if err != nil {
			discordLog.Println("Issue with init blacklisting user")
			return ""
		}
	} else {
		_, err := db.Exec("UPDATE blacklist SET times = times+1, start_date = Now(), who = (?), status = true WHERE name = (?)", userFull, fuckboy)
		if err != nil {
			discordLog.Println("Error updating blacklist")
			return ""
		}
	}
	return fmt.Sprintf("%s has blacklisted %s. Sucks to suck.", userFull, fuckboy)
}

func sqlCMDReport(info *inputInfo) string {

	user := info.user
	input := info.dat
	userFull := fmt.Sprintf("%s#%s", user.Username, user.Discriminator)
	fuckboy := input.args[0]
	var amount int

	if input.length != 1 {
		discordLog.Println("Not enough arguments")
		return ""
	}

	err := db.QueryRow("SELECT times FROM blacklist WHERE name=(?)", fuckboy).Scan(&amount)
	if err != nil && amount == 0 {
		// Add to table
		_, err := db.Exec("INSERT INTO blacklist (name, status, reports) VALUES (?, false, 1)", fuckboy)
		if err != nil {
			discordLog.Println("Issue with init blacklisting user")
			return ""
		}
	} else {
		_, err := db.Exec("UPDATE blacklist SET reports = reports+1 WHERE name = (?)", fuckboy)
		if err != nil {
			discordLog.Println("Error updating blacklist")
			return ""
		}
	}

	return fmt.Sprintf("%s has reported %s.", userFull, fuckboy)
}

func sqlScriptSET(name, text string) string {

	res, err := db.Exec("INSERT INTO library (name, script) VALUES (?, ?)", name, text)
	if err != nil {
		errLog.Println(err)
		return ""
	}
	lastID, err := res.LastInsertId()
	if err != nil {
		log.Fatal(err)
	}

	return fmt.Sprintf("%d", lastID)
}

func sqlScriptGET(id string) string {
	var script string
	err := db.QueryRow("SELECT script FROM library WHERE id=(?)", id).Scan(&script)
	if err != nil {
		errLog.Println(err)
		return ""
	}
	return script
}

func sqlScriptMOD(name, text string) bool {
	_, err := db.Exec("UPDATE library SET script=(?) WHERE name=(?)", text, name)
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
