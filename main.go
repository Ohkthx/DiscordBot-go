package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/bwmarrin/discordgo"
)

var (
	_version   string
	db         *sql.DB            // SQL database - global
	dSession   *discordgo.Session // Discord session - global
	discordLog *log.Logger        // Logs Discord request actions - global
	errLog     *log.Logger        // Logs Err information such as SQL - global
)

func cleanup() {
	fmt.Println("Cleaning up...")
	dSession.Close()
	db.Close()
	os.Exit(0)
}

func main() {
	_version = "1.0.0"
	user := flag.Bool("user", false, "use username/password")
	id := flag.String("u", "", "username")
	pwd := flag.String("p", "", "password")
	token := flag.String("t", "", "token")
	debug := flag.Bool("debug", false, "enable debug mode")
	flag.Parse()

	dSession = setup(*debug, *user, *id, *pwd, *token)

	// Register messageCreate as a callback for the messageCreate events.
	dSession.AddHandler(messageHandler)

	// Open the websocket and begin listening.
	err := dSession.Open()
	if err != nil {
		errLog.Fatal("Error discrod opening connection:", err)
		return
	}

	log.Println("Bot is now running.  Press CTRL-C to exit.")
	// Simple way to keep program running until CTRL-C is pressed.
	//<-make(chan struct{})

	core()

}
