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

	s, err := New(db, session)
	if err != nil {
		log.Println(err)
		return
	}

	res := s.checkBG("ctf")
	if res.Err != nil {
		return
	}

	fmt.Println("Channel ID: ", s.BG.ChannelID)

	for {
		c := time.Tick(15 * time.Second)
		for now := range c {
			fmt.Println("Updating", now)
			res := s.editBG("ctf")
			if res.Err != nil {
				log.Println(res.Err)
			}
		}
	}
}
