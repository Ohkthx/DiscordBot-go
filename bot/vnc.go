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
	case "event":
		fallthrough
	case "events":
		state.dbNotifyAdd(state.User.ID, state.User.Username)
		fallthrough
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

func (state *Instance) getBattles(instance string) (res *Response) {

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

	name, err := convBGNameReverse(instance)
	if err != nil {
		err = fmt.Errorf("bad requested battle")
		res = makeResponse(err, err.Error(), "")
		return
	}

	var mainUpdate bool
	var mainText string
	sndmsg = fmt.Sprintf("%s:\n```", name)
	for n := 0; n < len(battles); n++ {
		if strings.ToLower(battles[n].Name) == strings.ToLower(name) {
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
					t := battle.Queued + battle.CurCapacity
					if t >= battle.MinCapacity-3 && strings.ToLower(battle.State) != "running" {
						// 3 or less from starting.
						mainUpdate = true
						mainText += state.battleNotifyText(i, battle)
					} else if t == 0 && strings.ToLower(battle.State) != "running" && state.Battle.Msg != nil {
						_, err = state.Session.ChannelMessageEdit(state.Battle.Msg.ChannelID, state.Battle.Msg.ID, fmt.Sprintf("```CTF #%2d: Everyone left queue.```", i))
						if err != nil {
							state.Battle.Reset()
							res = makeResponse(err, err.Error(), "")
							return
						}
						state.Battle.Reset()
					} else if t < battle.MinCapacity && strings.ToLower(battle.State) != "running" && state.Battle.Msg != nil {
						mainUpdate = true
						mainText += state.battleNotifyText(i, battle)
					} else if strings.ToLower(battle.State) == "running" && state.Battle.Msg != nil {
						if state.MainChan != nil {
							var xtxt string
							if battle.MaxCapacity-battle.CurCapacity > 0 {
								xtxt = fmt.Sprintf("%d more players can join.", battle.MaxCapacity-battle.CurCapacity)
							}
							_, err = state.Session.ChannelMessageSend(state.MainChan.ID, fmt.Sprintf("```CTF #%d has started with %d players.\n%s```", i, battle.CurCapacity, xtxt))
							if err != nil {
								state.Battle.Reset()
								res = makeResponse(err, err.Error(), "")
								return
							}
							state.Battle.Reset()
						}
					}
					sndmsg += battleEventText(i, battle)
				}
			}
		}
	}
	sndmsg += "```\n**>----------------------------------------<**"

	if mainUpdate {
		// Send notification here
		if state.Battle.Notified && state.Battle.Msg != nil {
			state.Session.ChannelMessageEdit(state.Battle.Msg.ChannelID, state.Battle.Msg.ID, mainText)
		} else if state.Battle.Notified {
			//state.Battle.Notified = false
			state.notify(notifyBattle, mainText)
		}
	}

	if i == 0 {
		name, err := convBGNameReverse(instance)
		if err != nil {
			name = "Unknown?"
		}
		res = makeResponse(nil, "", fmt.Sprintf("%s: \n```No one in queue.```**>----------------------------------------<**", name))
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
	}

	// Attempt to load battles here
	res = state.loadBGDB(name)
	if res.Err != nil {
		if res.Err.Error() == "not found" {
			if name == "event" {
				res = state.loadBGEvent()
			} else {
				res = state.loadBGAPIBattle(name)
			}
			if res.Err != nil {
				return
			}
		} else {
			return
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

	name, err = convBGNameReverse(name)
	if err != nil {
		res = makeResponse(err, err.Error(), "")
		return
	}
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
				state.BG.Battles = append(state.BG.Battles, BattleID{MsgID: sqlMsgID.String, Name: sqlName.String})
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
				state.BG.Battles = append(state.BG.Battles, BattleID{MsgID: sqlMsgID.String, Name: sqlName.String})
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
	var battles []BattleID
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

		battles = append(battles, BattleID{MsgID: m.ID, Name: b.Name})
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
		name, err = convBGNameReverse(name)
		if err != nil {
			res = makeResponse(err, err.Error(), "")
			return
		}
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
				state.BG.Battles = append(state.BG.Battles, BattleID{MsgID: m.ID, Name: b.Name})
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

		state.BG.Battles = append(state.BG.Battles, BattleID{MsgID: m.ID, Name: b.Name})
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
	case "event":
		r = state.dbEvent()
		r.Sndmsg += "\n**>----------------------------------------<**"
	default:
		r = state.getBattles(msgType)
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
		if strings.ToLower(tmp) == msgType {
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
	case "event":
		out = "event"
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
	case "event":
		out = "Event"
	default:
		err = fmt.Errorf("bad BG name")
	}
	return
}
