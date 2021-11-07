package holobot

import (
	"fmt"
	"github.com/apex/log"
	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/tacticalcatto/holobot/internal/hololive"
	"math"
	"regexp"
	"strconv"
	"strings"
)

func (bot *HoloBot) handleSourceCmd(e *gateway.InteractionCreateEvent, d *discord.CommandInteractionData) {
	if err := bot.Session.RespondInteraction(e.ID, e.Token, api.InteractionResponse{
		Type: api.MessageInteractionWithSource,
		Data: &api.InteractionResponseData{
			Embeds: &[]discord.Embed{
				{
					Title:       "**GitHub Repository**",
					Description: "https://github.com/tacticalcatto/holobot",
				},
			},
		},
	}); err != nil {
		log.WithError(err).Error("failed to send interaction callback")
	}
}

func (bot *HoloBot) handleScheduleCmd(e *gateway.InteractionCreateEvent, d *discord.CommandInteractionData) {
	timezone := bot.Config.Timezone
	if len(d.Options) == 1 {
		timezone = d.Options[0].String()
	}

	schedule, err := hololive.GetSchedule(timezone)
	if err != nil {
		log.WithError(err).Error("failed to get schedule from the api")

		if err := bot.Session.RespondInteraction(e.ID, e.Token, api.InteractionResponse{
			Type: api.MessageInteractionWithSource,
			Data: &api.InteractionResponseData{
				Embeds: &[]discord.Embed{
					{
						Title:       "**Hololive Schedule**",
						Description: "Failed to get schedule from the API.",
					},
				},
			},
		}); err != nil {
			log.WithError(err).Error("failed to send interaction callback")
		}
		return
	}

	if len(schedule.Streams) == 0 {
		if err := bot.Session.RespondInteraction(e.ID, e.Token, api.InteractionResponse{
			Type: api.MessageInteractionWithSource,
			Data: &api.InteractionResponseData{
				Embeds: &[]discord.Embed{
					{
						Title:       "**Hololive Schedule**",
						Description: "No one is streaming.",
					},
				},
			},
		}); err != nil {
			log.WithError(err).Error("failed to send interaction callback")
		}
		return
	}

	fields := streamFields(schedule.Streams)
	if len(schedule.Streams) > 5 {
		fields = fields[:5]
	}

	pages := int(math.Ceil(float64(len(schedule.Streams)) / float64(5)))

	if err := bot.Session.RespondInteraction(e.ID, e.Token, api.InteractionResponse{
		Type: api.MessageInteractionWithSource,
		Data: &api.InteractionResponseData{
			Embeds: &[]discord.Embed{
				{
					Title:       "**Hololive Schedule**",
					Description: fmt.Sprintf("Live: %d\nWaiting: %d\nTotal: %d", schedule.Stats.Live, schedule.Stats.Waiting, schedule.Stats.Total),
					Fields:      fields,
					Footer: &discord.EmbedFooter{
						Text: fmt.Sprintf("Page 1 of %d", pages),
					},
				},
			},
			Components: &[]discord.Component{
				schedulePaginateButtons(timezone),
			},
		},
	}); err != nil {
		log.WithError(err).Error("failed to send interaction callback")
	}
}

func (bot *HoloBot) handleScheduleComponent(e *gateway.InteractionCreateEvent, d *discord.ComponentInteractionData, cmd string) {
	split := strings.SplitN(cmd, ".", 2)
	timezone := bot.Config.Timezone
	if split[1] != "" {
		timezone = split[1]
	}

	embed := e.Message.Embeds[0]
	matches := regexp.MustCompile(`Page (\d+) of (\d+)`).FindStringSubmatch(embed.Footer.Text)

	if len(matches) != 3 {
		return
	}

	page, _ := strconv.Atoi(matches[1])

	switch split[0] {
	case "previous":
		page--
	case "next":
		page++
	case "exit":
		if err := bot.Session.RespondInteraction(e.ID, e.Token, api.InteractionResponse{
			Type: api.UpdateMessage,
			Data: &api.InteractionResponseData{
				Components: &[]discord.Component{},
			},
		}); err != nil {
			log.WithError(err).Error("failed to send interaction callback")
		}
		return
	}

	schedule, err := hololive.GetSchedule(timezone)
	if err != nil {
		log.WithError(err).Error("failed to get schedule from the api")

		if err := bot.Session.RespondInteraction(e.ID, e.Token, api.InteractionResponse{
			Type: api.UpdateMessage,
			Data: &api.InteractionResponseData{
				Embeds: &[]discord.Embed{
					{
						Title:       "**Hololive Schedule**",
						Description: "Failed to get schedule from the API.",
					},
				},
			},
		}); err != nil {
			log.WithError(err).Error("failed to send interaction callback")
		}
		return
	}

	fields := streamFields(schedule.Streams)
	pages := int(math.Ceil(float64(len(schedule.Streams)) / float64(5)))

	if page < 1 || page > pages {
		bot.Session.RespondInteraction(e.ID, e.Token, api.InteractionResponse{
			Type: api.UpdateMessage,
			Data: &api.InteractionResponseData{},
		})
		return
	}

	if page != pages {
		fields = fields[(page-1)*5 : page*5]
	} else {
		fields = fields[(page-1)*5:]
	}

	if err := bot.Session.RespondInteraction(e.ID, e.Token, api.InteractionResponse{
		Type: api.UpdateMessage,
		Data: &api.InteractionResponseData{
			Embeds: &[]discord.Embed{
				{
					Title:       "**Hololive Schedule**",
					Description: fmt.Sprintf("Live: %d\nWaiting: %d\nTotal: %d", schedule.Stats.Live, schedule.Stats.Waiting, schedule.Stats.Total),
					Fields:      fields,
					Footer: &discord.EmbedFooter{
						Text: fmt.Sprintf("Page %d of %d", page, pages),
					},
				},
			},
			Components: &[]discord.Component{
				schedulePaginateButtons(timezone),
			},
		},
	}); err != nil {
		log.WithError(err).Error("failed to send interaction callback")
	}
}

func streamFields(streams []hololive.HololiveStream) []discord.EmbedField {
	var fields []discord.EmbedField

	for _, stream := range streams {
		fields = append(fields, stream.ToDiscordField())
	}

	return fields
}

func schedulePaginateButtons(timezone string) *discord.ActionRowComponent {
	return &discord.ActionRowComponent{
		Components: []discord.Component{
			&discord.ButtonComponent{
				Label:    "Previous",
				CustomID: fmt.Sprintf("schedule.previous.%s", timezone),
				Style:    discord.PrimaryButton,
			},
			&discord.ButtonComponent{
				Label:    "Next",
				CustomID: fmt.Sprintf("schedule.next.%s", timezone),
				Style:    discord.PrimaryButton,
			},
			&discord.ButtonComponent{
				Label:    "Exit",
				CustomID: fmt.Sprintf("schedule.exit.%s", timezone),
				Style:    discord.DangerButton,
			},
		},
	}
}
