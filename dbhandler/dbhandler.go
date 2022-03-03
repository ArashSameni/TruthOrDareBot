package dbhandler

import (
	"database/sql"
	"errors"
)

const NIL = -1

var (
	ErrDBNotInitialized = errors.New("database is not initialized")
)

var db *sql.DB

func InitDB() error {
	var err error
	db, err = sql.Open("sqlite3", "ToD_DB.sqlite")
	return err
}
