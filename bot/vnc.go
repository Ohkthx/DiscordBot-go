package bot

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/d0x1p2/vncgo"
)

// VNCCore handles all things that are for Vita-Nex: Core API
func (state *Instance) VNCCore() (res *Response) {
	switch state.Cmd.Command {
	case "ctf":
		res = state.dbBattle()
	case "online":
		res = getMobile("0x19c03e", "online")
	case "player":
		res = getMobile(state.Cmd.Args[0], "status")
	case "item":
		res = getItem(state.Cmd.Args[0], false)
	case "update":
		if state.Cmd.Length > 0 {
			res = state.checkBG(state.Cmd.Args[0])
		}
	}
	return
}

func getMobile(mobileID, request string) (res *Response) {

	var sndmsg string

	s, err := vncgo.New()
	if err != nil {
		res = makeResponse(err, "", "")
		return
	}

	mobile, err := s.GetMobile(mobileID, false)
	if err != nil {
		res = makeResponse(err, "", "")
		return
	}

	switch strings.ToLower(request) {
	case "online":
		if mobile.Online == true {
			sndmsg = fmt.Sprintf("%s\n```\nStatus: Online\n```", mobile.FullName)
		} else {
			sndmsg = fmt.Sprintf("%s\n```\nStatus: Offline\n```", mobile.FullName)
		}
	case "status":
		sndmsg = fmt.Sprintf("%s\n```\nOnline: %v\nHealth: %d / %d\nDeaths: %d\nFame: %d, Karma: %d```",
			mobile.FullName, mobile.Online, mobile.Stats.Hits, mobile.Stats.HitsMax, mobile.Deaths, mobile.Fame, mobile.Karma)
	default:
		sndmsg = fmt.Sprintf("%s found.", mobile.FullName)
	}

	res = makeResponse(nil, "", sndmsg)
	return
}

func getBattles(instance string) (res *Response) {

	var sndmsg string

	s, err := vncgo.New()
	if err != nil {
		res = makeResponse(err, "", "")
		return
	}

	battles, err := s.GetBattlesID()
	if err != nil {
		res = makeResponse(err, "", "")
		return
	}

	var battle *vncgo.Battle
	var i int

	if strings.ToLower(instance) == "ctf" {
		sndmsg = fmt.Sprintf("Capture the Flag:\n```")
		for n := 0; n < len(battles); n++ {
			if strings.ToLower(battles[n].Name) == "capture the flag" {
				battle, err = s.GetBattle(battles[n].ID)
				if err != nil {
					res = makeResponse(err, "", "")
					return
				}
				switch strings.ToLower(battle.State) {
				case "internal":
				default:
					if battle.Queued > 0 || len(battle.Players) > 0 {
						i++
						sndmsg += fmt.Sprintf("\nCTF #%d:\n  Playing: %d\n  Queue:   %d\n", i, len(battle.Players), battle.Queued)
					}
				}
			}
		}
		sndmsg += "```"

	}

	//err = errors.New("no queues found")
	if i == 0 {
		name, err := convBGNameReverse(instance)
		if err != nil {
			name = "Unknown?"
		}
		res = makeResponse(nil, "", fmt.Sprintf("%s: \n```No one in queue.```", name))
		return
	}

	res = makeResponse(nil, "", sndmsg)

	return
}

func getItem(itemID string, internal bool) (res *Response) {
	const errStr = "unexpected error :)"

	s, err := vncgo.New()
	if err != nil {
		res = makeResponse(err, "", "")
		return
	}

	item, err := s.GetItem(itemID, false)
	if err != nil {
		res = makeResponse(err, "", "")
		return
	}

	itemDesc := vncgo.CleanHTML(item.Opl)
	res = makeResponse(nil, "", fmt.Sprintf("```\n%s```", itemDesc))

	return
}

/*

	BG editing of messages in #battlegrounds below

*/

