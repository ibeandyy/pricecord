package discord

import (
	"errors"
	"fmt"
	"log"
	"os"
	"pricecord/pkg/HTTP"
	"sync"
	"time"

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

type OtherStat struct {
	Name  string
	Value string
}

type GuildConfiguration struct {
	ID               string
	ConfiguredTokens []http.Token
	ConfiguredOthers []OtherStat
	ChannelID        string
	MessageID        string
	LastChecked      time.Time
}
type HandlerFunc func(a *Application, s *discordgo.Session, i *discordgo.InteractionCreate)

var DefaultEmbed = &discordgo.MessageEmbed{
	Title:       "Pricecord",
	Color:       0x82ff80,
	Timestamp:   time.Now().Format(time.RFC3339),
	Description: "Use the /track-* commands to add something for me to track!",
}

func NewApplication(token string) *Application {

	Client, err := discordgo.New("Bot " + token)

	if err != nil {
		fmt.Printf("error connecting client %v", err)
		os.Exit(1)
	}
	err = Client.Open()
	if err != nil {
		fmt.Printf("error opening connection %v", err)
		os.Exit(1)
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
			"track-other": func(a *Application, s *discordgo.Session, i *discordgo.InteractionCreate) {
				go a.TrackOther(s, i)
			},
			"remove-other": func(a *Application, s *discordgo.Session, i *discordgo.InteractionCreate) {
				go a.RemoveOther(s, i)
			},
		},
		Logger: log.New(log.Writer(), "Discord Client ", log.LstdFlags),
		Event:  make(chan Event, 5),
	}
	app.LogRequest("created new application")
	return app
}

func (a *Application) AddHandlers() error {
	a.LogRequest("adding handlers")
	//Slash Command Handler
	a.Client.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		a.LogRequest("Received interaction", i.Type.String())
		//if i.Type != discordgo.InteractionApplicationCommand {
		//	return
		//}
		if handler, ok := a.HandlerMap[i.ApplicationCommandData().Name]; ok {
			handler(a, a.Client, i)

		}

	})
	//Guild Join Handler
	a.Client.AddHandler(func(s *discordgo.Session, i *discordgo.GuildCreate) {
		a.LogRequest("Received GuildJoin Event")

		err := a.RegisterCommands(i.Guild.ID)
		if err != nil {
			a.LogError("error registering commands", err.Error())
			return
		}
		err = a.InitGuildConfig(i.Guild.ID)
		if err != nil {
			a.LogError("error initializing guild", err.Error())
			return
		}
	})
	//Guild Leave Handler
	a.Client.AddHandler(func(s *discordgo.Session, e *discordgo.GuildDelete) {
		a.LogRequest("Received GuildDelete Event")

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

	for _, cmd := range RawCommands {
		_, err := a.Client.ApplicationCommandCreate(a.Client.State.User.ID, guildID, cmd)
		if err != nil {
			return err
		}
	}

	a.RegisteredCommands = RawCommands
	return nil
}

func (a *Application) SendEmbed(gCfg GuildConfiguration) error {
	a.LogRequest("sending embed for guild", gCfg.ID, "in channel", gCfg.MessageID)

	_, err := a.Client.ChannelMessageSendEmbed(gCfg.ChannelID, DefaultEmbed)

	if err != nil {
		a.LogError("error sending embed", err.Error())
		return err
	}
	return nil
}

// ModifyField updates the fields of the embed
// If the token is not found, a new field is added
// If the provided price string is empty, the field is removed
func (a *Application) ModifyField(g GuildConfiguration, name, value string) error {
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
			if field.Name == name {
				if value == "" {
					emb.Fields = append(emb.Fields[:i], emb.Fields[i+1:]...)
					break EmbLoop
				} else {
					emb.Fields[i].Value = value
					break EmbLoop
				}

			}
		}
		if len(emb.Fields) <= 24 {
			emb.Fields = append(emb.Fields, &discordgo.MessageEmbedField{
				Name:   name,
				Value:  value,
				Inline: true,
			})
			break EmbLoop
		}
	}

	_, err = a.Client.ChannelMessageEditComplex(&discordgo.MessageEdit{
		ID:      msg.ID,
		Channel: msg.ChannelID,
		Embeds:  msg.Embeds,
		Content: nil,
	})
	return err
}

