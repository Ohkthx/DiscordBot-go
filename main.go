package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/bwmarrin/discordgo"
)

const _version = "1.1.1"

type inputDat struct {
	text     string
	command  string
	args     []string
	length   int
	modifier bool // If it is an "Add/Mod/Del" command
}

type inputInfo struct {
	admin     bool // if user and channel are set- true
	send      bool // Send data thru channel or send to console.
	user      *discordgo.User
	channel   *discordgo.Channel
	channelID string
	dat       *inputDat
}

var (
	db         *sql.DB            // SQL database - global
	dSession   *discordgo.Session // Discord session - global
	dUser      *discordgo.User
	discordLog *log.Logger // Logs Discord request actions - global
	errLog     *log.Logger // Logs Err information such as SQL - global
)

func cleanup() {
	fmt.Println("Cleaning up...")
	dSession.Close()
	db.Close()
	fmt.Println("Terminated.")
	os.Exit(0)
}

func main() {
	user := flag.Bool("user", false, "use username/password")
	id := flag.String("u", "", "username")
	pwd := flag.String("p", "", "password")
	token := flag.String("t", "", "token")
	debug := flag.Bool("debug", false, "enable debug mode")
	flag.Parse()

	var err error
	dSession = setup(*debug, *user, *id, *pwd, *token)

	// Register messageCreate as a callback for the messageCreate events.
	dSession.AddHandler(messageHandler)

	dUser, err = dSession.User("@me")
	if err != nil {
		discordLog.Println("Error obtaining account details:", err)
		return
	}

	// Open the websocket and begin listening.
	err = dSession.Open()
	if err != nil {
		errLog.Fatal("Error discord opening connection:", err)
		return
	}

	log.Println("Bot is now running.  Press CTRL-C to exit.")
	// Simple way to keep program running until CTRL-C is pressed.
	//<-make(chan struct{})

	core()

}
