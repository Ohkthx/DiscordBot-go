package bot

import (
	"fmt"
	"strings"

	"github.com/d0x1p2/vncgo"
)

func (state *Instance) battleNotifyText(cnt int, battle *vncgo.Battle) (str string) {
	var qNeed int
	var qStr, plusStr string
	min := battle.MinCapacity
	max := battle.MaxCapacity
	queued := battle.Queued
	cur := battle.CurCapacity
	// max = 10, cur = 6, queue = 2
	// total = 8
	if (min-cur)-queued > 0 {
		addNeeded := (min - cur) - queued
		plusStr = fmt.Sprintf("(+%d additional players) ", addNeeded)
	}

	if queued > 0 {
		qNeed = queued
		if (cur + queued) > min {
			qNeed = min - cur
		}
		qStr = fmt.Sprintf("Need: %d of the queue(d) %sjoin or need", qNeed, plusStr)
	} else {
		qStr = "Need:"
	}
	h1 := fmt.Sprintf("CTF #%d: (*check <#%s> for updates*)```State: %s\nQueued: %d\nWaiting in game: %d\n\nRequires: %d - %d players.", cnt, state.Events.ChannelID, battle.State, queued, cur, min, max)
	h2 := fmt.Sprintf("\n%s %d more players to start.```", qStr, min-cur)
	str = h1 + h2
	return
}

func battleEventText(cnt int, battle *vncgo.Battle) (str string) {
	return fmt.Sprintf("```C\n%s #%d:\n  Playing: %d\n  Queued:  %d```", battle.Name, cnt, len(battle.Players), battle.Queued)
}

func (state *Instance) getBattleMSG(class string) (string, error) {
	var sndmsg string
	battles := state.Battles

	var i int

	vnc, err := vncgo.New()
	if err != nil {
		return "", err
	}

	sndmsg = fmt.Sprintf("%s:", class)

	for _, b := range battles {
		if strings.ToLower(b.Type) == strings.ToLower(class) {
			battle, err := vnc.GetBattle(int64(b.ID))
			if err != nil {
				return "", err
			}
			if strings.ToLower(battle.State) != "internal" {
				if battle.Queued == 0 && len(battle.Players) == 0 {
					sndmsg += fmt.Sprintf("\n```[%s] No one in queue.```", battle.Name)
				} else {
					i++
					sndmsg += battleEventText(i, battle)
				}
			}
		}
	}

	sndmsg += fmt.Sprintf("\n%s", MSGDIV)

	if i == 0 {
		sndmsg = fmt.Sprintf("\n%s:```No one in queue.```%s", class, MSGDIV)
		return sndmsg, nil
	}

	return sndmsg, nil
}

func typeConvert2String(class int) string {
	switch {
	case class&EventBattle == EventBattle:
		return "EventBattle"
	case class&CTFBattle == CTFBattle:
		return "CTFBattle"
	case class&FFABattle == FFABattle:
		return "FFABattle"
	case class&TvTBattle1v1 == TvTBattle1v1:
		return "TvTBattle1v1"
	case class&TvTBattle2v2 == TvTBattle2v2:
		return "TvTBattle2v2"
	case class&TvTBattle3v3 == TvTBattle3v3:
		return "TvTBattle3v3"
	case class&TvTBattle4v4 == TvTBattle4v4:
		return "TvTBattle4v4"
	case class&TvTBattle5v5 == TvTBattle5v5:
		return "TvTBattle5v5"
	default:
		return ""
	}
}

func typeConvert2Int(class string) int {
	switch strings.ToLower(class) {
	case "eventbattle":
		return EventBattle
	case "ctfbattle":
		return CTFBattle
	case "ffabattle":
		return FFABattle
	case "tvtbattle1v1":
		return TvTBattle1v1
	case "tvtbattle2v2":
		return TvTBattle2v2
	case "tvtbattle3v3":
		return TvTBattle3v3
	case "tvtbattle4v4":
		return TvTBattle4v4
	case "tvtbattle5v4":
		return TvTBattle5v5
	default:
		return 0
	}
}
