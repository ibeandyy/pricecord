package controller

import (
	"github.com/bwmarrin/discordgo"
	discord "pricecord/pkg/Discord"
	http "pricecord/pkg/HTTP"
)

func (c *Controller) ConvertTokenToChoice(token []*http.Token) []*discordgo.ApplicationCommandOptionChoice {
	var choices []*discordgo.ApplicationCommandOptionChoice
	if len(token) == 0 {
		token = c.DefaultTokens
	}
	for _, t := range token {
		choices = append(choices, &discordgo.ApplicationCommandOptionChoice{
			Name:  t.Name,
			Value: t.ID,
		})
		if len(choices) == 25 {
			break
		}
	}
	return choices

}

func (c *Controller) DefaultTokenCheck(list []*discordgo.ApplicationCommandOptionChoice, defaultTokens []*http.Token) []*discordgo.ApplicationCommandOptionChoice {
	if len(list) > 0 && len(list) <= 25 {
		return list
	} else {
		return c.ConvertTokenToChoice(defaultTokens)
	}

}

func OutputConfiguredTokens(input string, g *discord.GuildConfiguration) []*http.Token {
	if input == "" {
		return g.ConfiguredTokens
	}

	return nil
}
