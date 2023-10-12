package discord

import (
	"github.com/bwmarrin/discordgo"
)

var HandlerMap = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){}
