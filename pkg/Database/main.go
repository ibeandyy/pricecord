package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
	"pricecord/pkg/Discord"
)

type Database struct {
	*sql.DB
	Logger *log.Logger
}

func NewDatabase() *Database {
	db, err := sql.Open("sqlite3", "./data.db")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return &Database{db,
		log.New(log.Writer(), "Database", log.LstdFlags)}
}

func (d *Database) CreateTables() error {
	d.LogRequest("Creating tables")
	_, err := d.Exec(`CREATE TABLE IF NOT EXISTS guilds (
		id TEXT PRIMARY KEY,
		configured_tokens TEXT,
		channel_id TEXT,
		message_id TEXT
	)`)
	if err != nil {
		d.LogError("Error creating tables", err.Error())
		return err
	}
	return nil
}

func (d *Database) GetConfig() ([]discord.GuildConfiguration, error) {
	d.LogRequest("Getting guilds")
	var gs []discord.GuildConfiguration
	rows, err := d.Query(`SELECT * FROM guilds`)
	if err != nil {
		d.LogError("Error getting guilds", err.Error())
		return nil, err
	}
	defer func() {
		err := rows.Close()
		if err != nil {
			d.LogRequest("Error closing rows", err.Error())
		}
	}()
	for rows.Next() {
		var g discord.GuildConfiguration
		err := rows.Scan(&g.ID, &g.ConfiguredTokens, &g.ChannelID, &g.MessageID)
		if err != nil {
			d.LogError("Error scanning rows ", err.Error())
			return nil, err
		}
		gs = append(gs, g)
	}
	return gs, nil
}

func (d *Database) GetGuild(id string) (discord.GuildConfiguration, error) {
	d.LogRequest("Getting guild ", id)
	var g discord.GuildConfiguration
	err := d.QueryRow(`SELECT * FROM guilds WHERE id = ?`, id).Scan(&g.ID, &g.ConfiguredTokens, &g.ChannelID, &g.MessageID)
	if err != nil {
		d.LogError("Error getting guild ", err.Error())
		return discord.GuildConfiguration{}, err
	}
	return g, nil
}

func (d *Database) AddGuild(g discord.GuildConfiguration) error {
	d.LogRequest("Adding guild ", g.ID)
	_, err := d.Exec(`INSERT INTO guilds VALUES (?, ?, ?, ?)`, g.ID, g.ConfiguredTokens, g.ChannelID, g.MessageID)
	if err != nil {
		d.LogError("Error adding guild ", err.Error())
		return err
	}
	return nil
}

func (d *Database) UpdateGuild(g discord.GuildConfiguration) error {
	d.LogRequest("Updating guild ", g.ID)
	_, err := d.Exec(`UPDATE guilds SET configured_tokens = ?, channel_id = ?, message_id = ? WHERE id = ?`, g.ConfiguredTokens, g.ChannelID, g.MessageID, g.ID)
	if err != nil {
		d.LogError("Error updating guild", err.Error())
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

func (d *Database) LogRequest(message ...string) {
	log.Printf("[I] %v", message)
}

func (d *Database) LogError(error ...string) {
	log.Printf("[E] %v", error)
}
