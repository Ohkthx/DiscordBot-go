package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
)

const _version = "0.3.2"

var (
	db     *sql.DB     // SQL database - global
	errLog *log.Logger // Logs Err information such as SQL - global
	dmLog  *log.Logger // DM = PM logger
	debug  *bool       // If debug is enabled

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

	core(dSession)

}
