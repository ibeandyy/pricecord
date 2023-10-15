package discord

import (
	"github.com/bwmarrin/discordgo"
	utils "pricecord/pkg/Discord/Utils"
)

var RawCommands = []*discordgo.ApplicationCommand{{
	Name:        "track-token",
	Description: "Add a crypto currency to the list",
	Options: []*discordgo.ApplicationCommandOption{
		{
			Name:         "symbol",
			Type:         discordgo.ApplicationCommandOptionString,
			Description:  "The token symbol to track",
			Required:     true,
			Autocomplete: true,
		},
	},
},
	{
		Name:        "remove-token",
		Description: "Remove a crypto currency from the list",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:         "symbol",
				Type:         discordgo.ApplicationCommandOptionString,
				Description:  "The token symbol to remove",
				Required:     true,
				Autocomplete: true,
			},
		},
	},
	{
		Name:        "track-other",
		Description: "Add other statical data to the list",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:         "data-type",
				Type:         discordgo.ApplicationCommandOptionString,
				Description:  "The type of data to add",
				Required:     true,
				Choices:      utils.DefaultOtherChoices,
				Autocomplete: true,
			},
		},
	},
	{
		Name:        "remove-other",
		Description: "Add other statical data to the list",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:         "data-type",
				Type:         discordgo.ApplicationCommandOptionString,
				Description:  "The type of data to add",
				Required:     true,
				Autocomplete: true,
			},
		},
	},
}
