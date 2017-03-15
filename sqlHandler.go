package main

import (
	"fmt"
	"log"
	"time"

	"github.com/bwmarrin/discordgo"
)

func sqlCheckPerm(id string, grant bool) bool {
	query := fmt.Sprintf("SELECT id FROM permissions WHERE id=('%s')", id)
	qPrep, err := db.Prepare(query)
	if err != nil {
		errLog.Println(err)
		return false
	}

	var val string
	err = qPrep.QueryRow().Scan(&val)
	if err != nil {
		errLog.Println(err)
		return false
	}

	return true
}

func sqlCMDGrant(author *discordgo.User, channel string, input []string) string {

	user := fmt.Sprintf("%s#%s", author.Username, author.Discriminator)
	if input[0] != "grant" || len(input) != 2 {
		discordLog.Println("Could not grant permissions.")
		return ""
	} else if sqlCheckPerm(author.ID, true) != true {
		discordLog.Println(user + "(" + author.ID + ") attempted to grant permissions to " + input[1] + ".")
		return ""
	}

	// Find user, get ID
	addeeID := userFind(channel, input[1], true)
	if addeeID == "" {
		discordLog.Printf("User [%s] not found. Missing discriminator (#000)?\n", input[1])
		return ""
	}
	addeeUsername := fmt.Sprintf("%s", input[1])

	insertDat := fmt.Sprintf("INSERT INTO permissions (id, username, allow, date_added, accountable) VALUES ('%s', '%s', false, Now(), '%s')",
		addeeID, addeeUsername, user)

	insPrep, err := db.Prepare(insertDat)
	if err != nil {
		errLog.Println(err)
	}

	_, err = insPrep.Exec()
	if err != nil {
		errLog.Println(err)
	}

	return fmt.Sprintf("%s granted permissions to use `,add` by %s", addeeUsername, user)
}

func sqlCMDAdd(author *discordgo.User, input []string, text string, inputLen int) string {

	user := fmt.Sprintf("%s#%s", author.Username, author.Discriminator)

	if input[0] != "add" || len(input) < 3 {
		discordLog.Println("Bad add request")
		return ""
	} else if sqlCheckPerm(author.ID, false) != true {
		discordLog.Println(user + "(" + author.ID + ") attempted to add a command.")
		return ""
	}

	// Check if it exists already
	existing := sqlCMDSearch(input[1:])
	if existing != "" {
		discordLog.Println("Already exists in database")
		return ""
	}

	// Make the insert string
	var added string
	var err error

	switch inputLen {
	case 2:
		_, err = db.Exec("INSERT INTO commands (command, text, author, date_added, author_mod, date_mod) VALUES (?, ?, ?, Now(), ?, Now())", input[1], text, user, user)
		if len(text) > 40 {
			added = fmt.Sprintf("[Added: %s] %s -> %s...", user, input[1], text[0:40])
		} else {
			added = fmt.Sprintf("[Added: %s] %s -> %s", user, input[1], text)
		}
	case 3:
		if input[1] == "script" {
			id := sqlScriptSET(input[2], text)
			_, err = db.Exec("INSERT INTO commands (command, arg1, text, author, date_added, author_mod, date_mod) VALUES (?, ?, ?, ?, Now(), ?, Now())", input[1], input[2], id, user, user)
		} else {
			_, err = db.Exec("INSERT INTO commands (command, arg1, text, author, date_added, author_mod, date_mod) VALUES (?, ?, ?, ?, Now(), ?, Now())", input[1], input[2], text, user, user)
		}
		if len(text) > 40 {
			added = fmt.Sprintf("[Added: %s] %s %s -> %s...", user, input[1], input[2], text[0:40])
		} else {
			added = fmt.Sprintf("[Added: %s] %s %s -> %s", user, input[1], input[2], text)
		}
	case 4:
		_, err = db.Exec("INSERT INTO commands (command, arg1, arg2, text, author, date_added, author_mod, date_mod) VALUES (?, ?, ?, ?, ?, Now(), ?, Now())", input[1], input[2], input[3], text, user, user)
		if len(text) > 40 {
			added = fmt.Sprintf("[Added: %s] %s %s %s -> %s...", user, input[1], input[2], input[3], text[0:40])
		} else {
			added = fmt.Sprintf("[Added: %s] %s %s %s -> %s", user, input[1], input[2], input[3], text)
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

func sqlCMDSearch(input []string) string {

	var text, author, date, authorMod, dateMod string
	var request, query string
	var err error

	// Change format of the command structure.
	if input[0] == "info" {
		request = "author, date_added, author_mod, date_mod"

		switch len(input) {
		case 2:
			query = fmt.Sprintf("SELECT %s FROM commands WHERE command=('%s')", request, input[1])
		case 3:
			query = fmt.Sprintf("SELECT %s FROM commands WHERE command=('%s') AND arg1=('%s')", request, input[1], input[2])
		case 4:
			query = fmt.Sprintf("SELECT %s FROM commands WHERE command=('%s') AND arg1=('%s') AND arg2=('%s')", request, input[1], input[2], input[3])
		default:
			return ""
		}

		qPrep, err := db.Prepare(query)
		if err != nil {
			errLog.Println(err)
		}

		err = qPrep.QueryRow().Scan(&author, &date, &authorMod, &dateMod)
		if err != nil {
			errLog.Println(err)
			return ""
		}

		return fmt.Sprintf("Author: %s [%s], Last Modified: %s [%s]", author, date, authorMod, dateMod)

	}

	switch len(input) {
	case 1:
		query = fmt.Sprintf("SELECT text FROM commands WHERE command=('%s')", input[0])
	case 2:
		query = fmt.Sprintf("SELECT text FROM commands WHERE command=('%s') AND arg1=('%s')", input[0], input[1])
	case 3:
		query = fmt.Sprintf("SELECT text FROM commands WHERE command=('%s') AND arg1=('%s') AND arg2=('%s')", input[0], input[1], input[2])
	default:
		return ""
	}

	qPrep, err := db.Prepare(query)
	if err != nil {
		errLog.Println(err)
	}

	err = qPrep.QueryRow().Scan(&text)
	if err != nil {
		errLog.Println(err)
		return ""
	}

	if input[0] == "script" {
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
	err := db.QueryRow("SELECT script FROM library WHERE (?)", id).Scan(&script)
	if err != nil {
		errLog.Println(err)
		return ""
	}
	return script
}
