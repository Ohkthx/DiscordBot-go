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

func setupSQL() (err error) {

	// CREATE TABLE IF NOT EXISTS messages (msgid varchar(26), type varchar(32), name varchar(40))"

	q := [...]string{
		"CREATE TABLE IF NOT EXISTS `battlegrounds` (`id` bigint(20) DEFAULT NULL,`msgid` varchar(26) CHARACTER SET utf8 DEFAULT NULL,`name` varchar(26) CHARACTER SET utf8 DEFAULT NULL) ENGINE=InnoDB DEFAULT CHARSET=utf8;",
		"CREATE TABLE IF NOT EXISTS `blacklist` (`name` varchar(40) CHARACTER SET utf8 DEFAULT NULL,`status` tinyint(1) DEFAULT '0',`reports` bigint(20) DEFAULT '0',`times` bigint(20) DEFAULT '0',`start_date` date DEFAULT NULL,`who` varchar(40) CHARACTER SET utf8 DEFAULT NULL) ENGINE=InnoDB DEFAULT CHARSET=utf8;",
		"CREATE TABLE IF NOT EXISTS `channels` (`id` varchar(26) CHARACTER SET utf8 DEFAULT NULL,`name` varchar(40) CHARACTER SET utf8 DEFAULT NULL,`msg_id` varchar(26) CHARACTER SET utf8 DEFAULT NULL,`msg_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,`amount` bigint(20) DEFAULT '0') ENGINE=InnoDB DEFAULT CHARSET=utf8;",
		"CREATE TABLE IF NOT EXISTS `commands` (`id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,`command` varchar(20) CHARACTER SET utf8 DEFAULT NULL,`arg1` varchar(20) CHARACTER SET utf8 DEFAULT NULL,`arg2` varchar(20) CHARACTER SET utf8 DEFAULT NULL,`arg3` varchar(20) CHARACTER SET utf8 DEFAULT NULL,`text` varchar(2000) CHARACTER SET utf8 DEFAULT NULL,`description` varchar(255) CHARACTER SET utf8 DEFAULT NULL,`author` varchar(32) CHARACTER SET utf8 DEFAULT NULL,`date_added` date DEFAULT NULL,`author_mod` varchar(32) CHARACTER SET utf8 DEFAULT NULL,`date_mod` date DEFAULT NULL,`count` bigint(20) DEFAULT NULL,UNIQUE KEY `id` (`id`)) ENGINE=InnoDB DEFAULT CHARSET=utf8;",
		"CREATE TABLE IF NOT EXISTS `config` (`id` varchar(40) CHARACTER SET utf8 DEFAULT NULL,`name` varchar(40) CHARACTER SET utf8 DEFAULT NULL) ENGINE=InnoDB DEFAULT CHARSET=utf8;",
		"CREATE TABLE IF NOT EXISTS `events` (`id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,`weekday` varchar(20) CHARACTER SET utf8 DEFAULT NULL,`time` varchar(16) DEFAULT NULL,UNIQUE KEY `id` (`id`)) ENGINE=InnoDB DEFAULT CHARSET=utf8;",
		"CREATE TABLE IF NOT EXISTS `library` (`id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,`name` varchar(40) CHARACTER SET utf8 DEFAULT NULL,`script` varchar(2000) CHARACTER SET utf8 DEFAULT NULL,UNIQUE KEY `id` (`id`)) ENGINE=InnoDB DEFAULT CHARSET=utf8;",
		"CREATE TABLE IF NOT EXISTS `messages` (`msgid` varchar(26) CHARACTER SET utf8 DEFAULT NULL,`type` varchar(32) CHARACTER SET utf8 DEFAULT NULL,`name` varchar(40) CHARACTER SET utf8 DEFAULT NULL) ENGINE=InnoDB DEFAULT CHARSET=utf8;",
		"CREATE TABLE IF NOT EXISTS `permissions` (`id` varchar(26) CHARACTER SET utf8 DEFAULT NULL,`username` varchar(40) CHARACTER SET utf8 DEFAULT NULL,`allow` tinyint(1) DEFAULT '0',`date_added` date DEFAULT NULL,`accountable` varchar(40) CHARACTER SET utf8 DEFAULT NULL) ENGINE=InnoDB DEFAULT CHARSET=utf8;",
		"CREATE TABLE IF NOT EXISTS `subscriptions` (`id` varchar(40) CHARACTER SET utf8 DEFAULT NULL,`name` varchar(255) CHARACTER SET utf8 DEFAULT NULL,`sub` tinyint(1) DEFAULT NULL) ENGINE=InnoDB DEFAULT CHARSET=utf8;",
		"CREATE TABLE IF NOT EXISTS `tokens` (`id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,`name` varchar(32) CHARACTER SET utf8 DEFAULT NULL,`token` varchar(64) CHARACTER SET utf8 DEFAULT NULL,UNIQUE KEY `id` (`id`)) ENGINE=InnoDB DEFAULT CHARSET=utf8;",
		"CREATE TABLE IF NOT EXISTS `users` (`id` varchar(26) CHARACTER SET utf8 DEFAULT NULL,`username` varchar(255) CHARACTER SET utf8 DEFAULT NULL,`discriminator` varchar(5) CHARACTER SET utf8 DEFAULT NULL,`msg_last` timestamp NOT NULL,`msg_count` bigint(20) DEFAULT '0',`status` tinyint(1) DEFAULT '0',`vote` varchar(3) CHARACTER SET utf8 DEFAULT NULL) ENGINE=InnoDB DEFAULT CHARSET=utf8;",
	}

	for _, query := range q {
		_, err = db.Exec(query)
		if err != nil {
			log.Fatal(err)
		}
	}

	return nil
}
