package db

import (
	"database/sql"
	"errors"
	_ "modernc.org/sqlite"
	"os"
)

const schema = `
CREATE TABLE scheduler (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    date CHAR(8) NOT NULL DEFAULT "",
    title VARCHAR(255) NOT NULL DEFAULT "",
    comment TEXT NOT NULL DEFAULT "",
    repeat VARCHAR(128) NOT NULL DEFAULT ""
);
CREATE INDEX IF NOT EXISTS scheduler_date_idx ON scheduler (date);
`

var DB *sql.DB

func Init(dbFile string) error {
	install := false
	if _, err := os.Stat(dbFile); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			install = true
		} else {
			return err
		}
	}

	db, err := sql.Open("sqlite", dbFile)
	if err != nil {
		return err
	}

	if install {
		if _, err := db.Exec(schema); err != nil {
			_ = db.Close()
			return err
		}
	}

	DB = db
	return nil
}
