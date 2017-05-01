package bot

import (
	"database/sql"
	"fmt"
)

const (
	notifyEvent = 1 << iota
	notifyBattle
)

// NewNotifier creates a new instance of Notifier and returns a pointer to it.
func NewNotifier(t int, message string) (n *Notifier, err error) {
	n = &Notifier{
		Type:    t,
		Message: message,
	}
	return
}

func (state *Instance) notify(t int, message string) (err error) {
	var users []User
	s := state.Session
	switch {
	case t&notifyEvent == notifyEvent:
		users, err = state.dbNotifyGet()
		if err != nil {
			return
		}
		for _, u := range users {
			err = state.notifyUserSend(u.ID, u.Name, message, false)
			if err != nil {
				return
			}
		}
		_, err = s.ChannelMessageSend(state.MainChan.ID, message)
		return
	case t&notifyBattle == notifyBattle:
		_, err = s.ChannelMessageSend(state.MainChan.ID, message)
		return
	default:
		err = fmt.Errorf("unknown object(s) to notify")
	}
	return
}

func (state *Instance) dbNotifyGet() (u []User, err error) {
	var id, name sql.NullString
	db := state.Database

	rows, err := db.Query("SELECT id, name FROM subscriptions WHERE sub=true")
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&id, &name)
		if err != nil {
			return
		}
		if id.Valid && name.Valid {
			u = append(u, User{ID: id.String, Name: name.String})
		}
	}
	return
}

func (state *Instance) dbNotifyAdd(id, name string) (err error) {
	db := state.Database
	var res sql.NullString
	var sub sql.NullBool

	err = db.QueryRow("SELECT name, sub FROM subscriptions WHERE id=(?)", id).Scan(&res, &sub)
	if err != nil {
		if err == sql.ErrNoRows {
			// No rows, attempt an add.
			_, err = db.Query("INSERT INTO subscriptions (id, name, sub) VALUES (?, ?, true)", id, name)
		}
		// Another issue, just return
		return
	}
	if res.Valid && sub.Valid {
		// Exists already
		if sub.Bool == false {
			_, err = db.Query("UPDATE subscriptions SET sub=true WHERE id=(?)", id)
		}
		return
	}
	err = fmt.Errorf("something unexpected with adding/checking subscription")
	return
}

// NotifyUnsub is a facing wrapper for Unsubscripting from database
func (state *Instance) NotifyUnsub() (res *Response) {
	if state.Channel.IsPrivate == false {
		res = makeResponse(fmt.Errorf("not a private message"), "Please unsubscribe in here(private)", "")
		if res.Err != nil {
			return
		}
		res = makeResponse(nil, "", "")
		return
	}
	err := state.dbNotifyUnsub(state.User.ID)
	if err != nil {
		res = makeResponse(err, "", "")
		return
	}
	res = makeResponse(nil, "", "")
	return
}

func (state *Instance) dbNotifyUnsub(id string) (err error) {
	db := state.Database
	_, err = db.Query("UPDATE subscriptions SET sub=false WHERE id=(?)", id)
	if err != nil {
		return
	}
	err = state.notifyUserSend(id, "", "", true)
	return
}

func (state *Instance) notifyUserSend(id, name, info string, unsub bool) (err error) {
	s := state.Session

	channel, err := s.UserChannelCreate(id)
	if err != nil {
		return
	}

	var subtxt string
	if unsub == false {
		usubtxt := fmt.Sprintf("If you wish to unsubscribe from these alerts, type:\n`,unsubscribe`")
		subtxt = fmt.Sprintf("Automated subscription services for %s:```%s```%s", name, info, usubtxt)
	} else {
		subtxt = fmt.Sprintf("You have successfully unsubscribed! If you wish to resubscribe use:\n`,event` or `,events`")
	}

	_, err = s.ChannelMessageSend(channel.ID, subtxt)
	if err != nil {
		return
	}
	return
}
