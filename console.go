package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/d0x1p2/DiscordBot-go/bot"
)

func core(session *discordgo.Session) {
	// Main loop for processing user input to console.
	state, err := bot.New(db, session)
	if err != nil {
		errLog.Println(err)
		return
	}
	state.Admin = false
	state.User = nil
	state.Channel = nil
	state.Sendmsg = false

	go bot.BattlegroundUpdater(state.Database, state.Session)

	time.Sleep(150 * time.Millisecond)
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf("[%s] > ", time.Now().Format(time.Stamp))
		input, _ := reader.ReadString('\n')
		temp := strings.Fields(input)
		state.Cmd = inputText(strings.Join(temp, " "))
		if state.Cmd.Command != "" {
			err := ioHandler(state)
			if err != nil {
				fmt.Println(err)
			}
		}
	}
}

func ioHandler(state *bot.Instance) (err error) {
	// Used to parse user input to console/cli
	input := state.Cmd
	user := state.User
	channel := state.Channel
	session := state.Session

	//var sndmsg string

	switch input.Command {
	/*case "ch-base":
		sndmsg, err = state.DBCore("channels", bot.Add)
		if err != nil {
			return
		}
		log.Printf(sndmsg)
	case "ch-update":
		sndmsg, err = state.DBCore("channels", bot.Modify)
		if err != nil {
			return
		}
		log.Printf(sndmsg)
	*/
	case "update":
		res := state.VNCCore()
		if res.Err != nil {
			log.Println(res.Err)
			break
		}
		log.Println(res.Sndmsg)
	case "stream":
		// Enable and Disable flags.
		if state.Streaming == false {
			err = session.UpdateStreamingStatus(0, "Ultima-Shards: AOS", "https://www.twitch.tv/d0x1p2")
			if err != nil {
				log.Println(err)
				break
			}
			state.Streaming = true
			log.Println("enabled streaming.")
			break
		}
		err = session.UpdateStreamingStatus(0, "", "")
		if err != nil {
			log.Println(err)
			break
		}
		session.UpdateStatus(0, "Ultima-Shards: AOS")
		state.Streaming = false
		log.Println("disabled streaming.")

	case "user":
		fallthrough
	case "admin":
		if channel == nil {
			err = errors.New("set channel first")
			break
		}
		user, err = state.UserFind(input.Args[0])
		if err != nil {
			break
		}
		state.Admin = true
		state.User = user

	case "channel":
		channel, err = state.ChannelFind(input.Args[0])
		if err != nil {
			break
		}
		state.Channel = channel

	case "msg":
		if channel == nil {
			err = errors.New("set channel first")
			break
		}
		_, _ = session.ChannelMessageSend(channel.ID, strings.Join(input.Args, " "))
	case "privmsg":
		if input.Length < 2 || input.Text == "" {
			err = fmt.Errorf("bad privmsg")
			break
		}
		recep, err := state.UserFind(input.Args[0])
		if err != nil {
			break
		}
		c, _ := session.UserChannelCreate(recep.ID)
		_, _ = session.ChannelMessageSend(c.ID, strings.Join(input.Args[1:], " "))

	default:
		if state.Admin == true {
			res := inputParser(state)
			if state.Sendmsg {
				_, _ = session.ChannelMessageSend(state.Channel.ID, res.Sndmsg)
			} else {
				err = fmt.Errorf(res.Errmsg)
			}
			break
		} else if channel != nil && user != nil {
			break
		}
		err = fmt.Errorf("Set channel and/or admin")

	case "send":
		if state.Sendmsg == true {
			state.Sendmsg = false
		} else {
			state.Sendmsg = true
		}

		tmp := fmt.Sprintf("send to channel: %s", strconv.FormatBool(state.Sendmsg))
		err = fmt.Errorf(tmp)

	// kill-server will be used by remote consoles
	case "kill-server":
		fallthrough
	case "quit":
		fallthrough
	case "exit":
		state.Session.Close()
		cleanup()
		os.Exit(0)
	}
	return
}
