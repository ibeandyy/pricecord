package discord

type EventType int

const (
	AddCurrency EventType = iota
	RemoveCurrency
)

type Event struct {
	Type     EventType
	Data     interface{}
	Response chan interface{}
}
