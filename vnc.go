package main

import (
	"errors"
	"fmt"
	"strings"

	"github.com/d0x1p2/vncgo"
)

func getMobile(mobileID, request string) (sndmsg string, err error) {
	const errStr = "unexpected error :)"

	s, err := vncgo.New()
	if err != nil {
		errLog.Println(err)
		err = errors.New(errStr)
		return
	}

	mobile, err := s.GetMobile(mobileID, false)
	if err != nil {
		errLog.Println(err)
		err = errors.New(errStr)
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

	return
}

func getBattles(instance string) (sndmsg string, err error) {
	const errStr = "unexpected issue :)"

	s, err := vncgo.New()
	if err != nil {
		errLog.Println(err)
		err = errors.New(errStr)
		return
	}

	battles, err := s.GetBattlesID()
	if err != nil {
		errLog.Println(err)
		err = errors.New(errStr)
		return
	}

	var battle *vncgo.Battle

	if strings.ToLower(instance) == "ctf" {
		for _, b := range battles {
			if strings.ToLower(b.Name) == "capture the flag" {
				battle, err = s.GetBattle(b.ID)
				if err != nil {
					errLog.Println(err)
					err = errors.New(errStr)
					return
				}
				if battle.Queued > 0 || len(battle.Players) > 0 {
					sndmsg = fmt.Sprintf("%s:\n```\nPlaying: %d\nQueue:   %d\n```", battle.Name, len(battle.Players), battle.Queued)
					return
				}
			}
		}
	}

	//err = errors.New("no queues found")
	return "No one in queue.", nil
}

func getItem(itemID string, internal bool) (sndmsg string, err error) {
	const errStr = "unexpected error :)"

	s, err := vncgo.New()
	if err != nil {
		errLog.Println(err)
		err = errors.New(errStr)
		return
	}

	item, err := s.GetItem(itemID, false)
	if err != nil {
		errLog.Println(err)
		err = errors.New(errStr)
		return
	}

	itemDesc := vncgo.CleanHTML(item.Opl)
	sndmsg = fmt.Sprintf("```\n%s```", itemDesc)

	return
}
