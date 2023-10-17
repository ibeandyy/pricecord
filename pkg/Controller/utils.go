package controller

import (
	"github.com/bwmarrin/discordgo"
	discord "pricecord/pkg/Discord"
	http "pricecord/pkg/HTTP"
)

func ConvertTokenToChoice(token []http.Token) []*discordgo.ApplicationCommandOptionChoice {
	var choices []*discordgo.ApplicationCommandOptionChoice
	for _, t := range token {
		choices = append(choices, &discordgo.ApplicationCommandOptionChoice{
			Name:  t.Name,
			Value: t.ID,
		})
	}
	return choices

}

func DefaultTokenCheck(list []*discordgo.ApplicationCommandOptionChoice, defaultTokens []http.Token) []*discordgo.ApplicationCommandOptionChoice {
	if len(list) > 0 {
		return list
	} else {
		return ConvertTokenToChoice(defaultTokens)
	}

}

func OutputConfiguredTokens(input string, g discord.GuildConfiguration) []http.Token {
	if input == "" {
		return g.ConfiguredTokens
	}

	return nil
}
