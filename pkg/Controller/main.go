package controller

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"log"
	"pricecord/pkg/Database"
	"pricecord/pkg/Discord"
	"pricecord/pkg/Discord/Utils"
	"pricecord/pkg/HTTP"
	"strings"
)

// Controller is the main entrypoint for the application
type Controller struct {
	DiscordClient *discord.Application
	Database      *database.Database
	HTTPClient    *http.Client
	Logger        *log.Logger
	TokenCache    map[string]*http.Token
	DefaultTokens []*http.Token
}

// NewController returns a new controller
func NewController(token string) *Controller {
	return &Controller{
		DiscordClient: discord.NewApplication(token),
		Database:      database.NewDatabase(),
		HTTPClient:    http.NewClient(),
		Logger:        log.New(log.Writer(), "Controller ", log.LstdFlags),
		TokenCache:    make(map[string]*http.Token),
		DefaultTokens: []*http.Token{},
	}
}

func (c *Controller) Initialize() {

	err := c.Database.CreateTables()
	if err != nil {
		c.LogError("Error creating tables", err.Error())
		return
	}
	tokens, err := c.HTTPClient.GetTokens()
	if err != nil {
		c.LogError("Error fetching tokens", err.Error())
		return
	}
	c.DefaultTokens = tokens

	for _, token := range tokens {
		c.TokenCache[strings.ToLower(token.Name)] = token
	}

	guilds, err := c.Database.GetConfig()
	if err != nil {
		c.LogError("Error fetching guilds", err.Error())
	}
	for _, guild := range guilds {
		c.DiscordClient.GuildMap[guild.ID] = guild
	}
	err = c.DiscordClient.AddHandlers()
	if err != nil {
		c.LogError("Error adding handlers", err.Error())
		return
	}

}

