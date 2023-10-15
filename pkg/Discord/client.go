package discord

import (
	"errors"
	"fmt"
	"log"
	"os"
	"pricecord/pkg/HTTP"
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
	ConfiguredTokens []http.Token
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
			"track-token": func(a *Application, s *discordgo.Session, i *discordgo.InteractionCreate) {
				go a.TrackToken(s, i)
			},
			"remove-token": func(a *Application, s *discordgo.Session, i *discordgo.InteractionCreate) {
				go a.RemoveToken(s, i)
			},
		},
		Logger: log.New(log.Writer(), "Discord Client", log.LstdFlags),
		Event:  make(chan Event, 5), //TODO: Benchmark this
	}
	app.LogRequest("created new application")
	return app
}

func (a *Application) AddHandlers() error {
	a.LogRequest("adding handlers")
	//Slash Command Handler
	a.Client.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if i.Type != discordgo.InteractionApplicationCommand {
			return
		}
		if handler, ok := a.HandlerMap[i.ApplicationCommandData().Name]; ok {
			handler(a, a.Client, i)
			return
		}
		return
	})
	//Guild Join Handler
	a.Client.AddHandler(func(s *discordgo.Session, i *discordgo.GuildCreate) {

		err := a.RegisterCommands(i.Guild.ID)
		if err != nil {
			return
		}
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
	return nil
}

func (a *Application) RegisterCommands(guildID string) error {
	a.LogRequest("registering commands for guild", guildID)

	cmd, err := a.Client.ApplicationCommandBulkOverwrite(a.Client.State.User.ID, guildID, RawCommands)
	if err != nil {
		return err
	}
	a.RegisteredCommands = cmd
	return nil
}

func (a *Application) SendEmbed(gCfg GuildConfiguration) error {
	a.LogRequest("sending embed for guild", gCfg.ID, "in channel", gCfg.MessageID)

	_, err := a.Client.ChannelMessageSendEmbed(gCfg.ChannelID, &discordgo.MessageEmbed{
		Title: "Crypto Bot",
		//TODO:ADD DEFAULT MESSAGE EMBED
		Description: "",
	})

	if err != nil {
		return err
	}
	return nil
}

// ModifyField updates the fields of the embed
// If the token is not found, a new field is added
// If the provided price string is empty, the field is removed
func (a *Application) ModifyField(g GuildConfiguration, tkn, price string) error {
	a.LogRequest("modifying embed for guild ", g.ID, "in channel ", g.ChannelID)

	msg, err := a.Client.ChannelMessage(g.ChannelID, g.MessageID)
	if err != nil {
		return err
	}

	if len(msg.Embeds) == 0 {
		return errors.New("no embeds in the message")
	}

EmbLoop:
	for eIdx := range msg.Embeds {
		emb := msg.Embeds[eIdx] // Get the embed using index

		// Find the field with the token name
		// If none found, add a new field to the first embed with <= 24 fields
		// If no embeds have <= 24 fields, add a new embed
		for i, field := range emb.Fields {
			if field.Name == tkn {
				if price == "" {
					emb.Fields = append(emb.Fields[:i], emb.Fields[i+1:]...)
					break EmbLoop
				} else {
					emb.Fields[i].Value = price
					break EmbLoop
				}

			}
		}
		if len(emb.Fields) <= 24 {
			emb.Fields = append(emb.Fields, &discordgo.MessageEmbedField{
				Name:   tkn,
				Value:  price,
				Inline: true,
			})
			break EmbLoop
		}
	}

	_, err = a.Client.ChannelMessageEditComplex(&discordgo.MessageEdit{
		ID:      msg.ID,
		Channel: msg.ChannelID,
		Embeds:  msg.Embeds,
	})
	return err
}

func (a *Application) InitGuildConfig(guildID string) {
	a.LogRequest("populating guild", guildID)
	a.GuildMap[guildID] = GuildConfiguration{
		ID:               guildID,
		ConfiguredTokens: []http.Token{},
		ChannelID:        "",
		MessageID:        "",
	}

}

func (a *Application) ConfigureGuild(g GuildConfiguration, newTokens ...http.Token) {
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
