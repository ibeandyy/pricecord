package discord

import (
	"fmt"
	"log"
	"os"

	"github.com/bwmarrin/discordgo"
)

type Application struct {
	Client             *discordgo.Session
	RegisteredCommands []*discordgo.ApplicationCommand
	HandlerMap         map[string]HandlerFunc
	GuildConfiguration map[string]GuildConfiguration
	Logger             *log.Logger
}

type GuildConfiguration struct {
	GuildID          string
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
		Client:             Client,
		GuildConfiguration: make(map[string]GuildConfiguration),
		HandlerMap: map[string]HandlerFunc{
			"add-currency": func(a *Application, s *discordgo.Session, i *discordgo.InteractionCreate) {
				a.AddCurrency(s, i)
			},
			"remove-currency": func(a *Application, s *discordgo.Session, i *discordgo.InteractionCreate) {
				a.RemoveCurrency(s, i)
			},
		},
		Logger: log.New(log.Writer(), "Discord Client", log.LstdFlags),
	}
	app.LogRequest("created new application")
	return app
}

func (a *Application) AddHandlers() {
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

		a.RegisterCommands(i.Guild.ID, RawCommands)
		a.InitGuildConfig(i.Guild.ID)
	})
	//Guild Leave Handler
	a.Client.AddHandler(func(s *discordgo.Session, e *discordgo.GuildDelete) {
		for i := range a.RegisteredCommands {

			err := s.ApplicationCommandDelete(a.Client.State.User.ID, e.Guild.ID, a.RegisteredCommands[i].ID)

			if err != nil {
				fmt.Printf("error deleting command %v", err)
				os.Exit(1)
			}

		}
		a.removeGuild(e.Guild.ID)
	})

}

func (a *Application) RegisterCommands(guildID string, commands []*discordgo.ApplicationCommand) []*discordgo.ApplicationCommand {
	a.LogRequest("registering commands for guild", guildID)
	registeredCommands := make([]*discordgo.ApplicationCommand, len(commands))

	cmd, err := a.Client.ApplicationCommandBulkOverwrite(a.Client.State.User.ID, guildID, commands)
	if err != nil {
		a.LogError(err)
		os.Exit(1)
	}
	registeredCommands = cmd
	return registeredCommands
}

func (a *Application) SendEmbed(guildID, channelID, message string) {
	a.LogRequest("sending embed for guild", guildID, "in channel", channelID, "with message", message)

	msg, err := a.Client.ChannelMessageSendEmbed(channelID, &discordgo.MessageEmbed{
		Title:       "Crypto Bot",
		Description: message,
	})

	if err != nil {
		a.LogError(err)
		os.Exit(1)
	}
	a.ConfigureGuild(guildID, channelID, msg.ID, []string{})
}

func (a *Application) ModifyEmbed(guildID, channelID, messageID string, payload []*discordgo.MessageEmbedField) {
	a.LogRequest("modifying embed for guild ", guildID, "in channel ", channelID)

	msg, err := a.Client.ChannelMessageEditEmbed(channelID, messageID, &discordgo.MessageEmbed{
		Fields: payload,
	})
	if err != nil {
		a.LogError(err)
		os.Exit(1)
	}

	a.ConfigureGuild(guildID, channelID, msg.ID, []string{})
}

func (a *Application) InitGuildConfig(guildID string) {
	a.LogRequest("populating guild", guildID)
	a.GuildConfiguration[guildID] = GuildConfiguration{
		GuildID:          guildID,
		ConfiguredTokens: []string{},
		ChannelID:        "",
		MessageID:        "",
	}

}

func (a *Application) ConfigureGuild(guildID, channelID, messageID string, newTokens []string) {
	a.LogRequest("configuring guild", guildID)
	cfg, ok := a.GuildConfiguration[guildID]
	if !ok {
		a.InitGuildConfig(guildID)
		cfg = a.GuildConfiguration[guildID]
	}
	if cfg.ChannelID != "" {
		cfg.ChannelID = channelID
	}
	if cfg.MessageID != "" {
		cfg.MessageID = messageID
	}
	if len(cfg.ConfiguredTokens) > 0 {
		cfg.ConfiguredTokens = newTokens
	}
}

func (a *Application) removeGuild(guildID string) {
	a.LogRequest("removing guild", guildID)
	delete(a.GuildConfiguration, guildID)
}

func (a *Application) LogError(error error) {
	a.Logger.Printf("[E] %v", error)
}

func (a *Application) LogRequest(method ...string) {
	a.Logger.Printf("[I] %v", method)
}

//SEND EMBED

//MODIFY EMBED

//CURRENCY-IN-FIELD-CHECK

//
