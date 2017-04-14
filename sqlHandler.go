package main

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Check the permissions of a USER by ID to verify if they
// are allowed to manipulate other tables.
func sqlCheckPerm(id string) bool {
	var val sql.NullString
	// Grab ID from table (sure indicator that it is a valid person)
	err := db.QueryRow("SELECT id FROM permissions WHERE id=(?)", id).Scan(&val)
	if err != nil {
		errLog.Println("checking permissions", err)
		return false
		// If information is not NULL
	} else if val.Valid {
		return true
	}

	return false
}

// Grants the person the ability to manipulate tables.
// Potentially very dangerous to do!
func sqlCMDGrant(info *inputInfo) (string, error) {
	input := info.dat
	who := info.user
	whoFull := fmt.Sprintf("%s#%s", who.Username, who.Discriminator)
	var err error

	// Check length, minimum request is: grant [username]
	if input.length != 1 {
		err = errors.New("not enough arguments. Want: grant [username#1234]")
		return "", err
	} else if sqlCheckPerm(who.ID) == false {
		err = errors.New("you do not have permissions to do that")
		return "", err
	}

	if info.channel.ID == "" {
		err = errors.New("unexpected channel error while granting permissions")
		return "", err
	}

	// Find user, get ID. Return on bad ID
	addee, err := userFind(info, input.args[0])
	if err != nil {
		return "", err
	}

	addeeFull := fmt.Sprintf("%s#%s", addee.Username, addee.Discriminator)

	// Make SQL request to grant user ability to manipulate others.
	_, err = db.Exec("INSERT INTO permissions (id, username, allow, date_added, accountable) VALUES (?, ?, false, Now(), ?)",
		addee.ID, addeeFull, whoFull)
	if err != nil {
		errLog.Printf("Error inserting [%s] into database for Grant permissions.\n", addeeFull)
		err = errors.New("could not grant permissions")
		return "", err
	}

	return fmt.Sprintf("%s granted permissions to use `,add` by %s", addeeFull, whoFull), nil
}

