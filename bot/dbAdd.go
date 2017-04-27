package bot

import (
	"errors"
	"fmt"
)

// Grants the person the ability to manipulate tables.
// Potentially very dangerous to do!
func (state *Instance) dbUserPermissionsAdd() (res *Response) {
	input := state.Cmd
	db := state.Database
	who := state.User
	whoFull := fmt.Sprintf("%s#%s", who.Username, who.Discriminator)

	// Check length, minimum request is: grant [username]
	if input.Length != 1 {
		err := fmt.Errorf("not enough arguments. Want: grant [username#1234]")
		res = makeResponse(err, err.Error(), "")
		return
	} else if state.dbUserPermissions(who.ID) == false {
		err := fmt.Errorf("you do not have permissions to do that")
		res = makeResponse(err, err.Error(), "")
		return
	}

	if state.Channel.ID == "" {
		err := errors.New("unexpected channel error while granting permissions")
		res = makeResponse(err, err.Error(), "")
		return
	}

	// Find user, get ID. Return on bad ID
	addee, err := state.UserFind(input.Args[0])
	if err != nil {
		res = makeResponse(err, "", "")
		return
	}

	addeeFull := fmt.Sprintf("%s#%s", addee.Username, addee.Discriminator)

	// Make SQL request to grant user ability to manipulate others.
	_, err = db.Exec("INSERT INTO permissions (id, username, allow, date_added, accountable) VALUES (?, ?, false, Now(), ?)",
		addee.ID, addeeFull, whoFull)
	if err != nil {
		res = makeResponse(err, "", "")
		return
	}

	res = makeResponse(nil, "", fmt.Sprintf("%s granted permissions to use `,add` by %s", addeeFull, whoFull))
	return
}

// Add commands to selected tables
// Requires Granted permissions.
func (state *Instance) dbCommandAdd() (res *Response) {
	db := state.Database
	input := state.Cmd
	who := state.User
	whoFull := fmt.Sprintf("%s#%s", who.Username, who.Discriminator)

	// Minimum length 2:  [command] [text]
	if input.Length < 2 {
		err := errors.New("not enough arguments. Require at least a command and text")
		res = makeResponse(err, err.Error(), "")
		return
	} else if state.dbUserPermissions(who.ID) == false {
		err := errors.New("you do not have permissions to do that")
		res = makeResponse(err, err.Error(), "")
		return
	}

	// Check if request already exists.
	r := state.dbSearch(input.Length - 1)
	if r.Err == nil {
		res = makeResponse(r.Err, "command already exists. Did you mean to modify?", "")
		return
	}

	var added string
	var err error

	// Make inserts into tables.
	switch input.Length {
	case 2:
		_, err = db.Exec("INSERT INTO commands (command, text, author, date_added, author_mod, date_mod) VALUES (?, ?, ?, Now(), ?, Now())", input.Args[0], input.Text, whoFull, whoFull)
		if len(input.Text) > 40 {
			added = fmt.Sprintf("[Added: %s] %s -> %s...", whoFull, input.Args[0], input.Text[0:40])
		} else {
			added = fmt.Sprintf("[Added: %s] %s -> %s", whoFull, input.Args[0], input.Text)
		}
	case 3:
		if input.Attr&cmdEVENT == cmdEVENT || input.Attr&cmdSCRIPT == cmdSCRIPT {
			return state.dbProxyAdd()
		}
		_, err = db.Exec("INSERT INTO commands (command, arg1, text, author, date_added, author_mod, date_mod) VALUES (?, ?, ?, ?, Now(), ?, Now())", input.Args[0], input.Args[1], input.Text, whoFull, whoFull)
		if len(input.Text) > 40 {
			added = fmt.Sprintf("[Added: %s] %s %s -> %s...", whoFull, input.Args[0], input.Args[1], input.Text[0:40])
		} else {
			added = fmt.Sprintf("[Added: %s] %s %s -> %s", whoFull, input.Args[0], input.Args[1], input.Text)
		}
	case 4:
		_, err = db.Exec("INSERT INTO commands (command, arg1, arg2, text, author, date_added, author_mod, date_mod) VALUES (?, ?, ?, ?, ?, Now(), ?, Now())", input.Args[0], input.Args[1], input.Args[2], input.Text, whoFull, whoFull)
		if len(input.Text) > 40 {
			added = fmt.Sprintf("[Added: %s] %s %s %s -> %s...", whoFull, input.Args[0], input.Args[1], input.Args[2], input.Text[0:40])
		} else {
			added = fmt.Sprintf("[Added: %s] %s %s %s -> %s", whoFull, input.Args[0], input.Args[1], input.Args[2], input.Text)
		}
	default:
		err = fmt.Errorf("too many arguments")
		res = makeResponse(err, err.Error(), "")
		return
	}

	// Handle any issues with inserting data.
	if err != nil {
		res = makeResponse(err, "", "")
		return
	}

	res = makeResponse(nil, "", added)
	return
}

func (state *Instance) dbChannelAdd(cID string, total int64) (err error) {
	db := state.Database
	_, err = db.Exec("UPDATE channels SET amount=(?) WHERE id=(?)", total, cID)
	if err != nil {
		return err
	}
	return nil
}
