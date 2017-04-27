package bot

import (
	"database/sql"
	"errors"
	"fmt"
)

// Check the permissions of a USER by ID to verify if they
// are allowed to manipulate other tables.
func (state *Instance) dbUserPermissions(id string) bool {
	db := state.Database
	var val sql.NullString
	// Grab ID from table (sure indicator that it is a valid person)
	err := db.QueryRow("SELECT id FROM permissions WHERE id=(?)", id).Scan(&val)
	if err != nil {
		return false
		// If information is not NULL
	} else if val.Valid {
		return true
	}

	return false
}

// Attempt to find and return a command.
func (state *Instance) dbSearch(length int) (res *Response) {
	var text sql.NullString
	var err error
	input := state.Cmd
	db := state.Database
	i := input.Args

	if input.Command == "help" {
		return state.dbHelp(input.Args)
	} else if state.modifierSet() == false {
		i = state.cmdconv()
	}

	switch length {
	case 1:
		err = db.QueryRow("SELECT text FROM commands WHERE command=(?) AND arg1 IS NULL", i[0]).Scan(&text)
	case 2:
		err = db.QueryRow("SELECT text FROM commands WHERE command=(?) AND arg1=(?) AND arg2 IS NULL", i[0], i[1]).Scan(&text)
	case 3:
		err = db.QueryRow("SELECT text FROM commands WHERE command=(?) AND arg1=(?) AND arg2=(?)", i[0], i[1], i[2]).Scan(&text)
	default:
		err = fmt.Errorf("too many arguments")
		res = makeResponse(err, err.Error(), "")
		return
	}
	if err != nil {
		if err == sql.ErrNoRows {
			res = makeResponse(err, "Command not found", "")
			return
		}
		res = makeResponse(err, "", "")
		return
	}

	if input.Command == "script" {
		text, err := state.dbProxyLinkGet(input.Command, text.String)
		if err != nil {
			res = makeResponse(err, err.Error(), "")
			return
		}
		res = makeResponse(nil, "", text[0])
		return
	}

	if text.Valid {
		res = makeResponse(nil, "", text.String)
		return
	}

	err = errors.New("invalid command")
	res = makeResponse(err, err.Error(), "")
	return
}

func (state *Instance) dbChannelGet(cID string) (string, int64, error) {
	db := state.Database
	var id sql.NullString
	var amt sql.NullInt64
	var err error

	err = db.QueryRow("SELECT msg_id, amount FROM channels WHERE id=(?)", cID).Scan(&id, &amt)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", 0, nil
		}
		return "", 0, err
	}
	if id.Valid && amt.Valid {
		return id.String, amt.Int64, nil
	}
	return "", 0, nil
}
