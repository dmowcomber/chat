package main

import (
	"log"

	"github.com/dmowcomber/chat/server"
)

func main() {
	srv := server.NewServer(8080)
	err := srv.Run()
	if err != nil {
		log.Fatalf("error: %s", err.Error())
	}
}