// Add commands to selected tables
// Requires Granted permissions.
func sqlCMDAdd(info *inputInfo) (string, error) {
	input := info.dat
	who := info.user
	whoFull := fmt.Sprintf("%s#%s", who.Username, who.Discriminator)
	var err error

	// Minimum length 2:  [command] [text]
	if input.length < 2 {
		err = errors.New("not enough arguments. Require at least a command and text")
		return "", err
	} else if sqlCheckPerm(who.ID) == false {
		err = errors.New("you do not have permissions to do that")
		return "", err
	}

	// Check if request already exists.
	_, err = sqlCMDSearch(input, input.length-1)
	if err == nil {
		err = errors.New("command already exists. Did you mean to modify?")
		return "", err
	}

	var added string

	// Make inserts into tables.
	switch input.length {
	case 2:
		_, err = db.Exec("INSERT INTO commands (command, text, author, date_added, author_mod, date_mod) VALUES (?, ?, ?, Now(), ?, Now())", input.args[0], input.text, whoFull, whoFull)
		if len(input.text) > 40 {
			added = fmt.Sprintf("[Added: %s] %s -> %s...", whoFull, input.args[0], input.text[0:40])
		} else {
			added = fmt.Sprintf("[Added: %s] %s -> %s", whoFull, input.args[0], input.text)
		}
	case 3:
		if input.attr&cmdEVENT == cmdEVENT || input.attr&cmdSCRIPT == cmdSCRIPT {
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
		err = errors.New("too many arguments")
		return "", err
	}

	// Handle any issues with inserting data.
	if err != nil {
		errLog.Println("Issues adding commands", err)
		err = errors.New("unexpected issue adding commands")
		return "", err
	}

	return added, nil
}

// Delete commands from selected tables
// Requires Granted permissions.
func sqlCMDDel(info *inputInfo) (string, error) {
	input := info.dat
	who := info.user
	whoFull := fmt.Sprintf("%s#%s", who.Username, who.Discriminator)
	var err error

	// Requires an input of at least 1.
	if input.length < 1 {
		err = errors.New("not enough arguments to delete")
		return "", err
	} else if sqlCheckPerm(who.ID) == false {
		err = errors.New("you do not have permissions to do that")
		return "", err
	}

	// Check if it exists already
	_, err = sqlCMDSearch(input, input.length)
	if err != nil {
		err = errors.New("that command does not exist")
		return "", err
	}

	var deleted string

	// Perform deletion.
	switch input.length {
	case 1:
		_, err = db.Exec("DELETE FROM commands WHERE command=(?) AND arg1 IS NULL AND author=(?)", input.args[0], whoFull)
		deleted = fmt.Sprintf("[%s deleted]: -> %s", whoFull, input.args[0])
	case 2:
		if input.attr&cmdEVENT == cmdEVENT || input.attr&cmdSCRIPT == cmdSCRIPT {
			return sqlProxyDel(info)
		}
		_, err = db.Exec("DELETE FROM commands WHERE command=(?) AND arg1=(?) AND arg2 IS NULL AND author=(?)", input.args[0], input.args[1], whoFull)
		deleted = fmt.Sprintf("[%s deleted]: -> %s %s", whoFull, input.args[0], input.args[1])
	case 3:
		_, err = db.Exec("DELETE FROM commands WHERE command=(?) AND arg1=(?) AND arg2=(?) AND author=(?)", input.args[0], input.args[1], input.args[0], whoFull)
		deleted = fmt.Sprintf("[%s deleted]: -> %s %s %s", whoFull, input.args[0], input.args[1], input.args[2])
	default:
		err = errors.New("too many arguments")
		return "", err
	}

	if err != nil {
		errLog.Println("Issues deleting commands", err)
		err = errors.New("unexpected issue deleting commands")
		return "", err
	}

	return deleted, nil
}

// Modify and existing command
// Requires Granted permissions.
func sqlCMDMod(info *inputInfo) (string, error) {
	input := info.dat
	who := info.user
	whoFull := fmt.Sprintf("%s#%s", who.Username, who.Discriminator)
	var err error

	// Minimum of 2 commands required.
	if input.length < 2 {
		err = errors.New("not enough arguments. want: [mod] [command(s)] [text]")
		return "", err
	} else if sqlCheckPerm(who.ID) == false {
		err = errors.New("you do not have permissions to do that")
		return "", err
	}

	// Check if it exists already
	_, err = sqlCMDSearch(input, input.length-1)
	if err != nil {
		err = errors.New("that command does not exist")
		return "", err
	}

	var modified string
	switch input.length {
	case 2:
		_, err = db.Exec("UPDATE commands SET text=(?), author_mod=(?), date_mod=Now() WHERE command=(?) arg1 IS NULL AND author=(?)", input.text, whoFull, input.args[0], whoFull)
		modified = fmt.Sprintf("[%s updated]: -> %s", whoFull, input.args[0])
	case 3:
		if input.attr&cmdEVENT == cmdEVENT || input.attr&cmdSCRIPT == cmdSCRIPT {
			return sqlProxyMod(info)
		}
		_, err = db.Exec("UPDATE commands SET text=(?), author_mod=(?), date_mod=Now() WHERE command=(?) AND arg1=(?) AND arg2 IS NULL AND author=(?)", input.text, whoFull, input.args[0], input.args[1], whoFull)
		modified = fmt.Sprintf("[%s updated]: -> %s %s", whoFull, input.args[0], input.args[1])
	case 4:
		_, err = db.Exec("UPDATE commands SET text=(?), author_mod=(?), date_mod=Now() WHERE command=(?) AND arg1=(?) AND author=(?)", input.text, whoFull, input.args[0], input.args[1], input.args[2], whoFull)
		modified = fmt.Sprintf("[%s updated]: -> %s %s %s", whoFull, input.args[0], input.args[1], input.args[2])
	default:
		errLog.Println("too many arguments.")
		return "", err
	}

	if err != nil {
		errLog.Println("Issues modifying commands", err)
		err = errors.New("unexpected issue modifying commands")
		return "", err
	}

	return modified, nil
}

// Attempt to find and return a command.
func sqlCMDSearch(input *inputDat, length int) (string, error) {
	var err error
	var text sql.NullString
	i := input.args

	if input.command == "help" {
		return sqlCMDHelp(input.args)
	} else if modifierSet(input) == false {
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
		err = errors.New("too many arguments")
		return "", err
	}
	if err != nil {
		if err == sql.ErrNoRows {
			err = errors.New("command not found")
			return "", err
		}
		errLog.Println("searching table", err)
		err = errors.New("issue looking up command")
		return "", err
	}

	if input.command == "script" {
		text, err := sqlProxyLinkGET(input.command, text.String)
		return text[0], err
	}

	if text.Valid {
		return text.String, nil
	}
	err = errors.New("invalid command")
	return "", err
}

// Compares current time to that which is stored and returns
// the result for all that are described in the table.
func sqlCMDEvent() (string, error) {

	var weekday, hhmmFull, retText string
	var hh, mm, cnt int
	var err error

	now := time.Now()
	retText = "Events Coming Soon\n```"

	rows, err := db.Query("SELECT weekday, time FROM events")
	if err != nil {
		errLog.Println("Event lookup (getting rows): ", err)
		err = errors.New("could not get events")
		return "", err
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&weekday, &hhmmFull)
		if err != nil {
			errLog.Println("Event lookup (proc rows): ", err)
			err = errors.New("no events found?")
			return "", err
		}
		hhmm := strings.Split(hhmmFull, ":")
		hh, err = strconv.Atoi(hhmm[0])
		if err != nil {
			errLog.Println("event hour conv: ", err)
			err = errors.New("could not convert hours")
			return "", err
		}
		mm, err = strconv.Atoi(hhmm[1])
		if err != nil {
			errLog.Println("event min conv: ", err)
			err = errors.New("could not conver minutes")
			return "", err
		}
		// DO SOME STUFF WITH EVENT
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

		if *debug {
			fmt.Printf("num: %d\talt: %d\tdayAdd: %d\n", num, alt, dayAdd)
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
		errLog.Println("Event lookup (unknown): ", err)
		err = errors.New("no events found? (unknown)")
		return "", err
	}

	if cnt < 1 {
		return "`No events found.`", nil
	}

	retText = strings.Trim(retText, " \n")
	return retText + "```check #events", nil
}

// Looks up a command returns all other commands with the same base
func sqlCMDHelp(input []string) (string, error) {
	var retStr string
	var arg0, arg1, arg2 sql.NullString
	var err error

	cnt := 0
	retStr = fmt.Sprintf("%s help: ```\n", input[0])

	rows, err := db.Query("SELECT command, arg1, arg2 FROM commands WHERE command=(?)", input[0])
	if err != nil {
		errLog.Printf("%s lookup (getting rows): %s", input[0], err)
		err = errors.New("could not retrieve information")
		return "", err
	}
	defer rows.Close()

	for rows.Next() {
		var str0, str1, str2 string
		err := rows.Scan(&arg0, &arg1, &arg2)
		if err != nil {
			errLog.Printf("%s lookup (proc rows): %s", input[0], err)
			err = errors.New("issue processing information")
			return "", err
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
		errLog.Printf("%s lookup (unknown): %s", input[0], err)
		err = errors.New("unknown issue")
		return "No help :'(", err
	}

	// Prevent returning ugly text.
	if cnt < 1 {
		return "No help for you. :'(", nil
	}
	retText := strings.Trim(retStr, " \n")
	return retText + "```", nil
}

// Add a to main table and adds to the linked table.
func sqlProxyAdd(info *inputInfo) (string, error) {
	input := info.dat
	who := info.user
	whoFull := fmt.Sprintf("%s#%s", who.Username, who.Discriminator)
	var err error

	/*
		Could add additional error checking here to see if exist.
		All calling functions already make this check tho.
	*/

	id, err := sqlProxyLinkSET(input.args, input.text)
	if err != nil {
		return "", err
	}

	_, err = db.Exec("INSERT INTO commands (command, arg1, text, author, date_added, author_mod, date_mod) VALUES (?, ?, ?, ?, Now(), ?, Now())", input.args[0], input.args[1], id, whoFull, whoFull)
	if err != nil {
		errLog.Println("adding with proxy", err)
		err = errors.New("unable to add command")
		return "", err
	}

	if len(input.text) > 40 {
		return fmt.Sprintf("[Added: %s] %s %s -> %s...", whoFull, input.args[0], input.args[1], input.text[0:10]), nil
	}
	return fmt.Sprintf("[Added: %s] %s %s -> %s", whoFull, input.args[0], input.args[1], input.text), nil
}

// Modifies the main table with updates to the linked table as well.
func sqlProxyMod(info *inputInfo) (string, error) {
	input := info.dat
	who := info.user
	whoFull := fmt.Sprintf("%s#%s", who.Username, who.Discriminator)
	var err error

	/*
		Could add additional error checking here to see if exist.
		All calling functions already make this check tho.
	*/

	err = sqlProxyLinkMOD(input.args)
	if err != nil {
		return "", err
	}

	// Perform updating of the main table
	_, err = db.Exec("UPDATE commands SET text=(?), author_mod=(?), date_mod=Now() WHERE command=(?) AND arg1=(?) AND arg2 IS NULL AND author=(?)", input.text, whoFull, input.args[0], input.args[1], whoFull)
	if err != nil {
		errLog.Println("modifying with proxy", err)
		err = errors.New("unable to modify command")
		return "", err
	}

	return fmt.Sprintf("[%s updated]: -> %s %s", whoFull, input.args[0], input.args[1]), nil
}

// Responsible for deleting entry and the table it links too.
func sqlProxyDel(info *inputInfo) (string, error) {
	input := info.dat
	who := info.user
	whoFull := fmt.Sprintf("%s#%s", who.Username, who.Discriminator)
	var err error

	/*
		Could add additional error checking here to see if exist.
		All calling functions already make this check tho.
	*/

	// Remove the proxy, may need to rearrange these. I don't want to get rid of link first tho.
	err = sqlProxyLinkDEL(input.args)
	if err != nil {
		return "", err
	}

	// Perform deletion
	_, err = db.Exec("DELETE FROM commands WHERE command=(?) AND arg1=(?) AND arg2 IS NULL AND author=(?)", input.args[0], input.args[1], whoFull)
	if err != nil {
		errLog.Println("deleting with proxy", err)
		err = errors.New("unable to delete command")
		return "", err
	}

	return fmt.Sprintf("[%s deleted]: -> %s %s", whoFull, input.args[0], input.args[1]), nil
}

// Blacklists a player and prevents them from using.
// This could be for a number of things such as,
// Bot abuse, in-service abuse, etc
func sqlCMDBlacklist(info *inputInfo) (string, error) {
	user := info.user
	input := info.dat
	userFull := fmt.Sprintf("%s#%s", user.Username, user.Discriminator)
	var err error

	if input.length != 1 {
		err = errors.New("not enough arguments")
		return "", err
	} else if sqlCheckPerm(user.ID) == false {
		err = errors.New("you do not have permissions to do that")
		return "", err
	}

	// Find user, get ID. Return on bad ID
	reported, err := userFind(info, input.args[0])
	if err != nil {
		return "", err
	}

	criminal := fmt.Sprintf("%s#%s", reported.Username, reported.Discriminator)
	// Check if it exists already
	var status sql.NullBool
	err = db.QueryRow("SELECT status FROM blacklist WHERE name=(?)", criminal).Scan(&status)
	if err != nil {
		if err == sql.ErrNoRows {
			// Add to table
			_, err = db.Exec("INSERT INTO blacklist (name, status, times, start_date, who) VALUES (?, true, 1, Now(), ?)", criminal, userFull)
			if err != nil {
				errLog.Println("issue inserting new blacklistee", err)
				err = errors.New("could not add new blacklistee")
				return "", err
			}
			// Rows were found, another issue occured.
			errLog.Println("getting reports", err)
			err = errors.New("could not report user")
			return "", err
		}
	}
	if status.Valid && status.Bool == false {
		_, err = db.Exec("UPDATE blacklist SET times = times+1, start_date = Now(), who = (?), status = true WHERE name = (?)", userFull, criminal)
		if err != nil {
			errLog.Println("error updating with new blacklistee")
			err = errors.New("could not blacklist user")
			return "", err
		}
	}
	return fmt.Sprintf("%s has blacklisted %s. Sucks to suck.", userFull, criminal), nil
}

// Updates a counter for amount of reports on a player.
// Adds a new entry if it is not existant.
func sqlCMDReport(info *inputInfo) (string, error) {
	user := info.user
	input := info.dat
	userFull := fmt.Sprintf("%s#%s", user.Username, user.Discriminator)
	var amount sql.NullInt64
	var err error

	reportUser, err := userFind(info, input.args[0])
	if err != nil {
		return "", err
	}

	criminal := fmt.Sprintf("%s#%s", reportUser.Username, reportUser.Discriminator)

	if input.length != 1 {
		err = errors.New("not enough arguments. want: [report] [username#1234]")
		return "", err
	}

	// Get info for amount of times previously reported. If 0, add.
	err = db.QueryRow("SELECT times FROM blacklist WHERE name=(?)", criminal).Scan(&amount)
	if err != nil {
		// If rows are not found. Insert into table.
		if err == sql.ErrNoRows {
			_, err := db.Exec("INSERT INTO blacklist (name, status, reports) VALUES (?, false, 1)", criminal)
			if err != nil {
				errLog.Println("issue with init reporting user", err)
				err = errors.New("could not report user")
				return "", err
			}
		}
		// Rows were found, another issue occured.
		errLog.Println("getting reports", err)
		err = errors.New("could not report user")
		return "", err
	}
	// Update the tables counter for reporting.
	if amount.Valid && amount.Int64 > 0 {
		_, err := db.Exec("UPDATE blacklist SET reports = reports+1 WHERE name = (?)", criminal)
		if err != nil {
			errLog.Println("updating report", err)
			err = errors.New("could not report user")
			return "", err
		}
	}

	return fmt.Sprintf("%s has reported %s.", userFull, criminal), nil
}

// Create a new link (may replace with ADD)
func sqlProxyLinkSET(args []string, text string) (string, error) {
	var err error
	var res sql.Result

	switch args[0] {
	case "script":
		res, err = db.Exec("INSERT INTO library (name, script) VALUES (?, ?)", args[1], text)
	case "event":
		res, err = db.Exec("INSERT INTO events (weekday, time) VALUES (?, ?)", args[1], text)
	default:
		err = errors.New("not option found for setting link")
		return "", err
	}
	if err != nil {
		errLog.Println("could not add link", err)
		err = errors.New("could not add to database")
		return "", err
	}
	lastID, err := res.LastInsertId()
	if err != nil {
		errLog.Println("could not get ID for new link", err)
		err = errors.New("could not add")
		return "", err
	}

	return fmt.Sprintf("%d", lastID), nil
}

// Get ID of foreign coloum
func sqlProxyLinkGET(command, id string) ([2]string, error) {
	var info1, info2 sql.NullString
	var strs [2]string
	var err error

	switch command {
	case "script":
		err = db.QueryRow("SELECT script FROM library WHERE id=(?)", id).Scan(&info1)
	case "event":
		err = db.QueryRow("SELECT weekday, time FROM events WHERE id=(?)", id).Scan(&info1, &info2)
	default:
		err = errors.New("bad request")
		return strs, err
	}
	if err != nil {
		if err == sql.ErrNoRows {
			err = errors.New("command doesn't exist")
			return strs, err
		}
		errLog.Println("getting link", err)
		err = errors.New("request failed")
		return strs, err
	}

	if info1.Valid && info2.Valid {
		strs[0] = info1.String
		strs[1] = info2.String
		return strs, nil
	} else if info1.Valid && info2.Valid == false {
		strs[0] = info1.String
		return strs, nil
	} else if info1.Valid == false && info2.Valid {
		strs[1] = info2.String
		return strs, err
	}

	err = errors.New("results not found?")
	return strs, err
}

// Modify foreign table
func sqlProxyLinkMOD(info []string) error {
	var err error
	switch info[0] {
	case "script":
		_, err = db.Exec("UPDATE library SET script=(?) WHERE name=(?)", info[2], info[1])
	case "event":
		_, err = db.Exec("UPDATE events SET time=(?) WHERE weekday=(?)", info[2], info[1])
	default:
		err = errors.New("option not found for modifying link")
		return err
	}

	if err != nil {
		errLog.Println("modifying link", err)
		return err
	}

	return nil
}

// Delete foreign table
func sqlProxyLinkDEL(info []string) error {
	var err error
	switch info[0] {
	case "script":
		_, err = db.Exec("DELETE FROM library WHERE name=(?)", info[1])
	case "event":
		_, err = db.Exec("DELETE FROM events WHERE weekday=(?)", info[1])
	default:
		err = errors.New("option not found for deleting link")
		return err
	}

	if err != nil {
		errLog.Println("deleting link", err)
		return err
	}

	return nil
}

// Check if user is blacklisted from using the bot
func sqlBlacklistGET(input string) bool {
	var status sql.NullBool
	err := db.QueryRow("SELECT status FROM blacklist WHERE name=(?)", input).Scan(&status)
	if err != nil {
		// If row doesn't exist, return false
		if err == sql.ErrNoRows {
			return false
		}
		errLog.Println("getting blacklistings", err)
		return false
	} else if status.Valid && status.Bool {
		return true
	}
	return false
}
