package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	var err error
	fmt.Println("DATABASE UPGRADE to schema for GMA server versions 5.34.0 and later")

	if len(os.Args) != 2 {
		log.Fatal("Requires a database name as the argument to the command.")
	}
	dbname := os.Args[1]
	if _, err = os.Stat(dbname); os.IsNotExist(err) {
		log.Fatalf("database file \"%s\" does not exist", dbname)
	}
	log.Printf("Upgrading database file \"%s\"", dbname)

	db, err := sql.Open("sqlite3", "file:"+dbname)
	if err != nil {
		log.Fatalf("unable to open sqlite3 database \"%s\": %v", dbname, err)
	}

	var result sql.Result
	var rows *sql.Rows
	var row *sql.Row
	updateList := make(map[int]string)

	row = db.QueryRow(`select sender from chats limit 1`)
	if row.Err() == nil {
		log.Println("this database appears to already have a sender column")
	} else {
		if result, err = db.Exec(`alter table chats add column sender text`); err != nil {
			log.Fatalf("error adding new column to database: %v", err)
		}
		affected, err := result.RowsAffected()
		if err != nil {
			log.Fatalf("error checking operation result: %v", err)
		}
		log.Printf("updated chats table; rows affected=%d", affected)
	}

	log.Print("reading existing messages...")
	if rows, err = db.Query(`select msgid, msgtype, rawdata from chats`); err != nil {
		log.Fatalf("error reading databasae: %v", err)
	}
	defer rows.Close()
	for rows.Next() {
		var messageID, messageType int
		var rawData string
		var sender string
		type senderReceiver struct {
			RequestedBy string
			Sender      string
		}
		var name senderReceiver

		if err = rows.Scan(&messageID, &messageType, &rawData); err != nil {
			log.Fatalf("error reading result row: %v", err)
		}
		if err = json.Unmarshal([]byte(rawData), &name); err != nil {
			log.Fatalf("message id=%d raw data cannot be decoded: %v; data=%v", messageID, err, rawData)
		}
		switch messageType {
		case 0:
			sender = name.RequestedBy
			log.Printf("id=%d, type=CC, sender=%s, message=%s", messageID, sender, rawData)
		case 1:
			sender = name.Sender
			log.Printf("id=%d, type=TO, sender=%s, message=%s", messageID, sender, rawData)
		case 2:
			sender = name.Sender
			log.Printf("id=%d, type=ROLL, sender=%s, message=%s", messageID, sender, rawData)
		default:
			log.Fatalf("id=%d, type=%d: unsupported type", messageID, messageType)
		}
		updateList[messageID] = sender
	}

	log.Printf("read %d message%s from database", len(updateList), func(n int) string {
		if n == 1 {
			return ""
		} else {
			return "s"
		}
	}(len(updateList)))
	log.Print("proceeding to update the above entries...")
	i := 0
	for messageID, sender := range updateList {
		result, err = db.Exec(`update chats set sender=? where msgid=?`, sender, messageID)
		if err != nil {
			log.Fatalf("**ERROR** updating database record for message id=%d: %v", messageID, err)
		}
		affected, err := result.RowsAffected()
		if err != nil {
			log.Printf("WARNING unable to see results of updating database record for message id=%d: %v", messageID, err)
		} else if affected != 1 {
			log.Printf("WARNING message id=%d, sender=%s, rows affected=%d", messageID, sender, affected)
		}

		i++
		if (i % 100) == 0 {
			log.Printf("updated %d/%d", i, len(updateList))
		}
	}
	log.Printf("updated %d/%d", i, len(updateList))
	log.Println("Operation completed.")

	db.Close()
}
