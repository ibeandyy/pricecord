package discord

import (
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/bwmarrin/discordgo"
)

type Application struct {
	Client             *discordgo.Session
	RegisteredCommands []*discordgo.ApplicationCommand
	HandlerMap         map[string]HandlerFunc
	GuildMapMutex      sync.RWMutex
	GuildMap           map[string]GuildConfiguration
	Logger             *log.Logger
	Event              chan Event
}

type GuildConfiguration struct {
	ID               string
	ConfiguredTokens []string
	ChannelID        string
	MessageID        string
}
type HandlerFunc func(a *Application, s *discordgo.Session, i *discordgo.InteractionCreate)

func NewApplication() *Application {
	tkn := os.Getenv("TOKEN")

	if tkn == "" {
		fmt.Printf("unable to fetch Discord token %v", tkn)
	}

	Client, err := discordgo.New(tkn)

	if err != nil {
		fmt.Printf("error connecting client %v", err)
	}

	app := &Application{
		Client:   Client,
		GuildMap: make(map[string]GuildConfiguration),
		HandlerMap: map[string]HandlerFunc{
			"add-currency": func(a *Application, s *discordgo.Session, i *discordgo.InteractionCreate) {
				go a.AddCurrency(s, i)
			},
			"remove-currency": func(a *Application, s *discordgo.Session, i *discordgo.InteractionCreate) {
				go a.RemoveCurrency(s, i)
			},
		},
		Logger: log.New(log.Writer(), "Discord Client", log.LstdFlags),
		Event:  make(chan Event, 5), //TODO: Benchmark this
	}
	app.LogRequest("created new application")
	return app
}

func (a *Application) AddHandlers() {
	a.LogRequest("adding handlers")
	//Slash Command Handler
	a.Client.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if discordgo.InteractionApplicationCommand != i.Type {
			return
		}
		if handler, ok := a.HandlerMap[i.ApplicationCommandData().Name]; ok {
			handler(a, a.Client, i)
		}
	})
	//Guild Join Handler
	a.Client.AddHandler(func(s *discordgo.Session, i *discordgo.GuildCreate) {

		a.RegisterCommands(i.Guild.ID)
		a.InitGuildConfig(i.Guild.ID)
	})
	//Guild Leave Handler
	a.Client.AddHandler(func(s *discordgo.Session, e *discordgo.GuildDelete) {
		for i := range a.RegisteredCommands {

			err := s.ApplicationCommandDelete(a.Client.State.User.ID, e.Guild.ID, a.RegisteredCommands[i].ID)

			if err != nil {
				a.LogError("error deleting command ", err.Error())
				os.Exit(1)
			}

		}
		a.removeGuild(e.Guild.ID)
	})

}

func (a *Application) RegisterCommands(guildID string) []*discordgo.ApplicationCommand {
	a.LogRequest("registering commands for guild", guildID)
	registeredCommands := make([]*discordgo.ApplicationCommand, len(RawCommands))

	cmd, err := a.Client.ApplicationCommandBulkOverwrite(a.Client.State.User.ID, guildID, RawCommands)
	if err != nil {
		a.LogError("error registering commands ", err.Error())
		os.Exit(1)
	}
	registeredCommands = cmd
	return registeredCommands
}

func (a *Application) SendEmbed(gCfg GuildConfiguration) string {
	a.LogRequest("sending embed for guild", gCfg.ID, "in channel", gCfg.MessageID)

	msg, err := a.Client.ChannelMessageSendEmbed(gCfg.ChannelID, &discordgo.MessageEmbed{
		Title: "Crypto Bot",
		//TODO:ADD DEFAULT MESSAGE EMBED
		Description: "",
	})

	if err != nil {
		a.LogError("error sending embed ", err.Error())
		os.Exit(1)
	}
	return msg.ID
}

func (a *Application) ModifyEmbed(g GuildConfiguration, price string) {
	a.LogRequest("modifying embed for guild ", g.ID, "in channel ", g.ChannelID)

	_, err := a.Client.ChannelMessageEditEmbed(g.ChannelID, g.MessageID, &discordgo.MessageEmbed{
		//TODO:FIX
		Fields: nil,
	})
	if err != nil {
		a.LogError("error editing embed ", err.Error())
		os.Exit(1)
	}

}

func (a *Application) InitGuildConfig(guildID string) {
	a.LogRequest("populating guild", guildID)
	a.GuildMap[guildID] = GuildConfiguration{
		ID:               guildID,
		ConfiguredTokens: []string{},
		ChannelID:        "",
		MessageID:        "",
	}

}

func (a *Application) ConfigureGuild(g GuildConfiguration, newTokens ...string) {
	a.LogRequest("configuring guild", g.ID)
	cfg, ok := a.GuildMap[g.ID]
	if !ok {
		a.InitGuildConfig(g.ID)
		cfg = a.GuildMap[g.ID]
	}
	if cfg.ChannelID != "" {
		cfg.ChannelID = g.ChannelID
	}
	if cfg.MessageID != "" {
		cfg.MessageID = g.MessageID
	}
	if len(cfg.ConfiguredTokens) > 0 {
		cfg.ConfiguredTokens = append(cfg.ConfiguredTokens, newTokens...)
	}
}

func (a *Application) removeGuild(guildID string) {
	a.LogRequest("removing guild", guildID)
	delete(a.GuildMap, guildID)
}

func (a *Application) LogError(error ...string) {
	a.Logger.Printf("[E] %v", error)
}

func (a *Application) LogRequest(method ...string) {
	a.Logger.Printf("[I] %v", method)
}

//SEND EMBED

//MODIFY EMBED

//CURRENCY-IN-FIELD-CHECK

//
