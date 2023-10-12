package discord

import "github.com/bwmarrin/discordgo"

var RawCommands = []*discordgo.ApplicationCommand{{
	Name:        "ping",
	Description: "Ping!",
	Options: []*discordgo.ApplicationCommandOption{
		{
			Name:        "to-ping",
			Type:        discordgo.ApplicationCommandOptionString,
			Description: "Who to ping",
			Required:    true,
		},
	},
},
}
