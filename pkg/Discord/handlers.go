package discord

import (
	"strings"

	"github.com/bwmarrin/discordgo"
)

func (a *Application) TrackToken(s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.Type {
	case discordgo.InteractionApplicationCommandAutocomplete:
		data := i.ApplicationCommandData().Options[0].StringValue()

		a.GuildMapMutex.RLock() // Acquire a write lock
		guild := a.GuildMap[i.GuildID]
		a.GuildMapMutex.RUnlock() // Release the write lock

		event := Event{
			Type:       Autocomplete,
			Guild:      guild,
			ACValue:    data,
			ACType:     AddCurr,
			ACResponse: make(chan []*discordgo.ApplicationCommandOptionChoice),
		}

		a.Event <- event

		responseData := <-event.Response
		choices := <-event.ACResponse
		close(event.ACResponse)
		close(event.Response)

		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionApplicationCommandAutocompleteResult,
			Data: &discordgo.InteractionResponseData{
				Choices: choices, // This is basically the whole purpose of autocomplete interaction - return custom options to the user.
			},
		})
		if err != nil {
			a.LogError("error generating interaction response", err.Error())
		}

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
			Type:     TrackToken,
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

func (a *Application) RemoveToken(s *discordgo.Session, i *discordgo.InteractionCreate) {
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

func (a *Application) TrackOther(s *discordgo.Session, i *discordgo.InteractionCreate) {

}

func (a *Application) RemoveOther(s *discordgo.Session, i *discordgo.InteractionCreate) {}