func (state *Instance) checkBG(name string) (res *Response) {
	if state.BG == nil {
		// Get channel
		c, err := state.ChannelExist("battlegrounds")
		if err != nil {
			res = makeResponse(err, err.Error(), "")
			return
		}
		state.BG = &Battleground{ChannelID: c.ID}
		// Attempt to load battles here
		res = state.loadBGDB(name)
		if res.Err != nil {
			if res.Err.Error() == "not found" {
				res = state.loadBGAPIBattle(name)
				if res.Err != nil {
					return
				}
			} else {
				return
			}
		}
	}
	// Update here.
	res = state.editBG(name)
	if res.Err != nil {
		return
	}
	return
}

func (state *Instance) loadBGDB(name string) (res *Response) {
	// Check SQL DB for message IDs
	// if not, get battles + create intial messages.
	//var battles []Battle
	var found, exists bool
	var sqlMsgID, sqlName sql.NullString
	db := state.Database

	// Load rows from DB
	rows, err := db.Query("SELECT msgid, name FROM battlegrounds")
	if err != nil {
		if err == sql.ErrNoRows {
			res = makeResponse(err, "not found", "")
			return
		}
		res = makeResponse(err, err.Error(), "")
		return
	}
	defer rows.Close()

	name, _ = convBGNameReverse(name)
	name = strings.ToLower(name)

	// Process rows from Db
	for rows.Next() {
		err = rows.Scan(&sqlMsgID, &sqlName)
		if err != nil {
			res = makeResponse(err, err.Error(), "")
			return
		}
		if sqlMsgID.Valid && sqlName.Valid {
			if strings.ToLower(sqlName.String) == name {
				for _, p := range state.BG.Battles {
					if strings.ToLower(p.Name) == name {
						res = makeResponse(nil, "", "Already added to battles")
						return
					}
				}
				state.BG.Battles = append(state.BG.Battles, Battle{MsgID: sqlMsgID.String, Name: sqlName.String})
				res = makeResponse(nil, "", "Loaded from DB")
				return
			} else if name == "" {
				found = true
				for _, p := range state.BG.Battles {
					if p.Name == name {
						exists = true
						break
					}
				}
				if exists {
					continue
				}
				state.BG.Battles = append(state.BG.Battles, Battle{MsgID: sqlMsgID.String, Name: sqlName.String})
			}
		}
	}

	if found {
		res = makeResponse(nil, "", "Loaded from DB")
		return
	}

	err = fmt.Errorf("not found")
	res = makeResponse(err, err.Error(), "")
	return
}

func (state *Instance) loadBGAPI(name string) (res *Response) {
	var battles []Battle
	var c *discordgo.Channel
	var err error

	s, err := vncgo.New()
	if err != nil {
		res = makeResponse(err, err.Error(), "")
		return
	}

	apibattles, err := s.GetBattlesID()
	if err != nil {
		res = makeResponse(err, err.Error(), "")
		return
	}

	// Create initial message.
	c, err = state.ChannelFind("battlegrounds")
	if err != nil {
		res = makeResponse(err, err.Error(), "")
		return
	}

	// Assign Structure (Battleground)
	var assigned bool
	for _, b := range apibattles {
		assigned = false
		m, err := state.Session.ChannelMessageSend(c.ID, "``` [PlaceHolder] ```")
		if err != nil {
			res = makeResponse(err, err.Error(), "")
			return
		}

		if len(battles) > 0 {
			for _, bb := range battles {
				if bb.Name == b.Name {
					assigned = true
					break
				}
			}
		}

		if assigned {
			continue
		}

		battles = append(battles, Battle{MsgID: m.ID, Name: b.Name})
	}

	state.BG.Battles = battles

	// Save to SQL to prevent lookups.
	res = state.dbBGSave()
	if res.Err != nil {
		return
	}

	res = makeResponse(nil, "", "Loaded and saved.")
	return
}

