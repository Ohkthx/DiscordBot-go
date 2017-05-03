package bot

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func (state *Instance) dbUpdate(table, extra string, code int) (sndmsg string, err error) {

	switch table {
	case "channels":
		sndmsg, err = state.dbUpdateChannels(extra)
	case "users":
	case "commands":
	case "blacklist":
	default:
		err = fmt.Errorf("unknown table name to update: %s", table)
		return

	}

	return
}

func (state *Instance) dbUpdateChannels(name string) (sndmsg string, err error) {
	var total int64

	c, err := state.ChannelFind(name)
	if err != nil {
		return
	}

	fmt.Println("Channel: " + c.Name)
	total, err = state.procChannel(c, true)
	if err != nil {
		return
	}

	sndmsg = fmt.Sprintf("Processed: %d messages.\n", total)
	return
}

func (state *Instance) procChannel(channel *discordgo.Channel, update bool) (count int64, err error) {
	s := state.Session
	c := channel
	var m *discordgo.Message
	var msgs []*discordgo.Message
	var id, afterID, beforeID, lastID, firstProcd string
	var dberr DBError

	if update {
		lastID, count, err = state.dbChannelGet(channel.ID)
		if err != nil {
			return
		}
		firstProcd = lastID
	}

	for {
		// beforeID, afterID
		// beforeID - If provided, all messages returned will be BEFORE this ID.
		// afterID  - if provided, all messages returned will be AFTER this ID.
		if update {
			afterID = lastID
			beforeID = ""
		} else {
			beforeID = lastID
			afterID = ""
		}
		msgs, err = s.ChannelMessages(c.ID, 100, beforeID, afterID)
		if err != nil {
			break
		}
		// Iterate messages
		for _, m = range msgs {
			dberr = state.dbUserSubmit(m)
			if dberr.Err != nil {
				if dberr.Level < 0 {
					err = fmt.Errorf("error with database: %s", dberr.Err.Error())
					return
				}
				// Not a damaging error.
				continue
			}
			if firstProcd == m.ID {
				return
			} else if firstProcd == "" {
				firstProcd = m.ID
				dberr := state.dbChannelSubmit(channel, m.ID, tsConvert(m.Timestamp))
				if dberr.Err != nil {
					err = dberr.Err
					count = 0
				}
			} else {
				if update {
					dberr = state.dbChannelSubmit(channel, m.ID, tsConvert(m.Timestamp))
					if dberr.Err != nil {
						err = dberr.Err
						count = 0
					}
				}
				if strings.Contains(m.Content, ",event") {
					state.dbNotifyAdd(m.Author.ID, m.Author.Username)
					fmt.Printf("Added: %s[%s]\n", m.Author.Username, m.Author.ID)
				}
			}
			count++
			id = m.ID

			fmt.Printf("Processed: %10d\r", count)
		}
		if id == lastID {
			break
		}
		lastID = id
	}
	err = state.dbChannelAdd(channel.ID, count)
	if err != nil {
		return
	}
	return
}

// dbUserUpdate will process a message and add it to a database.
func (state *Instance) dbUserSubmit(msg *discordgo.Message) (err DBError) {
	var ts string
	db := state.Database
	m := msg
	u := m.Author
	ts = fmt.Sprintf("%s", m.Timestamp)
	ts = tsConvert(msg.Timestamp)
	var id sql.NullString

	// Check if exists
	err.Err = db.QueryRow("SELECT id FROM users WHERE id=(?)", u.ID).Scan(&id)
	if err.Err != nil {
		if err.Err == sql.ErrNoRows {
			_, err.Err = db.Exec("INSERT INTO users (id, username, discriminator, msg_last, msg_count, status, vote) VALUES (?, ?, ?, ?, 1, false, false)", u.ID, u.Username, u.Discriminator, ts)
			if err.Err != nil {
				return
			}
			return
		}
		return
	}
	_, err.Err = db.Exec("UPDATE users SET username=(?), discriminator=(?), msg_count=msg_count+1, status=true, msg_last=(?) WHERE id=(?)", u.Username, u.Discriminator, ts, u.ID)
	if err.Err != nil {
		return
	}

	return
}

func (state *Instance) dbChannelSubmit(c *discordgo.Channel, mID, mTS string) (err DBError) {
	db := state.Database
	var cID = c.ID
	var id sql.NullString
	err.Err = db.QueryRow("SELECT msg_id FROM channels WHERE id=(?)", cID).Scan(&id)
	if err.Err != nil {
		if err.Err == sql.ErrNoRows {
			_, err.Err = db.Exec("INSERT INTO channels (id, name, msg_id, msg_time, amount) VALUES (?, ?, ?, ?, 1)", cID, c.Name, mID, mTS)
			if err.Err != nil {
				return
			}
			return
		}
		return
	}

	if id.String == mID {
		return
	}
	_, err.Err = db.Exec("UPDATE channels SET msg_id=(?), msg_time=(?), name=(?) amount=amount+1 WHERE id=(?)", mID, mTS, c.Name, cID)
	if err.Err != nil {
		return
	}

	return
}

// Modify an existing command
// Requires Granted permissions.
func (state *Instance) dbCommandModify() (res *Response) {
	input := state.Cmd
	db := state.Database
	who := state.User
	whoFull := fmt.Sprintf("%s#%s", who.Username, who.Discriminator)
	var err error

	// Minimum of 2 commands required.
	if input.Length < 2 {
		err = errors.New("not enough arguments. want: [mod] [command(s)] [text]")
		res = makeResponse(err, err.Error(), "")
		return
	} else if state.dbUserPermissions(who.ID) == false {
		err = errors.New("you do not have permissions to do that")
		res = makeResponse(err, err.Error(), "")
		return
	}

	// Check if it exists already
	r := state.dbSearch(input.Length - 1)
	if r.Err != nil {
		res = makeResponse(err, "that command does not exist", "")
		return
	}

	var modified string
	switch input.Length {
	case 2:
		_, err = db.Exec("UPDATE commands SET text=(?), author_mod=(?), date_mod=Now() WHERE command=(?) AND arg1 IS NULL", input.Text, whoFull, input.Args[0])
		modified = fmt.Sprintf("[%s updated]: -> %s", whoFull, input.Args[0])
	case 3:
		if input.Attr&cmdEVENT == cmdEVENT || input.Attr&cmdSCRIPT == cmdSCRIPT {
			return state.dbProxyModify()
		}
		_, err = db.Exec("UPDATE commands SET text=(?), author_mod=(?), date_mod=Now() WHERE command=(?) AND arg1=(?) AND arg2 IS NULL", input.Text, whoFull, input.Args[0], input.Args[1])
		modified = fmt.Sprintf("[%s updated]: -> %s %s", whoFull, input.Args[0], input.Args[1])
	case 4:
		_, err = db.Exec("UPDATE commands SET text=(?), author_mod=(?), date_mod=Now() WHERE command=(?) AND arg1=(?)", input.Text, whoFull, input.Args[0], input.Args[1], input.Args[2])
		modified = fmt.Sprintf("[%s updated]: -> %s %s %s", whoFull, input.Args[0], input.Args[1], input.Args[2])
	default:
		err = fmt.Errorf("too many arguments")
		res = makeResponse(err, err.Error(), "")
		return
	}

	if err != nil {
		res = makeResponse(err, "", "")
		return
	}

	res = makeResponse(nil, "", modified)
	return
}
