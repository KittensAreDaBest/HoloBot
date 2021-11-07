package holobot

import (
	"github.com/apex/log"
	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/session"
	"github.com/tacticalcatto/holobot/internal/config"
	"strings"
)

type HoloBot struct {
	Session *session.Session
	Me      *discord.User
	Config  config.Configuration
}

var commands = []api.CreateCommandData{
	{
		Name:        "source",
		Description: "Get GitHub repository URL for this bot.",
	},
	{
		Name:        "schedule",
		Description: "Get Hololive schedule.",
		Options: []discord.CommandOption{
			{
				Name:        "timezone",
				Description: "This will override timezone specified in configuration.",
				Type:        discord.StringOption,
				Required:    false,
			},
		},
	},
}

func (bot *HoloBot) OnCommand(e *gateway.InteractionCreateEvent) {
	switch data := e.Data.(type) {
	case *discord.CommandInteractionData:
		switch data.Name {
		case "schedule":
			bot.handleScheduleCmd(e, data)
		case "source":
			bot.handleSourceCmd(e, data)
		}
	case *discord.ComponentInteractionData:
		command := strings.SplitN(data.CustomID, ".", 2)

		switch command[0] {
		case "schedule":
			bot.handleScheduleComponent(e, data, command[1])
		}
	}
}

func (bot *HoloBot) LoadCommands() error {
	for _, command := range commands {
		log.WithField("command", command.Name).Info("loading command")

		if _, err := bot.Session.CreateCommand(discord.AppID(bot.Me.ID), command); err != nil {
			return err
		}
	}
	return nil
}
