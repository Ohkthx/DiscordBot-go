package bot

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/d0x1p2/vncgo"
)

// Error constants for Update Handler
var (
	ErrNilMem = errors.New("accessing nil memory")
)

// EventsUpdater will continuously update #events
func EventsUpdater(db *sql.DB, session *discordgo.Session) {
	var err error
	s := New(db, session)
	s.BattleNotifier = NewNotifier(notifyBattle)
	s.EventNotifier = NewNotifier(notifyEvent)

	err = s.SetChannels(3)
	if err != nil {
		log.Println(err)
		return
	}

	err = s.updatePrepValidation()
	if err != nil {
		log.Println(err)
		return
	}

	for {
		c := time.Tick(15 * time.Second)
		for now := range c {
			s.updateMaintain()
			err := s.UpdateHandler(now)
			if err != nil {
				log.Println(err)
			}

		}
	}
}

// UpdateHandler just checks what battles need to be updated.
func (state *Instance) UpdateHandler(ctime time.Time) (err error) {
	err = state.bgChecks()
	return
}

func (state *Instance) updatePrepValidation() (err error) {
	var c *discordgo.Channel
	var battles []*vncgo.Battle
	var msgid string
	var found bool

	// Check and make sure the Update Info struct exists
	if state.Events == nil {
		// Get channel
		c, err = state.ChannelExist(UpdateChannel)
		if err != nil {
			return
		}
		state.Events = &Events{ChannelID: c.ID}
	}

	// Load all Battles
	s, err := vncgo.New()
	if err != nil {
		return err
	}

	battles, err = s.GetBattles()
	if err != nil {
		return err
	}

	if len(state.Events.Battles) > 0 {
		for _, b := range state.Events.Battles {
			if b.Type == "Event" {
				found = true
			}
		}
	}

	if found == false {
		msgid, err = state.checkCreate("Event")
		if err != nil {
			return
		}
		state.Events.Battles = append(state.Events.Battles, BattleID{Type: "Event", MsgID: msgid})
	}

	for _, b := range battles {
		// Check if MSGid exists in DB.
		msgid, err = state.checkCreate(b.Type)
		if err != nil {
			fmt.Println(err)
		}

		state.Battles = append(state.Battles, b)
		state.Events.Battles = append(state.Events.Battles, BattleID{Type: b.Type, MsgID: msgid})
	}

	/*
		if len(state.Battles) < 1 || len(state.Events.Battles) < 1 {
			return ErrNilMem
		}
	*/

	return nil
}

func (state *Instance) checkCreate(class string) (msgid string, err error) {
	var c *discordgo.Channel
	var m *discordgo.Message

	// Check and make sure the Update Info struct exists
	if state.Events == nil {
		// Get channel
		c, err = state.ChannelExist(UpdateChannel)
		if err != nil {
			return
		}
		state.Events = &Events{ChannelID: c.ID}
	}

	msgid, err = state.dbCheckBattle(class)
	if err != nil {
		if err == ErrNotFound {
			// Database entry doesn't exist. Create message/Save
			m, err = state.Session.ChannelMessageSend(state.Events.ChannelID, "``` [PlaceHolder] ```")
			if err != nil {
				return "", err
			}
			msgid = m.ID
			err = state.bgSaveDB(class, msgid)
			if err != nil {
				return
			}
		} else {
			return
		}
	}
	return
}
