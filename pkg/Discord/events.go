package discord

type EventType int

const (
	AddCurrency EventType = iota
	RemoveCurrency
	Autocomplete
)

type Event struct {
	Type     EventType
	Guild    GuildConfiguration
	Symbol   string
	Response chan bool
}
