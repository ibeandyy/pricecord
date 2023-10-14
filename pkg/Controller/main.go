package controller

import (
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
			// Fetch data using c.HTTPClient
			//event.Response <- "fetched data"
		case discord.RemoveCurrency:
			// Save data using c.Database
			//event.Response <- "save confirmation"
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

	//Load config from database if exists
	//
}

func (c *Controller) LogRequest(message ...string) {
	c.Logger.Printf("[I] %v", message)
}

func (c *Controller) LogError(message ...string) {
	c.Logger.Printf("[E] %v", message)
}
