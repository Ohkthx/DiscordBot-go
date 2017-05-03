package main

import (
	"database/sql"
	"log"
	"os"

	"github.com/bwmarrin/discordgo"
	_ "github.com/go-sql-driver/mysql"
)

func setup(debug bool) *discordgo.Session {
	var dg *discordgo.Session
	var err error
	var dbinfo = "root@/discord"

	err = setupLogger()
	if err != nil {
		errLog.Fatal(err)
	}

	if debug == true {
		dbinfo = "root@/debug"
	}

	db, err = sql.Open("mysql", dbinfo+"?charset=utf8")
	if err != nil {
		errLog.Fatal(err)
	}

	var query string
	if debug == true {
		log.Println("Using debug auth token.")
		query = "SELECT token FROM tokens WHERE name=('d0xy')"
	} else if debug == false {
		log.Println("Using normal auth token.")
		query = "SELECT token FROM tokens WHERE name=('Gatekeeper')"
	}

	qPrep, err := db.Prepare(query)
	if err != nil {
		errLog.Fatal(err)
	}

	var token string
	err = qPrep.QueryRow().Scan(&token)
	if err != nil {
		errLog.Fatal(err)
	}

	dg, err = discordgo.New("Bot " + token)

	if err != nil {
		errLog.Fatal("An error occured creating Discord session:", err)
	}

	return dg
}

func setupLogger() error {
	errLog = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
	dmLog = log.New(os.Stderr, "", log.Ldate|log.Ltime)

	f, err := os.OpenFile("stderr.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	s, err := os.OpenFile("privmsg.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return err
	}

	dmLog.SetOutput(s)
	errLog.SetOutput(f)
	return nil
}
