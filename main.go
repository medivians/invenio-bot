package main

import (
	"log"
	"os"
	"os/signal"

	"github.com/lramosduarte/god-sql/discord"
	"github.com/lramosduarte/god-sql/scraper/medivia"
	"github.com/lramosduarte/god-sql/scraper/wiki"
)

func main() {
	mediviaCli := medivia.New()
	cli, err := discord.Start(mediviaCli, mediviaCli, wiki.New())
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
