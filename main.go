package main

import (
	"log"
	"os"
	"todo/pkg/db"
	"todo/pkg/server"
)

func main() {
	dbFile := "scheduler.db"
	if envDBFile := os.Getenv("TODO_DBFILE"); envDBFile != "" {
		dbFile = envDBFile
	}
	if err := db.Init(dbFile); err != nil {
		log.Fatal(err)
	}
	if err := server.Run("web"); err != nil {
		log.Fatal(err)
	}
}
