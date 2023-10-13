package discord

import (
	"github.com/bwmarrin/discordgo"
)

func (a *Application) AddCurrency(s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.Type {
	case discordgo.InteractionApplicationCommandAutocomplete:
		// autocomplete
		// get coingecko currency list cache
		// filter by input
		// return list of currencies
	case discordgo.InteractionApplicationCommand:
		//verify currency is valid
		//verify currency is not already configured
		//normalize currency
		//add currency to database
		//add currency to guild map
		//add the currency to message
		//reset guildLastChecked timer
		//a.GuildMap[i.GuildID].ConfiguredTokens = append(a.GuildMap[i.GuildID].ConfiguredTokens, i.ApplicationCommandData().Options[0].StringValue())

	}
}
func (a *Application) RemoveCurrency(s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		//verify currency exists for guild
		//remove currency from database
		//remove currency from guild map
		//remove currency from message
		//reset guildLastChecked timer
	case discordgo.InteractionApplicationCommandAutocomplete:
		//fetch current guild token list
		//filter by input
		//return list of currencies
	}
}
