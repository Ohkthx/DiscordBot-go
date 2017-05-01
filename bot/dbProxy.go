package bot

import (
	"database/sql"
	"errors"
	"fmt"
)

// Add a to main table and adds to the linked table.
func (state *Instance) dbProxyAdd() (res *Response) {
	var err error
	db := state.Database
	input := state.Cmd
	who := state.User
	whoFull := fmt.Sprintf("%s#%s", who.Username, who.Discriminator)

	/*
		Could add additional error checking here to see if exist.
		All calling functions already make this check tho.
	*/

	res = state.dbProxyLinkAdd(input.Args, input.Text)
	if res.Err != nil {
		return
	}

	_, err = db.Exec("INSERT INTO commands (command, arg1, text, author, date_added, author_mod, date_mod) VALUES (?, ?, ?, ?, Now(), ?, Now())", input.Args[0], input.Args[1], res.Sndmsg, whoFull, whoFull)
	if err != nil {
		res = makeResponse(err, "Unable to add command", "")
		return
	}

	if len(input.Text) > 40 {
		res = makeResponse(nil, "", fmt.Sprintf("[Added: %s] %s %s -> %s...", whoFull, input.Args[0], input.Args[1], input.Text[0:10]))
		return
	}
	res = makeResponse(nil, "", fmt.Sprintf("[Added: %s] %s %s -> %s", whoFull, input.Args[0], input.Args[1], input.Text))
	return
}

// Modifies the main table with updates to the linked table as well.
func (state *Instance) dbProxyModify() (res *Response) {
	db := state.Database
	input := state.Cmd
	who := state.User
	whoFull := fmt.Sprintf("%s#%s", who.Username, who.Discriminator)
	var err error

	/*
		Could add additional error checking here to see if exist.
		All calling functions already make this check tho.
	*/

	res = state.dbProxyLinkModify(input.Args)
	if res.Err != nil {
		return
	}

	// Perform updating of the main table
	_, err = db.Exec("UPDATE commands SET text=(?), author_mod=(?), date_mod=Now() WHERE command=(?) AND arg1=(?) AND arg2 IS NULL AND author=(?)", input.Text, whoFull, input.Args[0], input.Args[1], whoFull)
	if err != nil {
		res = makeResponse(err, "Unable to modify", "")
		return
	}

	res = makeResponse(nil, "", fmt.Sprintf("[%s updated]: -> %s %s", whoFull, input.Args[0], input.Args[1]))
	return
}

// Responsible for deleting entry and the table it links too.
func (state *Instance) dbProxyDelete() (res *Response) {
	db := state.Database
	input := state.Cmd
	who := state.User
	whoFull := fmt.Sprintf("%s#%s", who.Username, who.Discriminator)
	var err error

	/*
		Could add additional error checking here to see if exist.
		All calling functions already make this check tho.
	*/

	// Remove the proxy, may need to rearrange these. I don't want to get rid of link first tho.
	res = state.dbProxyLinkDelete(input.Args)
	if res.Err != nil {
		return
	}

	// Perform deletion
	_, err = db.Exec("DELETE FROM commands WHERE command=(?) AND arg1=(?) AND arg2 IS NULL", input.Args[0], input.Args[1])
	if err != nil {
		res = makeResponse(err, "Unable to delete command", "")
		return
	}

	res = makeResponse(nil, "", fmt.Sprintf("[%s deleted]: -> %s %s", whoFull, input.Args[0], input.Args[1]))
	return
}

// Create a new link (may replace with ADD)
func (state *Instance) dbProxyLinkAdd(args []string, text string) (res *Response) {
	var err error
	var r sql.Result
	db := state.Database

	switch args[0] {
	case "script":
		r, err = db.Exec("INSERT INTO library (name, script) VALUES (?, ?)", args[1], text)
	case "event":
		r, err = db.Exec("INSERT INTO events (weekday, time) VALUES (?, ?)", args[1], text)
	default:
		err = errors.New("not option found for setting link")
		res = makeResponse(err, err.Error(), "")
		return
	}
	if err != nil {
		res = makeResponse(err, "Could not add to database", "")
		return
	}
	lastID, err := r.LastInsertId()
	if err != nil {
		res = makeResponse(err, "", "")
		return
	}

	res = makeResponse(nil, "", fmt.Sprintf("%d", lastID))
	return
}

// Get ID of foreign coloum
func (state *Instance) dbProxyLinkGet(command, id string) (strs [2]string, err error) {
	var info1, info2 sql.NullString
	db := state.Database

	switch command {
	case "script":
		err = db.QueryRow("SELECT script FROM library WHERE id=(?)", id).Scan(&info1)
	case "event":
		err = db.QueryRow("SELECT weekday, time FROM events WHERE id=(?)", id).Scan(&info1, &info2)
	default:
		err = errors.New("bad request")
		return
	}
	if err != nil {
		if err == sql.ErrNoRows {
			err = errors.New("command doesn't exist")
			return
		}
		err = errors.New("request failed")
		return
	}

	if info1.Valid && info2.Valid {
		strs[0] = info1.String
		strs[1] = info2.String
		return
	} else if info1.Valid && info2.Valid == false {
		strs[0] = info1.String
		return
	} else if info1.Valid == false && info2.Valid {
		strs[1] = info2.String
		return
	}

	err = errors.New("results not found?")
	return
}

// Modify foreign table
func (state *Instance) dbProxyLinkModify(info []string) (res *Response) {
	db := state.Database
	var err error

	switch info[0] {
	case "script":
		_, err = db.Exec("UPDATE library SET script=(?) WHERE name=(?)", info[2], info[1])
	case "event":
		_, err = db.Exec("UPDATE events SET time=(?) WHERE weekday=(?)", info[2], info[1])
	default:
		err = errors.New("option not found for modifying link")
		res = makeResponse(err, err.Error(), "")
		return
	}

	if err != nil {
		res = makeResponse(err, "", "")
		return
	}

	return
}

// Delete foreign table
func (state *Instance) dbProxyLinkDelete(info []string) (res *Response) {
	var err error
	db := state.Database
	switch info[0] {
	case "script":
		_, err = db.Exec("DELETE FROM library WHERE name=(?)", info[1])
	case "event":
		_, err = db.Exec("DELETE FROM events WHERE weekday=(?)", info[1])
	default:
		err = errors.New("option not found for deleting link")
		res = makeResponse(err, err.Error(), "")
		return
	}

	if err != nil {
		res = makeResponse(err, "", "")
		return
	}

	res = makeResponse(nil, "", "")
	return
}
