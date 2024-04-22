package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	conn *sql.DB
	lock *sync.Mutex
}

type Config struct {
	UpdatedTime time.Time `json:"updatedTime"`
	DBName      string    `json:"dbName"`
}

const ConfigFile = "dbconfig.json"

var db = DB{
	lock: &sync.Mutex{},
}

var config = Config{}

func readConfigFile() (Config, error) {
	configFile, err := os.Open(ConfigFile)
	if err != nil {
		return Config{}, err
	}

	json.NewDecoder(configFile).Decode(&config)
	configFile.Close()

	return config, nil
}

func writeConfigFile(config *Config) error {
	jsonText, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	err = os.WriteFile(ConfigFile, jsonText, 0755)
	if err != nil {
		return err
	}

	return nil
}

func createDB() error {
	db.lock.Lock()
	defer db.lock.Unlock()

	// Remove old db file
	if _, err := os.Stat(config.DBName); err == nil {
		err = os.Remove(config.DBName)
		if err != nil {
			return err
		}
	}

	// Initialize db connection
	connection, err := sql.Open("sqlite3", config.DBName)
	if err != nil {
		return err
	}
	db.conn = connection

	// Create new table
	_, err = db.conn.Exec(
		`CREATE TABLE test (
			name text PRIMARY KEY
		)`,
	)
	if err != nil {
		return err
	}

	// Insert database name into the first row
	stmt, err := db.conn.Prepare(`INSERT INTO test (name) VALUES (?)`)
	if err != nil {
		return err
	}
	stmt.Exec(config.DBName)

	return nil
}

func InitDB() error {
	configFile, err := readConfigFile()
	if err != nil {
		return err
	}
	config = configFile

	// Database connection
	fmt.Println("Creating new database:", config.DBName)

	err = createDB()
	if err != nil {
		return err
	}

	// Add timestamp
	if config.UpdatedTime.IsZero() {
		config.UpdatedTime = time.Now()
		writeConfigFile(&config)
	}

	return nil
}

func GetName() string {
	db.lock.Lock()
	defer db.lock.Unlock()

	row := db.conn.QueryRow("SELECT name FROM test LIMIT 1")
	if row.Err() != nil {
		log.Println(row.Err().Error())
		return ""
	}
	var title string
	row.Scan(&title)

	return title
}

func SwapDB() {
	now := time.Now()
	newConfig := Config{
		DBName:      "test.db" + "-" + now.Format(time.RFC3339),
		UpdatedTime: now,
	}
	writeConfigFile(&newConfig)

	InitDB()
}

func CloseDB() {
	db.conn.Close()
}
