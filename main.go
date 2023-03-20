package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/medivians/invenio-bot/discord"
	"github.com/medivians/invenio-bot/scraper/medivia"
	"github.com/medivians/invenio-bot/scraper/wiki"
)

func main() {
	mediviaCli := medivia.New()
	cli, err := discord.Start(mediviaCli, mediviaCli, wiki.New())
	if err != nil {
		log.Fatalf("starting discord bot %q", err)
	}

	go healthCheck("80")
	defer func() {
		cli.Close()
		log.Println("Stoping discord bot")
	}()
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop
	log.Println("Graceful shutdown")
}

func healthCheck(port string) {
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "ok")
	})

	err := http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
	if err != nil {
		log.Fatal(err)
	}
}
