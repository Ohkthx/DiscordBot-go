package bot

import (
	"fmt"
)

// Delete commands from selected tables
// Requires Granted permissions.
func (state *Instance) dbCommandDelete() (res *Response) {
	input := state.Cmd
	db := state.Database
	who := state.User
	whoFull := fmt.Sprintf("%s#%s", who.Username, who.Discriminator)
	var err error

	// Requires an input of at least 1.
	if input.Length < 1 {
		err = fmt.Errorf("not enough arguments to delete")
		res = makeResponse(err, err.Error(), "")
		return
	} else if state.dbUserPermissions(who.ID) == false {
		err = fmt.Errorf("you do not have permissions to do that")
		res = makeResponse(err, err.Error(), "")
		return
	}

	// Check if it exists already
	r := state.dbSearch(input.Length)
	if r.Err != nil {
		res = makeResponse(r.Err, "that command does not exist", "")
		return
	}

	var deleted string

	// Perform deletion.
	switch input.Length {
	case 1:
		_, err = db.Exec("DELETE FROM commands WHERE command=(?) AND arg1 IS NULL", input.Args[0])
		deleted = fmt.Sprintf("[%s deleted]: -> %s", whoFull, input.Args[0])
	case 2:
		if input.Attr&cmdEVENT == cmdEVENT || input.Attr&cmdSCRIPT == cmdSCRIPT {
			return state.dbProxyDelete()
		}
		_, err = db.Exec("DELETE FROM commands WHERE command=(?) AND arg1=(?) AND arg2 IS NULL", input.Args[0], input.Args[1])
		deleted = fmt.Sprintf("[%s deleted]: -> %s %s", whoFull, input.Args[0], input.Args[1])
	case 3:
		_, err = db.Exec("DELETE FROM commands WHERE command=(?) AND arg1=(?) AND arg2=(?) AND arg3 is NULL", input.Args[0], input.Args[1], input.Args[2])
		deleted = fmt.Sprintf("[%s deleted]: -> %s %s %s", whoFull, input.Args[0], input.Args[1], input.Args[2])
	default:
		err = fmt.Errorf("too many arguments")
		res = makeResponse(err, err.Error(), "")
		return
	}

	if err != nil {
		res = makeResponse(err, "", "")
		return
	}

	res = makeResponse(nil, "", deleted)
	return
}
