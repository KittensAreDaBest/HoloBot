package hololive

import (
	"encoding/json"
	"fmt"
	"github.com/diamondburned/arikawa/v3/discord"
	"net/http"
)

type HololiveSchedule struct {
	Streams []HololiveStream `json:"streams"`
	Stats   HololiveStats    `json:"stats"`
}

type HololiveStats struct {
	Live    int `json:"live"`
	Waiting int `json:"waiting"`
	Total   int `json:"total"`
}

type HololiveStream struct {
	Group  string `json:"group"`
	Member string `json:"member"`
	Stream struct {
		URL string `json:"url"`
	} `json:"stream"`
	Time struct {
		Raw int `json:"raw"`
	} `yaml:"time"`
}

func (stream HololiveStream) ToDiscordField() discord.EmbedField {
	return discord.EmbedField{
		Name:  stream.Member,
		Value: fmt.Sprintf("%s <t:%d:R> <t:%d:t>", stream.Stream.URL, stream.Time.Raw, stream.Time.Raw),
	}
}

func GetSchedule(timezone string) (HololiveSchedule, error) {
	var schedule HololiveSchedule

	res, err := http.Get(fmt.Sprintf("https://hololive-schedule.hyperkittens.repl.co/api/v1/schedule?group=all&timezone=%s", timezone))
	if err != nil {
		return schedule, err
	}
	defer res.Body.Close()

	if err := json.NewDecoder(res.Body).Decode(&schedule); err != nil {
		return schedule, err
	}
	return schedule, nil
}
