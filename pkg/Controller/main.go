package controller

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"log"
	"pricecord/pkg/Database"
	"pricecord/pkg/Discord"
	utils "pricecord/pkg/Discord/Utils"
	"pricecord/pkg/HTTP"
	"strings"
)

// Controller is the main entrypoint for the application
type Controller struct {
	DiscordClient *discord.Application
	Database      *database.Database
	HTTPClient    *http.Client
	Logger        *log.Logger
	TokenCache    map[string]http.Token
}

// NewController returns a new controller
func NewController() *Controller {
	return &Controller{
		DiscordClient: discord.NewApplication(),
		Database:      database.NewDatabase(),
		HTTPClient:    http.NewClient(),
		Logger:        log.New(log.Writer(), "Controller", log.LstdFlags),
		TokenCache:    make(map[string]http.Token),
	}
}

func (c *Controller) Initialize() {
	err := c.DiscordClient.AddHandlers()
	if err != nil {
		c.LogError("Error adding handlers", err.Error())
		return
	}
	err = c.Database.CreateTables()
	if err != nil {
		c.LogError("Error creating tables", err.Error())
		return
	}
	c.ListenToEvents()
	tokens, err := c.HTTPClient.GetTokens()
	if err != nil {
		c.LogError("Error fetching tokens", err.Error())
		return
	}
	for _, token := range tokens {
		c.TokenCache[token.Symbol] = token
	}

	guilds, err := c.Database.GetConfig()
	if err != nil {
		c.LogError("Error fetching guilds", err.Error())
	}
	for _, guild := range guilds {
		c.DiscordClient.GuildMap[guild.ID] = guild
	}

}

func (c *Controller) ListenToEvents() {
	for event := range c.DiscordClient.Event {
		switch event.Type {
		case discord.TrackToken:
			tkn, gTkns := event.Symbol, event.Guild.ConfiguredTokens
			newTrackToken, ok := c.TokenCache[tkn]

			if !ok {
				c.LogError("Token not found in cache")
				event.Response <- false
				continue
			}
			for _, gTkn := range gTkns {

				if tkn == gTkn.Symbol {
					c.LogError("Token already being tracked")
					event.Response <- false
					continue
				}
			}

			res, err := c.HTTPClient.GetTokenPrice(newTrackToken.ID)
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

			err = c.Database.UpdateGuild(guild)
			if err != nil {
				c.LogError("Error updating guild", err.Error())
				event.Response <- false
				continue
			}
			c.DiscordClient.GuildMapMutex.Lock()
			c.DiscordClient.ConfigureGuild(guild, newTrackToken)
			guild.ConfiguredTokens = append(guild.ConfiguredTokens, newTrackToken)
			c.DiscordClient.GuildMap[event.Guild.ID] = guild
			c.DiscordClient.GuildMapMutex.Unlock()

			err = c.DiscordClient.ModifyField(guild, tkn, price)
			if err != nil {
				c.LogError("Error modifying embed", err.Error())
				event.Response <- false
				continue
			}
			event.Response <- true
		case discord.RemoveToken:
			tkn, gTkns := event.Symbol, event.Guild.ConfiguredTokens
			guild, ok := c.DiscordClient.GuildMap[event.Guild.ID]
			if !ok {
				c.LogError("guild not found in map")
				event.Response <- false
				continue
			}
			for _, gTkn := range gTkns {

				if tkn == gTkn.Symbol {

					c.DiscordClient.GuildMapMutex.Lock()
					c.DiscordClient.ConfigureGuild(guild, gTkn)
					guild.ConfiguredTokens = append(guild.ConfiguredTokens, gTkn)
					c.DiscordClient.GuildMap[event.Guild.ID] = guild
					c.DiscordClient.GuildMapMutex.Unlock()
				}
			}

			err := c.Database.UpdateGuild(guild)
			if err != nil {
				c.LogError("Error updating guild", err.Error())
				event.Response <- false
				continue
			}

			err = c.DiscordClient.ModifyField(guild, tkn, "")
			if err != nil {
				c.LogError("Error modifying embed", err.Error())
				event.Response <- false
				continue
			}
			event.Response <- true

		case discord.Autocomplete:
			//TODO: VERIFY GOROUTINE NECESSITY
			go c.routeAutoComplete(event)

		}
	}
}

func (c *Controller) HandleACAddCurr(e discord.Event) {
	//TODO: IF USER INPUT IS EMPTY, RETURN ALL CONFIGURED TOKENS
	var matches []http.Token
	userInput := e.ACValue
	userInputLower := strings.ToLower(userInput)
	var defaultTokens []http.Token
	for _, token := range c.TokenCache {
		defaultTokens = append(defaultTokens, token)
		if strings.Contains(strings.ToLower(token.ID), userInputLower) ||
			strings.Contains(strings.ToLower(token.Symbol), userInputLower) ||
			strings.Contains(strings.ToLower(token.Name), userInputLower) {
			matches = append(matches, token)
		}
	}
	choiceList := ConvertTokenToChoice(matches)
	e.ACResponse <- DefaultTokenCheck(choiceList, defaultTokens)
	e.Response <- true
}

func (c *Controller) HandleACRemoveCurr(e discord.Event) {
	//TODO: IF USER INPUT IS EMPTY, RETURN ALL CONFIGURED TOKENS
	var matches []http.Token
	userInput := e.ACValue
	userInputLower := strings.ToLower(userInput)
	for _, token := range e.Guild.ConfiguredTokens {
		if strings.Contains(strings.ToLower(token.ID), userInputLower) ||
			strings.Contains(strings.ToLower(token.Symbol), userInputLower) ||
			strings.Contains(strings.ToLower(token.Name), userInputLower) {
			matches = append(matches, token)
		}
	}
	choiceList := ConvertTokenToChoice(matches)
	e.ACResponse <- DefaultTokenCheck(choiceList, e.Guild.ConfiguredTokens)
	e.Response <- true
}

func (c *Controller) HandleACAddOther(e discord.Event) {
	var matches []*discordgo.ApplicationCommandOptionChoice
	userInput := e.ACValue
	userInputLower := strings.ToLower(userInput)
	for _, choice := range utils.DefaultOtherChoices {
		if strings.Contains(strings.ToLower(choice.Name), userInputLower) {
			matches = append(matches, choice)
		}
	}
	e.ACResponse <- matches
	e.Response <- true
}

func (c *Controller) HandleACRemoveOther(e discord.Event) {
	var matches []*discordgo.ApplicationCommandOptionChoice
	userInput := e.ACValue
	userInputLower := strings.ToLower(userInput)
	for _, choice := range e.Guild.ConfiguredOthers {
		if strings.Contains(strings.ToLower(choice.Name), userInputLower) {
			newChoice := &discordgo.ApplicationCommandOptionChoice{
				Name:  choice.Name,
				Value: choice.Value,
			}
			matches = append(matches, newChoice)
		}
	}
	e.ACResponse <- matches
	e.Response <- true
}

func (c *Controller) routeAutoComplete(e discord.Event) {
	switch e.ACType {
	case discord.AddCurr:
		go c.HandleACAddCurr(e)
	case discord.RemoveCurr:
		go c.HandleACRemoveCurr(e)
	case discord.AddOther:
		go c.HandleACAddOther(e)
	case discord.RemoveOther:
		go c.HandleACRemoveOther(e)

	default:
		c.LogError("how did I get here ")
	}
}

func (c *Controller) LogRequest(message ...string) {
	c.Logger.Printf("[I] %v", message)
}

func (c *Controller) LogError(message ...string) {
	c.Logger.Printf("[E] %v", message)
}