func (state *Instance) loadBGAPIBattle(name string) (res *Response) {
	//var battles []Battle
	var c *discordgo.Channel
	var err error

	s, err := vncgo.New()
	if err != nil {
		res = makeResponse(err, err.Error(), "")
		return
	}

	apibattles, err := s.GetBattlesID()
	if err != nil {
		res = makeResponse(err, err.Error(), "")
		return
	}

	// Create initial message.
	c, err = state.ChannelFind("battlegrounds")
	if err != nil {
		res = makeResponse(err, err.Error(), "")
		return
	}

	if name != "" {
		name, _ = convBGNameReverse(name)
		name = strings.ToLower(name)
	}

	// Assign Structure (Battleground)
	var assigned bool
	for _, b := range apibattles {

		if strings.ToLower(name) != "" {

			for _, bb := range state.BG.Battles {
				if strings.ToLower(bb.Name) == name {
					res = makeResponse(nil, "", "Already assigned")
					return
				}
			}
			if strings.ToLower(b.Name) == name {
				m, err := state.Session.ChannelMessageSend(c.ID, "``` [PlaceHolder] ```")
				if err != nil {
					res = makeResponse(err, err.Error(), "")
					return
				}
				state.BG.Battles = append(state.BG.Battles, Battle{MsgID: m.ID, Name: b.Name})
				res = state.dbBGSave()
				if res.Err != nil {
					return
				}

				res = makeResponse(nil, "", "Loaded and saved.")
				return
			}
			continue

		}

		assigned = false
		m, err := state.Session.ChannelMessageSend(c.ID, "``` [PlaceHolder] ```")
		if err != nil {
			res = makeResponse(err, err.Error(), "")
			return
		}

		if len(state.BG.Battles) > 0 {
			for _, bb := range state.BG.Battles {
				if bb.Name == b.Name {
					assigned = true
					break
				}
			}
		}

		if assigned {
			continue
		}

		state.BG.Battles = append(state.BG.Battles, Battle{MsgID: m.ID, Name: b.Name})
	}

	// Save to SQL to prevent lookups.
	res = state.dbBGSave()
	if res.Err != nil {
		return
	}

	res = makeResponse(nil, "", "Loaded and saved.")
	return
}

func (state *Instance) dbBGSave() (res *Response) {
	db := state.Database
	battles := state.BG.Battles

	for _, b := range battles {
		_, err := db.Exec("INSERT INTO battlegrounds (msgid, name) VALUES (?, ?)", b.MsgID, b.Name)
		if err != nil {
			res = makeResponse(err, err.Error(), "")
			return
		}
	}

	res = makeResponse(nil, "", "Battles saved to Database")
	return
}

// editBG is a wrapper for ChannelMessageEdit from discordgo API
func (state *Instance) editBG(msgType string) (res *Response) {
	s := state.Session
	c := state.BG.ChannelID
	var r *Response

	switch strings.ToLower(msgType) {
	case "ctf":
		r = getBattles("ctf")
	case "events":
		fallthrough
	case "event":
		// Update ability here.
		break

	}

	if r.Err != nil {
		return
	}

	for _, p := range state.BG.Battles {
		tmp, err := convBGName(p.Name)
		if err != nil {
			res = makeResponse(err, err.Error(), "")
			return
		}
		if tmp == msgType {
			s.ChannelMessageEdit(c, p.MsgID, r.Sndmsg)
			res = makeResponse(nil, "", fmt.Sprintf("Updated %s[%s].", c, p.MsgID))
			return
		}
	}
	err := fmt.Errorf("bad channel to edit")
	res = makeResponse(err, err.Error(), "")
	return
}

func convBGName(in string) (out string, err error) {
	switch strings.ToLower(in) {
	case "capture the flag":
		out = "ctf"
	case "1 vs 1":
		out = "1v1"
	case "2 vs 2":
		out = "2v2"
	case "gladiator training":
		out = "gt"
	case "colosseum":
		out = "colosseum"
	case "free for all":
		out = "ffa"
	default:
		err = fmt.Errorf("bad BG name")
	}
	return
}
func convBGNameReverse(in string) (out string, err error) {
	switch strings.ToLower(in) {
	case "ctf":
		out = "Capture the Flag"
	case "1v1":
		out = "1 vs 1"
	case "2v2":
		out = "2 v 2"
	case "gt":
		out = "Gladiator Training"
	case "colosseum":
		out = "Colosseum"
	case "ffa":
		out = "Free for All"
	default:
		err = fmt.Errorf("bad BG name")
	}
	return
}
