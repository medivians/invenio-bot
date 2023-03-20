package main

import (
	"log"
	"os"
	"os/signal"

	"github.com/medivians/invenio-bot/discord"
	"github.com/medivians/invenio-bot/scraper/medivia"
)

func main() {
	mediviaCli := medivia.New()
	cli, err := discord.Start(mediviaCli, mediviaCli)
	if err != nil {
		log.Fatalf("starting discord bot %q", err)
	}
	defer func() {
		cli.Close()
		log.Println("Stoping discord bot")
	}()
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop
	log.Println("Graceful shutdown")
}
