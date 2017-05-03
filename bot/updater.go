package bot

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/bwmarrin/discordgo"
)

// BattlegroundUpdater will continuously update #battlegrounds
func BattlegroundUpdater(db *sql.DB, session *discordgo.Session) {
	var err error
	s := New(db, session)
	s.Battle = NewNotifier(notifyBattle)
	s.Event = NewNotifier(notifyEvent)

	err = s.SetChannels(3)
	if err != nil {
		log.Println(err)
		return
	}

	res := s.checkBG("event")
	if res.Err != nil {

	}

	res = s.checkBG("ctf")
	if res.Err != nil {
		log.Println(res.Err)
		return
	}

	for {
		c := time.Tick(15 * time.Second)
		for now := range c {
			s.updateMaintain()

			s.UpdateHandler(battleEvent|battleCTF, now)
			if res.Err != nil {
				log.Println(res.Err)
			}

		}
	}
}

// UpdateHandler just checks what battles need to be updated.
func (state *Instance) UpdateHandler(battles int, ctime time.Time) (res *Response) {
	for battles > 0 {
		switch {
		case battles&battleEvent == battleEvent:
			battles ^= battleEvent
			res = state.editBG("event")
		case battles&battle1v1 == battle1v1:
			battles ^= battle1v1
			res = state.editBG("1v1")
		case battles&battle2v2 == battle2v2:
			battles ^= battle2v2
			res = state.editBG("2v2")
		case battles&battleCTF == battleCTF:
			battles ^= battleCTF
			res = state.editBG("ctf")
		case battles&battleColosseum == battleColosseum:
			battles ^= battleColosseum
			res = state.editBG("colosseum")
		case battles&battleFFA == battleFFA:
			battles ^= battleFFA
			res = state.editBG("ffa")
		case battles&battleGT == battleGT:
			battles ^= battleGT
			res = state.editBG("gt")
		default:
			err := fmt.Errorf("bad battle")
			res = makeResponse(err, err.Error(), "")
			return
		}
		if res.Err != nil {
			return
		}
	}

	res = makeResponse(nil, "", "OK")
	return
}
