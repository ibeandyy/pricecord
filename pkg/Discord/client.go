package discord

import (
	"fmt"
	"os"
)

import (
	"github.com/bwmarrin/discordgo"
)

type Application struct {
	Client             *discordgo.Session
	RegisteredCommands []*discordgo.ApplicationCommand
	HandlerMap         map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate)
}

func NewApplication() *Application {
	tkn := os.Getenv("TOKEN")

	if tkn == "" {
		fmt.Printf("unable to fetch Discord token %v", tkn)
	}

	Client, err := discordgo.New(tkn)

	if err != nil {
		fmt.Printf("error connecting client %v", err)
	}

	a := &Application{
		Client,
		[]*discordgo.ApplicationCommand{},

		HandlerMap,
	}
	return a
}

func (dc *Application) AddHandlers() {
	//Slash Command Handler
	dc.Client.AddHandler(func(s *discordgo.Session, i *discordgo.ApplicationCommand) {

	})
	//Guild Join Handler
	dc.Client.AddHandler(func(s *discordgo.Session, i *discordgo.GuildCreate) {

		dc.RegisterCommands(i.Guild.ID, RawCommands)
	})
	//Guild Leave Handler
	dc.Client.AddHandler(func(s *discordgo.Session, e *discordgo.GuildDelete) {
		for i := range dc.RegisteredCommands {
			err := s.ApplicationCommandDelete(dc.Client.State.User.ID, e.Guild.ID, dc.RegisteredCommands[i].ID)

			if err != nil {
				fmt.Printf("error deleting command %v", err)
				os.Exit(1)
			}
		}
	})

}

func (dc *Application) RegisterCommands(guildID string, commands []*discordgo.ApplicationCommand) []*discordgo.ApplicationCommand {
	registeredCommands := make([]*discordgo.ApplicationCommand, len(commands))
	for i := range commands {
		cmd, err := dc.Client.ApplicationCommandCreate(dc.Client.State.User.ID, guildID, commands[i])
		if err != nil {
			fmt.Printf("error registering command %v", err)
			os.Exit(1)
		}
		registeredCommands[i] = cmd
	}
	return registeredCommands
}

//HANDLE COMMANDS

//SEND EMBED

//MODIFY EMBED

//CURRENCY-IN-FIELD-CHECK

//
