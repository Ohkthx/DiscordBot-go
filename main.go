package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/bwmarrin/discordgo"
)

const _version = "0.2.0"
const (
	cmdADD = 1 << iota
	cmdMODIFY
	cmdDELETE
	cmdEVENT
	cmdSCRIPT
	cmdVENDOR
)

type inputDat struct {
	text    string
	command string
	args    []string
	length  int
	attr    int
}

type inputInfo struct {
	admin   bool // if user and channel are set- true
	send    bool // Send data thru channel or send to console.
	user    *discordgo.User
	channel *discordgo.Channel
	dat     *inputDat
	session *discordgo.Session
}

var (
	db     *sql.DB     // SQL database - global
	errLog *log.Logger // Logs Err information such as SQL - global
	dmLog  *log.Logger
	debug  *bool // If debug is enabled
)

func cleanup() {
	fmt.Println("Cleaning up...")
	db.Close()
	fmt.Println("Terminated.")
	os.Exit(0)
}

func main() {
	debug = flag.Bool("debug", false, "enable debug mode")
	flag.Parse()

	var err error
	dSession := setup(*debug)

	go serverInit()

	// Register messageCreate as a callback for the messageCreate events.
	dSession.AddHandler(messageHandler)

	// Open the websocket and begin listening.
	err = dSession.Open()
	if err != nil {
		errLog.Fatal("Error discord opening connection:", err)
		return
	}
	dSession.UpdateStatus(0, "Ultima-Shards: AOS")

	core(dSession)

}
