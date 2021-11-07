package main

import (
	"context"
	"fmt"
	"github.com/apex/log"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/session"
	"github.com/tacticalcatto/holobot/internal/config"
	"github.com/tacticalcatto/holobot/internal/holobot"
	"os"
	"os/signal"
)

func main() {
	var bot holobot.HoloBot
	var err error

	log.WithField("config", "./config.yml").Info("parsing configuration file")
	bot.Config, err = config.ParseConfiguration("./config.yml")
	if err != nil {
		log.WithError(err).Fatal("failed to parse configuration file")
	}

	bot.Session, err = session.New(fmt.Sprintf("Bot %s", bot.Config.Token))
	if err != nil {
		log.WithError(err).Fatal("failed to create bot session")
	}

	bot.Session.AddHandler(bot.OnCommand)
	bot.Session.AddIntents(gateway.IntentGuilds)
	bot.Session.AddIntents(gateway.IntentGuildMessages)

	if err := bot.Session.Open(context.Background()); err != nil {
		log.WithError(err).Fatal("failed to open gateway connection")
	}

	log.Info("connection to gateway established")

	bot.Me, err = bot.Session.Me()
	if err != nil {
		log.WithError(err).Fatal("failed to get information about myself")
	}

	if err := bot.LoadCommands(); err != nil {
		log.WithError(err).Fatal("failed to load commands")
	}

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt)

	<-shutdown
	log.Info("closing the connection")
	bot.Session.Close()
}
