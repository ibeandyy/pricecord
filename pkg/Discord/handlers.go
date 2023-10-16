package discord

import (
	"strings"

	"github.com/bwmarrin/discordgo"
)

func (a *Application) TrackToken(s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.Type {

	case discordgo.InteractionApplicationCommandAutocomplete:
		a.handleAutoComplete(s, i, AddCurr)

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
		guild := a.getGuildReadLock(i.GuildID)

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

	case discordgo.InteractionApplicationCommandAutocomplete:
		a.handleAutoComplete(s, i, RemCurr)
	}
}

func (a *Application) TrackOther(s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.Type {
	case discordgo.InteractionApplicationCommandAutocomplete:
		a.handleAutoComplete(s, i, AddOther)
	case discordgo.InteractionApplicationCommand:
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
			Symbol:   newStat,
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
		a.handleAutoComplete(s, i, RemOther)
	case discordgo.InteractionApplicationCommand:
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
			Symbol:   newStat,
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
	data := i.ApplicationCommandData().Options[0].StringValue()

	guild := a.getGuildReadLock(i.GuildID)

	event := Event{
		Type:       Autocomplete,
		Guild:      guild,
		ACValue:    data,
		ACType:     acType,
		ACResponse: make(chan []*discordgo.ApplicationCommandOptionChoice),
	}

	a.Event <- event
	//TODO:VERIFY IF NEEDED
	responseData := <-event.Response
	choices := <-event.ACResponse
	close(event.ACResponse)
	close(event.Response)

	if !responseData {
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionApplicationCommandAutocompleteResult,
			Data: &discordgo.InteractionResponseData{
				Choices: choices,
			},
		})
		if err != nil {
			a.LogError("error generating interaction response", err.Error())
			return
		}
	}
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

func (a *Application) getGuildReadLock(guildID string) GuildConfiguration {
	a.GuildMapMutex.RLock()
	guild := a.GuildMap[guildID]
	a.GuildMapMutex.RUnlock()
	return guild
}
