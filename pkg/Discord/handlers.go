package discord

import (
	"github.com/bwmarrin/discordgo"
)

func (a *Application) AddCurrency(s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.Type {
	case discordgo.InteractionApplicationCommand:
	case discordgo.InteractionApplicationCommandAutocomplete:
	}
	// implementation
}
func (a *Application) RemoveCurrency(s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.Type {
	case discordgo.InteractionApplicationCommand:
	case discordgo.InteractionApplicationCommandAutocomplete:
	}
}
