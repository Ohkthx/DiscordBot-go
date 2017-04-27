package main

import (
	"database/sql"
	"errors"

	"github.com/d0x1p2/DiscordBot-go/bot"
)

func voteProcessor(info *bot.Instance) (string, error) {
	input := info.Cmd
	command := input.Command
	var err error

	if command != "yay" && command != "nay" {
		err = errors.New("bad command: " + command)
		return "", err
	}

	if command == "yay" || command == "nay" {
		var cnt sql.NullInt64
		// Check if it exists.
		err = db.QueryRow("SELECT count FROM commands WHERE command=(?) AND arg1 IS NULL", command).Scan(&cnt)
		if err != nil {
			if err == sql.ErrNoRows {
				// INSERT HERE
			}
		}
		// update if it exists.
	}

	return "", nil
}
