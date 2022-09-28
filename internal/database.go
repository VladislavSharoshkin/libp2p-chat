package internal

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"libp2p-chat/misk"
)

var DB *sql.DB

func DatabaseConnect() {
	db, err := sql.Open("sqlite3", "./db.db")
	misk.PanicOnError(err)
	DB = db
}
