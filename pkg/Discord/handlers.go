package discord

import (
	"strings"

	"github.com/bwmarrin/discordgo"
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
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{Flags: discordgo.MessageFlagsEphemeral},
		})
		if err != nil {
			a.LogError("error generating interaction response", err.Error())
		}
		a.GuildMapMutex.RLock() // Acquire a write lock
		guild := a.GuildMap[i.GuildID]
		a.GuildMapMutex.RUnlock() // Release the write lock

		event := Event{
			Type:     AddCurrency,
			Guild:    guild,
			Symbol:   tkn,
			Response: make(chan bool),
		}

		a.Event <- event
		responseData := <-event.Response
		close(event.Response)
		if !responseData {
			_, err := s.FollowupMessageCreate(i.Interaction, false, &discordgo.WebhookParams{
				Content: "Token " + tkn + " already being tracked or doesn't exist in coingecko API.",
			})
			if err != nil {
				a.LogError("error responding to interaction %v", err.Error())
				return
			}
		}

		_, err = s.FollowupMessageCreate(i.Interaction, false, &discordgo.WebhookParams{
			Content: "Token " + tkn + " added to tracking list.",
		})
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
