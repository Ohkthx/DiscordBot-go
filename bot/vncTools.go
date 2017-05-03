package bot

import (
	"fmt"

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
	h1 := fmt.Sprintf("CTF #%d: (*check <#%s> for updates*)```State: %s\nQueued: %d\nWaiting in game: %d\n\nRequires: %d - %d players.", cnt, state.BG.ChannelID, battle.State, queued, cur, min, max)
	h2 := fmt.Sprintf("\n%s %d more players to start.```", qStr, min-cur)
	str = h1 + h2
	return
}

func battleEventText(cnt int, battle *vncgo.Battle) (str string) {
	return fmt.Sprintf("\nCTF #%d:\n  Playing: %d\n  Queued:  %d\n", cnt, len(battle.Players), battle.Queued)
}
