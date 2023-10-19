package discord

import (
	"strings"

	"github.com/bwmarrin/discordgo"
)

func (a *Application) TrackToken(s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.Type {

	case discordgo.InteractionApplicationCommandAutocomplete:
		a.LogRequest("TrackToken autocomplete request", i.ApplicationCommandData().Options[0].StringValue())
		a.handleAutoComplete(s, i, AddCurr)

	case discordgo.InteractionApplicationCommand:
		a.LogRequest("TrackToken command request", i.ApplicationCommandData().Options[0].StringValue())
		tkn := i.ApplicationCommandData().Options[0].StringValue()
		tkn = strings.ToUpper(tkn)

		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{Flags: discordgo.MessageFlagsEphemeral},
		})
		if err != nil {
			a.LogError("error generating interaction response", err.Error())
		}
		guild := a.getGuildReadLock(i.GuildID)

		event := Event{
			Type:     TrackToken,
			Guild:    guild,
			Name:     tkn,
			Response: make(chan bool),
		}
		a.LogRequest("sending track token event")
		a.Event <- event
		responseData := <-event.Response
		a.LogRequest("track token response received")
		close(event.Response)
		if !responseData {
			_, err := s.FollowupMessageCreate(i.Interaction, false, &discordgo.WebhookParams{
				Content: "Token " + tkn + " already being tracked or doesn't exist in coingecko API.",
			})
			if err != nil {
				a.LogError("error responding to interaction %v", err.Error())
				return
			}
			return
		}

		_, err = s.FollowupMessageCreate(i.Interaction, false, &discordgo.WebhookParams{
			Content: "Token " + tkn + " added to tracking list.",
		})
	}

}

func (a *Application) RemoveToken(s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		a.LogRequest("RemoveToken command request", i.ApplicationCommandData().Options[0].StringValue())
		tkn := i.ApplicationCommandData().Options[0].StringValue()
		tkn = strings.ToUpper(tkn)

		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{Flags: discordgo.MessageFlagsEphemeral},
		})
		if err != nil {
			a.LogError("error generating interaction response", err.Error())
		}

		guild := a.getGuildReadLock(i.GuildID)

		event := Event{
			Type:     RemoveToken,
			Guild:    guild,
			Name:     tkn,
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

	case discordgo.InteractionApplicationCommandAutocomplete:
		a.LogRequest("RemoveToken autocomplete request", i.ApplicationCommandData().Options[0].StringValue())
		a.handleAutoComplete(s, i, RemCurr)
	}
}

func (a *Application) TrackOther(s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.Type {
	case discordgo.InteractionApplicationCommandAutocomplete:
		a.LogRequest("TrackOther autocomplete request", i.ApplicationCommandData().Options[0].StringValue())
		a.handleAutoComplete(s, i, AddOther)
	case discordgo.InteractionApplicationCommand:
		a.LogRequest("TrackOther command request", i.ApplicationCommandData().Options[0].StringValue())
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{Flags: discordgo.MessageFlagsEphemeral},
		})
		if err != nil {
			a.LogError("error generating interaction response", err.Error())
		}
		newStat := i.ApplicationCommandData().Options[0].StringValue()

		guild := a.getGuildReadLock(i.GuildID)

		event := Event{
			Type:     TrackOther,
			Guild:    guild,
			Name:     newStat,
			Response: make(chan bool),
		}

		a.Event <- event
		responseData := <-event.Response
		close(event.Response)
		if !responseData {
			_, err := s.FollowupMessageCreate(i.Interaction, false, &discordgo.WebhookParams{
				Content: "Statistic " + newStat + " already being tracked or doesn't exist in coingecko API.",
			})
			if err != nil {
				a.LogError("error responding to interaction %v", err.Error())
				return
			}
		}
	}
}

func (a *Application) RemoveOther(s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.Type {
	case discordgo.InteractionApplicationCommandAutocomplete:
		a.LogRequest("RemoveOther autocomplete request", i.ApplicationCommandData().Options[0].StringValue())
		a.handleAutoComplete(s, i, RemOther)
	case discordgo.InteractionApplicationCommand:
		a.LogRequest("RemoveOther command request", i.ApplicationCommandData().Options[0].StringValue())
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{Flags: discordgo.MessageFlagsEphemeral},
		})
		if err != nil {
			a.LogError("error generating interaction response", err.Error())
		}
		newStat := i.ApplicationCommandData().Options[0].StringValue()

		guild := a.getGuildReadLock(i.GuildID)

		event := Event{
			Type:     RemoveOther,
			Guild:    guild,
			Name:     newStat,
			Response: make(chan bool),
		}
		a.Event <- event
		responseData := <-event.Response
		close(event.Response)
		if !responseData {
			_, err := s.FollowupMessageCreate(i.Interaction, false, &discordgo.WebhookParams{
				Content: "Statistic " + newStat + " already being tracked or doesn't exist in coingecko API.",
			})
			if err != nil {
				a.LogError("error responding to interaction %v", err.Error())
				return
			}
		}
	}
}

func (a *Application) handleAutoComplete(s *discordgo.Session, i *discordgo.InteractionCreate, acType AutocompleteType) {
	a.LogRequest("handling autocomplete request", i.ApplicationCommandData().Options[0].StringValue())
	data := i.ApplicationCommandData().Options[0].StringValue()

	guild := a.getGuildReadLock(i.GuildID)

	event := Event{
		Type:       Autocomplete,
		Guild:      guild,
		ACValue:    data,
		ACType:     acType,
		Response:   make(chan bool),
		ACResponse: make(chan []*discordgo.ApplicationCommandOptionChoice),
	}
	a.LogRequest("sending autocomplete event")
	a.Event <- event
	a.LogRequest("waiting for autocomplete response")
	//TODO:VERIFY IF NEEDED
	choices := <-event.ACResponse
	close(event.ACResponse)
	a.LogRequest("autocomplete response 1 received")

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionApplicationCommandAutocompleteResult,
		Data: &discordgo.InteractionResponseData{
			Choices: choices,
		},
	})
	if err != nil {
		a.LogError("error generating interaction response", err.Error())
	}
}

func (a *Application) getGuildReadLock(guildID string) *GuildConfiguration {
	a.GuildMapMutex.RLock()
	guild := a.GuildMap[guildID]
	a.GuildMapMutex.RUnlock()
	return guild
}