func (a *Application) InitGuildConfig(guildID string) error {
	a.LogRequest("populating guild", guildID)
	if _, ok := a.GuildMap[guildID]; ok {
		return errors.New("guild already initialized")
	}
	ch, err := a.Client.GuildChannelCreate(guildID, "Pricecord", discordgo.ChannelTypeGuildText)
	if err != nil {
		return err
	}
	msg, err := a.Client.ChannelMessageSendEmbed(ch.ID, DefaultEmbed)
	if err != nil {
		return err
	}

	a.GuildMap[guildID] = GuildConfiguration{
		ID:               guildID,
		ConfiguredTokens: []http.Token{},
		ConfiguredOthers: []OtherStat{},
		ChannelID:        ch.ID,
		MessageID:        msg.ID,
		LastChecked:      time.Now(),
	}
	event := Event{
		Type:     InitGuild,
		Guild:    a.GuildMap[guildID],
		Response: make(chan bool),
	}
	a.Event <- event
	ok := <-event.Response
	if !ok {
		a.LogError("error saving guild to DB")
	}
	return err

}

func (a *Application) ConfigureGuild(g GuildConfiguration, newTokens []http.Token, newOther []OtherStat, delete bool) {
	a.LogRequest("configuring guild", g.ID)
	cfg, ok := a.GuildMap[g.ID]
	if !ok {
		err := a.InitGuildConfig(g.ID)
		if err != nil {
			a.LogError("error initializing guild", err.Error())
			return
		}
		cfg = a.GuildMap[g.ID]
	}
	if cfg.ChannelID != "" {
		cfg.ChannelID = g.ChannelID
	}
	if cfg.MessageID != "" {
		cfg.MessageID = g.MessageID

	}
	if len(cfg.ConfiguredTokens) > 0 {
		if delete {
			for i, tkn := range cfg.ConfiguredTokens {
				for _, newTkn := range newTokens {
					if tkn.ID == newTkn.ID {
						cfg.ConfiguredTokens = append(cfg.ConfiguredTokens[:i], cfg.ConfiguredTokens[i+1:]...)
					}
				}
			}
		} else {
			cfg.ConfiguredTokens = append(cfg.ConfiguredTokens, newTokens...)
		}
	}
	if len(cfg.ConfiguredOthers) > 0 {
		if delete {
			for i, stat := range cfg.ConfiguredOthers {
				for _, newStat := range newOther {
					if stat.Name == newStat.Name {
						cfg.ConfiguredOthers = append(cfg.ConfiguredOthers[:i], cfg.ConfiguredOthers[i+1:]...)
					}
				}
			}
		} else {
			cfg.ConfiguredOthers = append(cfg.ConfiguredOthers, newOther...)
		}
	}
	a.GuildMap[g.ID] = cfg

}

func (a *Application) Close() error {
	a.LogRequest("closing application")
	err := a.Client.Close()
	return err
}

func (a *Application) removeGuild(guildID string) {
	a.LogRequest("removing guild", guildID)
	event := Event{
		Type:     DeleteGuild,
		Guild:    a.GuildMap[guildID],
		Response: make(chan bool),
	}
	ok := <-event.Response
	if ok {
		delete(a.GuildMap, guildID)
	} else {
		a.LogError("error deleting guild from DB")
	}
}

func (a *Application) LogError(error ...string) {
	a.Logger.Printf("[E] %v", error)
}

func (a *Application) LogRequest(method ...string) {
	a.Logger.Printf("[I] %v", method)
}
