package discord

import "github.com/bwmarrin/discordgo"

type EventType int

const (
	TrackToken EventType = iota
	RemoveToken
	Autocomplete
)

type AutocompleteType int

const (
	AddCurr AutocompleteType = iota
	RemoveCurr
	AddOther
	RemoveOther
)

type Event struct {
	Type       EventType
	Guild      GuildConfiguration
	Symbol     string
	ACType     AutocompleteType
	ACValue    string
	ACResponse chan []*discordgo.ApplicationCommandOptionChoice
	Response   chan bool
}
