package controller

import (
	"fmt"
	"log"

	"pricecord/pkg/Database"
	"pricecord/pkg/Discord"
	"pricecord/pkg/HTTP"
)

// Controller is the main entrypoint for the application
type Controller struct {
	DiscordClient *discord.Application
	Database      *database.Database
	HTTPClient    *http.Client
	Logger        *log.Logger
	CoinCache     map[string]http.Coin
}

func (c *Controller) ListenToEvents() {
	for event := range c.DiscordClient.Event {
		switch event.Type {
		case discord.AddCurrency:
			tkn, gTkns := event.Symbol, event.Guild.ConfiguredTokens

			for _, gTkn := range gTkns {
				if tkn == gTkn {
					event.Response <- false
					continue
				}
			}

			res, err := c.HTTPClient.GetTokenPrice(tkn)
			if err != nil {
				c.LogError("Error fetching token price", err.Error())
				event.Response <- false
				continue
			}
			price := fmt.Sprintf("$%.2f", res.Data[tkn].USD)

			guild, ok := c.DiscordClient.GuildMap[event.Guild.ID]
			if !ok {
				c.LogError("guild not found in map")
				event.Response <- false
				continue
			}
			c.DiscordClient.GuildMapMutex.Lock()
			guild.ConfiguredTokens = append(guild.ConfiguredTokens, tkn)
			c.DiscordClient.GuildMap[event.Guild.ID] = guild
			c.DiscordClient.GuildMapMutex.Unlock()

			err = c.Database.UpdateGuild(guild)
			if err != nil {
				c.LogError("Error updating guild", err.Error())
				event.Response <- false
				continue
			}
			event.Response <- true

			c.DiscordClient.ModifyEmbed(guild, price)
			c.DiscordClient.ConfigureGuild(guild, tkn)

		case discord.RemoveCurrency:
			// Save data using c.Database
			//event.Response <- "save confirmation"

		case discord.Autocomplete:
		}
	}
}

// NewController returns a new controller
func NewController() *Controller {
	return &Controller{
		DiscordClient: discord.NewApplication(),
		Database:      database.NewDatabase(),
		HTTPClient:    http.NewClient(),
		Logger:        log.New(log.Writer(), "Controller", log.LstdFlags),
		CoinCache:     make(map[string]http.Coin),
	}
}

func (c *Controller) Initialize() {
	c.DiscordClient.AddHandlers()
	err := c.Database.CreateTables()
	if err != nil {
		c.LogError("Error creating tables", err.Error())
	}
	c.ListenToEvents()
	coins, err := c.HTTPClient.GetCoins()
	if err != nil {
		c.LogError("Error fetching coins", err.Error())
	}
	for _, coin := range coins {
		c.CoinCache[coin.Symbol] = coin
	}

	guilds, err := c.Database.GetConfig()
	if err != nil {
		c.LogError("Error fetching guilds", err.Error())
	}
	for _, guild := range guilds {
		c.DiscordClient.GuildMap[guild.ID] = guild
	}

}

func (c *Controller) LogRequest(message ...string) {
	c.Logger.Printf("[I] %v", message)
}

func (c *Controller) LogError(message ...string) {
	c.Logger.Printf("[E] %v", message)
}
