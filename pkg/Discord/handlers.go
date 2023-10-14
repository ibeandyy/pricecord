package discord

import (
	"github.com/bwmarrin/discordgo"
	"strings"
)

func (a *Application) AddCurrency(s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.Type {
	case discordgo.InteractionApplicationCommandAutocomplete:
		// get coingecko currency list cache
		// filter by input
		// return list of currencies
	case discordgo.InteractionApplicationCommand:
		tkn := i.ApplicationCommandData().Options[0].StringValue()
		tkn = strings.ToUpper(tkn)

		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{

			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{Flags: discordgo.MessageFlagsEphemeral, Content: "Added new token " + tkn},
		})
		if err != nil {
			a.Logger.Printf("error responding to interaction %v", err)
		}
		a.GuildMapMutex.RLock() // Acquire a read lock
		guild := a.GuildMap[i.GuildID]
		a.GuildMapMutex.RUnlock() // Release the read lock
		a.GuildMapMutex.Lock()    // Acquire a write lock
		guild.ConfiguredTokens = append(guild.ConfiguredTokens, tkn)
		a.GuildMapMutex.Unlock() // Release the write lock
		event := Event{
			Type:     AddCurrency,
			Data:     guild,
			Response: make(chan interface{}),
		}
		a.Event <- event
		responseData := <-event.Response
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
