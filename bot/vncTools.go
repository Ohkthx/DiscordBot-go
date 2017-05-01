package bot

import "fmt"

func battleNotifyText(bid, max, min, cur, queued int) (str string) {
	var qNeed int
	var qStr, plusStr string
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
	h1 := fmt.Sprintf("CTF #%d: ```Queued: %d\nWaiting in game: %d\n\nRequires: %d - %d players.", bid, queued, cur, min, max)
	h2 := fmt.Sprintf("\n%s %d more players to start.```", qStr, min-cur)
	str = h1 + h2
	return
}
