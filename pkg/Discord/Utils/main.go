package utils

import "github.com/bwmarrin/discordgo"

var DefaultOtherChoices = []*discordgo.ApplicationCommandOptionChoice{
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
}
