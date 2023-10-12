package database

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

type Database struct {
	*sql.DB
}
type Guild struct {
	ID               string
	ConfiguredTokens []string
	ChannelID        string
	MessageID        string
}

func NewDatabase() *Database {
	db, err := sql.Open("sqlite3", "./data.db")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return &Database{db}
}

func (d *Database) CreateTables() error {
	_, err := d.Exec(`CREATE TABLE IF NOT EXISTS guilds (
		id TEXT PRIMARY KEY,
		configured_tokens TEXT,
		channel_id TEXT,
		message_id TEXT
	)`)
	if err != nil {
		return err
	}
	return nil
}

func (d *Database) GetConfig() ([]Guild, error) {
	var gs []Guild
	rows, err := d.Query(`SELECT * FROM guilds`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var g Guild
		err := rows.Scan(&g.ID, &g.ConfiguredTokens, &g.ChannelID, &g.MessageID)
		if err != nil {
			return nil, err
		}
		gs = append(gs, g)
	}
	return gs, nil
}

func (d *Database) GetGuild(id string) (Guild, error) {
	var g Guild
	err := d.QueryRow(`SELECT * FROM guilds WHERE id = ?`, id).Scan(&g.ID, &g.ConfiguredTokens, &g.ChannelID, &g.MessageID)
	if err != nil {
		return Guild{}, err
	}
	return g, nil
}

func (d *Database) AddGuild(g Guild) error {
	_, err := d.Exec(`INSERT INTO guilds VALUES (?, ?, ?, ?)`, g.ID, g.ConfiguredTokens, g.ChannelID, g.MessageID)
	if err != nil {
		return err
	}
	return nil
}

func (d *Database) UpdateGuild(g Guild) error {
	_, err := d.Exec(`UPDATE guilds SET configured_tokens = ?, channel_id = ?, message_id = ? WHERE id = ?`, g.ConfiguredTokens, g.ChannelID, g.MessageID, g.ID)
	if err != nil {
		return err
	}
	return nil
}

func (d *Database) RemoveGuild(id string) error {
	_, err := d.Exec(`DELETE FROM guilds WHERE id = ?`, id)
	if err != nil {
		return err
	}
	return nil
}
