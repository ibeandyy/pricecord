package discord

import "github.com/bwmarrin/discordgo"

type EventType int

const (
	TrackToken EventType = iota
	RemoveToken
	Autocomplete
	TrackOther
	RemoveOther
)

type AutocompleteType int

const (
	AddCurr AutocompleteType = iota
	RemCurr
	AddOther
	RemOther
)

type Event struct {
	Type       EventType
	Guild      GuildConfiguration
	Symbol     string
	Stat       string
	ACType     AutocompleteType
	ACValue    string
	ACResponse chan []*discordgo.ApplicationCommandOptionChoice
	Response   chan bool
}
