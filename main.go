package main

import (
	"log"
	"todo/pkg/server"
)

func main() {
	if err := server.Run("web"); err != nil {
		log.Fatal(err)
	}
}
