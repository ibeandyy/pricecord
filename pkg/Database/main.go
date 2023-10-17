package database

import (
	"database/sql"
	"encoding/json"
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
	err := d.DB.Ping()
	if err != nil {
		d.LogError("Error pinging database", err.Error())
		return err
	}
	d.LogRequest("Creating tables")
	_, ok := d.Exec(`CREATE TABLE IF NOT EXISTS guilds (
		   ID TEXT PRIMARY KEY,
    ConfiguredTokens TEXT,  -- JSON serialized []http.Token
    ConfiguredOthers TEXT,  -- JSON serialized []OtherStat
    ChannelID TEXT,
    MessageID TEXT,
    LastChecked DATETIME
	)`)
	if ok != nil {
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
		var tokensJSON, othersJSON string
		err := rows.Scan(&g.ID, &tokensJSON, &othersJSON, &g.ChannelID, &g.MessageID, &g.LastChecked)
		if err != nil {
			d.LogError("Error scanning rows ", err.Error())
			return nil, err
		}
		// Deserialize JSON into slices
		json.Unmarshal([]byte(tokensJSON), &g.ConfiguredTokens)
		json.Unmarshal([]byte(othersJSON), &g.ConfiguredOthers)

		gs = append(gs, g)
	}
	return gs, nil
}
func (d *Database) GetGuild(id string) (discord.GuildConfiguration, error) {
	d.LogRequest("Getting guild ", id)
	var g discord.GuildConfiguration
	var tokensJSON, othersJSON string
	err := d.QueryRow(`SELECT * FROM guilds WHERE id = ?`, id).Scan(&g.ID, &tokensJSON, &othersJSON, &g.ChannelID, &g.MessageID, &g.LastChecked)
	if err != nil {
		d.LogError("Error getting guild ", err.Error())
		return discord.GuildConfiguration{}, err
	}
	// Deserialize JSON into slices
	json.Unmarshal([]byte(tokensJSON), &g.ConfiguredTokens)
	json.Unmarshal([]byte(othersJSON), &g.ConfiguredOthers)

	return g, nil
}
func (d *Database) AddGuild(g discord.GuildConfiguration) error {
	d.LogRequest("Adding guild ", g.ID)
	tokensJSON, _ := json.Marshal(g.ConfiguredTokens)
	othersJSON, _ := json.Marshal(g.ConfiguredOthers)
	_, err := d.Exec(`INSERT INTO guilds (ID, ConfiguredTokens, ConfiguredOthers, ChannelID, MessageID, LastChecked) VALUES (?, ?, ?, ?, ?, ?)`, g.ID, tokensJSON, othersJSON, g.ChannelID, g.MessageID, g.LastChecked)
	if err != nil {
		d.LogError("Error adding guild ", err.Error())
		return err
	}
	return nil
}

func (d *Database) UpdateGuild(g discord.GuildConfiguration) error {
	d.LogRequest("Updating guild ", g.ID)
	tokensJSON, _ := json.Marshal(g.ConfiguredTokens)
	othersJSON, _ := json.Marshal(g.ConfiguredOthers)
	_, err := d.Exec(`UPDATE guilds SET ConfiguredTokens = ?, ConfiguredOthers = ?, ChannelID = ?, MessageID = ?, LastChecked = ? WHERE ID = ?`, tokensJSON, othersJSON, g.ChannelID, g.MessageID, g.LastChecked, g.ID)
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
