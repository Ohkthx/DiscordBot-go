package main

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func dbUpdate(info *inputInfo, startup bool) (int64, error) {
	session := info.session

	var cnt, total int64
	var err error

	guilds, err := session.UserGuilds()
	if err != nil {
		errLog.Println("error getting guilds", err)
		err = errors.New("could not get server info")
		return -1, err
	}

	// Iterate guilds
	for _, g := range guilds {
		channels, err := session.GuildChannels(g.ID)
		if err != nil {
			errLog.Println("error getting guild channels", err)
			err = errors.New("error getting rooms")
			return -1, err
		}
		// Iterate channels
		for _, c := range channels {
			if c.Name != info.channel.Name {
				continue
			}
			fmt.Println("Channel: " + c.Name)
			cnt, err = procChannel(session, c, startup)
			if err != nil {
				errLog.Printf("%d: %s\n", cnt, err)
			}
			total += cnt
		}
	}

	return total, nil
}

func procChannel(session *discordgo.Session, channel *discordgo.Channel, startup bool) (int64, error) {
	s := session
	c := channel
	var m *discordgo.Message
	var msgs []*discordgo.Message
	var total int64
	var err error
	var id, afterID, beforeID, lastID, firstProcd string

	if startup {
		lastID, total, err = dbChannelGet(channel.ID)
		if err != nil {
			return -1, err
		}
		firstProcd = lastID
		fmt.Println("Preproc LASTID: " + firstProcd)
	}

	for {
		// beforeID, afterID
		// beforeID - If provided, all messages returned will be BEFORE this ID.
		// afterID  - if provided, all messages returned will be AFTER this ID.
		if startup {
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
			err = dbUserUpdate(m)
			if err != nil {
				errLog.Println("error with db " + err.Error())
				continue
			}
			if firstProcd == m.ID {
				if *debug {
					fmt.Printf("%s -> %s\n", "procChannel", "firstProcd matches ID")
				}
				return total, nil
			} else if firstProcd == "" {
				if *debug {
					fmt.Printf("%s -> %s\n", "procChannel", "firstProc is empty, assigning.")
				}
				firstProcd = m.ID
				err = dbChannelUpdate(channel.ID, m.ID, tsConvert(m.Timestamp))
				if err != nil {
					return 0, err
				}
			} else {
				if startup {
					if *debug {
						fmt.Printf("%s: %s -> Processed.\n", m.ID, m.Content)
					}
					err = dbChannelUpdate(channel.ID, m.ID, tsConvert(m.Timestamp))
					if err != nil {
						return 0, err
					}
				}
			}
			total++
			id = m.ID
			if *debug {
				fmt.Printf("%s: %s\n", m.ID, m.Content)
			}
			//fmt.Printf("Messages processed: %6d\n", total)
			//time.Sleep(500 * time.Millisecond)
		}
		if id == lastID {
			break
		}
		if *debug {
			fmt.Println("Last ID: " + lastID)
		}
		lastID = id
	}
	err = dbChannelAdd(channel.ID, total)
	if err != nil {
		return total, err
	}
	return total, nil
}

func dbUserUpdate(msg *discordgo.Message) error {
	var ts string
	m := msg
	u := m.Author
	ts = fmt.Sprintf("%s", m.Timestamp)
	ts = tsConvert(msg.Timestamp)
	var err error
	var id sql.NullString
	// Check if exists
	err = db.QueryRow("SELECT id FROM users WHERE id=(?)", u.ID).Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			_, err = db.Exec("INSERT INTO users (id, username, discriminator, msg_count, msg_last, status) VALUES (?, ?, ?, 1, ?, true)", u.ID, u.Username, u.Discriminator, ts)
			if err != nil {
				errLog.Println("adding user to users" + err.Error())
				return err
			}
			//err = dbChannelUpdate(m.ChannelID, m.ID, ts)
			return nil
		}
		errLog.Println("getting user information" + err.Error())
		return err
	}
	_, err = db.Exec("UPDATE users SET username=(?), discriminator=(?), msg_count=msg_count+1, status=true, msg_last=(?) WHERE id=(?)", u.Username, u.Discriminator, ts, u.ID)
	if err != nil {
		errLog.Println("updating user in users " + err.Error())
		return err
	}
	//err = dbChannelUpdate(m.ChannelID, m.ID, ts)
	//if err != nil {
	//	return err
	//}
	return nil
}

func dbChannelUpdate(cID, mID, mTS string) error {
	var err error
	var id sql.NullString
	err = db.QueryRow("SELECT msg_id FROM channels WHERE id=(?)", cID).Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			_, err = db.Exec("INSERT INTO channels (id, msg_id, msg_time, amount) VALUES (?, ?, ?, 1)", cID, mID, mTS)
			if err != nil {
				errLog.Println("adding a channel " + err.Error())
				return err
			}
			return nil
		}
		errLog.Println("getting channel information " + err.Error())
		return err
	}

	if id.String == mID {
		// DEBUGGING
		if *debug {
			fmt.Printf("%s -> %s\n", "dbChannelUpdate", "DB ID matches update ID")
		}
		return nil
	}
	_, err = db.Exec("UPDATE channels SET msg_id=(?), msg_time=(?), amount=amount+1 WHERE id=(?)", mID, mTS, cID)
	if err != nil {
		errLog.Println("updating channel in channels " + err.Error())
		return err
	}
	// DEBUGGING
	if *debug {
		fmt.Printf("%s -> %s\n", "dbChannelUpdate", "DB updated")
	}
	return nil
}

func dbChannelAdd(cID string, total int64) error {
	var err error
	_, err = db.Exec("UPDATE channels SET amount=(?) WHERE id=(?)", total, cID)
	if err != nil {
		return err
	}
	return nil
}

func dbChannelGet(cID string) (string, int64, error) {
	var err error
	var id sql.NullString
	var amt sql.NullInt64

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

func tsConvert(ts discordgo.Timestamp) string {
	a := strings.FieldsFunc(fmt.Sprintf("%s", ts), tsSplit)
	return fmt.Sprintf("%s %s", a[0], a[1])
}

func tsSplit(r rune) bool {
	return r == 'T' || r == '.' || r == '+'
}
