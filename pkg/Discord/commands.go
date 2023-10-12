package discord

import "github.com/bwmarrin/discordgo"

var RawCommands = []*discordgo.ApplicationCommand{{
	Name:        "add-currency",
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
		Name:        "remove-currency",
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
		Name:        "add-other",
		Description: "Add other statical data to the list",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:        "data-type",
				Type:        discordgo.ApplicationCommandOptionString,
				Description: "The type of data to add",
				Required:    true,
				Choices: []*discordgo.ApplicationCommandOptionChoice{
					{
						Name:  "Defi Market Cap",
						Value: "defi_market_cap",
					},
					{
						Name:  "Eth Market Cap",
						Value: "eth_market_cap",
					},
					{
						Name:  "DeFi To Eth Ratio",
						Value: "defi_to_eth_ratio",
					},
					{
						Name:  "Trading Volume 24H",
						Value: "trading_volume_24h",
					},
					{
						Name:  "Defi Dominance",
						Value: "defi_dominance",
					},
					{
						Name:  "Top DeFi Token Name",
						Value: "top_defi_token_name",
					},
					{
						Name:  "Top DeFi Token Dominance",
						Value: "top_defi_token_dominance",
					},
				},
			},
		},
	}}