func (c *Controller) ListenToEvents() {
	for event := range c.DiscordClient.Event {
		c.LogRequest("Responding to event")
		switch event.Type {
		case discord.TrackToken:
			tkn, gTkns := event.Name, event.Guild.ConfiguredTokens
			tkn = strings.ToLower(tkn)
			newTrackToken, ok := c.TokenCache[tkn]

			if !ok {
				c.LogError("Token not found in cache")
				event.Response <- false
				continue
			}
			for _, gTkn := range gTkns {

				if tkn == gTkn.Name {
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

			if _, ok := res[tkn]; !ok {
				c.LogError("error getting token price")
				event.Response <- false
				continue
			}

			price := fmt.Sprintf("$%.2f", res[tkn].USD)
			guild, ok := c.DiscordClient.GuildMap[event.Guild.ID]
			if !ok {
				c.LogError("guild not found in map")
				event.Response <- false
				continue
			}

			c.DiscordClient.GuildMapMutex.Lock()
			c.DiscordClient.ConfigureGuild(guild, []*http.Token{newTrackToken}, []*discord.OtherStat{}, false)
			c.DiscordClient.GuildMapMutex.Unlock()
			err = c.DiscordClient.ModifyField(guild, tkn, price)
			if err != nil {
				c.LogError("Error modifying embed", err.Error())
				event.Response <- false
				continue
			}
			c.LogRequest("DEBUG LOG: ", c.DiscordClient.GuildMap[guild.ID].ConfiguredTokens[0].ID)
			c.LogRequest("DEBUG LOG: ", guild.ConfiguredTokens[0].Name)
			err = c.Database.UpdateGuild(guild)
			if err != nil {
				c.LogError("Error updating guild", err.Error())
				event.Response <- false
				continue
			}

			event.Response <- true
		case discord.RemoveToken:
			tkn, gTkns := event.Name, event.Guild.ConfiguredTokens
			guild, ok := c.DiscordClient.GuildMap[event.Guild.ID]
			if !ok {
				c.LogError("guild not found in map")
				event.Response <- false
				continue
			}
			for _, gTkn := range gTkns {

				if tkn == gTkn.Name {
					c.DiscordClient.GuildMapMutex.Lock()
					c.DiscordClient.ConfigureGuild(guild, []*http.Token{gTkn}, make([]*discord.OtherStat, 0), true)
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
			c.routeAutoComplete(event)
		case discord.TrackOther:
			newStat := event.Stat
			guild, ok := c.DiscordClient.GuildMap[event.Guild.ID]
			if !ok {
				c.LogError("guild not found in map")
				event.Response <- false
				continue
			}

			for _, stat := range guild.ConfiguredOthers {

				if stat.Name == newStat {
					c.DiscordClient.GuildMapMutex.Lock()
					c.DiscordClient.ConfigureGuild(guild, make([]*http.Token, 0), []*discord.OtherStat{stat}, false)
					c.DiscordClient.GuildMapMutex.Unlock()
				}
			}
			fieldName, fieldValue := c.HTTPClient.GetDefiOther(newStat)
			if fieldName == "" {
				c.LogError("Stat not found in defi other")
				event.Response <- false
				continue
			}
			err := c.Database.UpdateGuild(guild)
			if err != nil {
				c.LogError("Error updating guild", err.Error())
				event.Response <- false
				continue
			}

			err = c.DiscordClient.ModifyField(guild, fieldName, fieldValue)
			if err != nil {
				c.LogError("Error modifying embed", err.Error())
				event.Response <- false
				continue
			}
			event.Response <- true

		case discord.RemoveOther:
			c.LogRequest("Received RemoveOther Event")
			newStat := event.Stat
			guild, ok := c.DiscordClient.GuildMap[event.Guild.ID]
			if !ok {
				c.LogError("guild not found in map")
				event.Response <- false
				continue
			}

			for _, stat := range guild.ConfiguredOthers {

				if stat.Name == newStat {
					c.DiscordClient.GuildMapMutex.Lock()
					c.DiscordClient.ConfigureGuild(guild, make([]*http.Token, 0), []*discord.OtherStat{stat}, true)
					c.DiscordClient.GuildMapMutex.Unlock()
				}
			}

			err := c.Database.UpdateGuild(guild)
			if err != nil {
				c.LogError("Error updating guild", err.Error())
				event.Response <- false
				continue
			}

			err = c.DiscordClient.ModifyField(guild, newStat, "")
			if err != nil {
				c.LogError("Error modifying embed", err.Error())
				event.Response <- false
				continue
			}
			event.Response <- true
		case discord.InitGuild:
			c.LogRequest("Received InitGuild Event")
			err := c.Database.AddGuild(event.Guild)
			if err != nil {
				c.LogError("Error adding guild", err.Error())
				event.Response <- false
				continue
			}
			event.Response <- true

		case discord.DeleteGuild:
			c.LogRequest("Received DeleteGuild Event")
			err := c.Database.RemoveGuild(event.Guild.ID)
			if err != nil {
				c.LogError("Error deleting guild", err.Error())
				event.Response <- false
				continue
			}
			event.Response <- true

		}
	}
	c.LogRequest("Event channel closed")
}

func (c *Controller) HandleACAddCurr(e discord.Event) {
	var matches []*http.Token
	userInput := e.ACValue
	fmt.Println("USER INPUT IS ", userInput)
	if userInput == "" {
		e.ACResponse <- c.ConvertTokenToChoice(c.DefaultTokens)

		return
	}
	userInputLower := strings.ToLower(userInput)
	var defaultTokens []*http.Token
	//for _, token := range c.TokenCache {
	//	defaultTokens = append(defaultTokens, token)
	//	if strings.Contains(strings.ToLower(token.ID), userInputLower) ||
	//		strings.Contains(strings.ToLower(token.Name), userInputLower) ||
	//		strings.Contains(strings.ToLower(token.Name), userInputLower) {
	//		matches = append(matches, token)
	//	}
	//}
	for _, token := range c.TokenCache {
		defaultTokens = append(defaultTokens, token)
		if strings.ToLower(token.Name) == userInputLower {

			matches = append(matches, token)
		}
	}
	choiceList := c.ConvertTokenToChoice(matches)
	e.ACResponse <- c.DefaultTokenCheck(choiceList, defaultTokens)
}

func (c *Controller) HandleACRemoveCurr(e discord.Event) {
	var matches []*http.Token
	userInput := e.ACValue
	if userInput == "" {
		e.ACResponse <- c.ConvertTokenToChoice(e.Guild.ConfiguredTokens)
		return
	}
	userInputLower := strings.ToLower(userInput)
	for _, token := range e.Guild.ConfiguredTokens {
		if strings.Contains(strings.ToLower(token.Name), userInputLower) {
			matches = append(matches, token)
		}
	}
	choiceList := c.ConvertTokenToChoice(matches)
	e.ACResponse <- c.DefaultTokenCheck(choiceList, e.Guild.ConfiguredTokens)
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
}

func (c *Controller) Close() {
	err := c.DiscordClient.Close()
	if err != nil {
		c.LogError("error closing client", err.Error())
	}
}

func (c *Controller) routeAutoComplete(e discord.Event) {
	switch e.ACType {
	case discord.AddCurr:
		c.HandleACAddCurr(e)
	case discord.RemCurr:
		c.HandleACRemoveCurr(e)
	case discord.AddOther:
		c.HandleACAddOther(e)
	case discord.RemOther:
		c.HandleACRemoveOther(e)

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
