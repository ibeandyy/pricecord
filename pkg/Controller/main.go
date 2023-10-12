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
}

// NewController returns a new controller
func NewController() *Controller {
	return &Controller{
		DiscordClient: discord.NewApplication(),
		Database:      database.NewDatabase(),
		HTTPClient:    http.NewClient(),
		Logger:        log.New(log.Writer(), "Controller", log.LstdFlags),
	}
}

func (c *Controller) Initialize() {
	c.DiscordClient.AddHandlers()
	err := c.Database.CreateTables()
	if err != nil {

	}

}

func (c *Controller) LogRequest(message ...string) {
	c.Logger.Printf("[I] %v", message)
}

func (c *Controller) LogError(message ...string) {
	c.Logger.Printf("[E] %v", message)
}
