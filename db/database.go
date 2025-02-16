package db

import (
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	models "krumpin/models"
	"log"
	"os"
	"path/filepath"
	"runtime"
)

//go:embed krumpin.db
var embeddedDb embed.FS

var db *sql.DB

func getDatabasePath() string {
	var configDir string
	if runtime.GOOS == "windows" {
		configDir = filepath.Join(os.Getenv("APPDATA"), "Krumpin")
	} else {
		configDir = filepath.Join(os.Getenv("HOME"), ".config", "krumpin")
	}

	// Ensure the directory exists
	if err := os.MkdirAll(configDir, os.ModePerm); err != nil {
		panic(fmt.Sprintf("Failed to create config directory: %v", err))
	}

	return filepath.Join(configDir, "krumpin.db")
}

func extractDatabase(dbPath string, db embed.FS) {
	log.Printf("Attemping to extract database from %s", dbPath)
	data, err := fs.ReadFile(db, "krumpin.db")
	if err != nil {
		panic(fmt.Sprintf("Failed to read embedded database: %v", err))
	}

	err = os.WriteFile(dbPath, data, 0644)
	if err != nil {
		panic(fmt.Sprintf("Failed to write database to disk: %v", err))
	}
}

func InitialiseDb() error {
	log.Printf("Attemping to set up database")
	dbPath := getDatabasePath()
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		log.Printf("Database was not found at path: %s", dbPath)
		extractDatabase(dbPath, embeddedDb)
	}
	log.Printf("Using database at path: %s", dbPath)
	var err error
	db, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Printf("Failed to open database at path: %s", dbPath)
		log.Printf("Error: %s", err)
		return err
	}
	return nil
}

func LoadBaseSkills() ([]*models.Skill, error) {
	rows, err := db.Query("select name, id, description from skills")
	if err != nil {
		log.Printf("Failed to load Skills: %v", err)
		return nil, err
	}
	next := rows.NextResultSet()
	cols, err := rows.Columns()
	log.Printf("Loaded Skills: %s", next)
	log.Printf("Loaded Skills: %s", cols)
	defer rows.Close()

	var skills []*models.Skill
	for rows.Next() {
		var skill models.Skill
		var name string
		var id int64
		var description string
		err := rows.Scan(&id, &name, &description)
		if err != nil {
			log.Printf("Failed to scan row: %v", err)
			return nil, err
		}
		log.Printf("Name: %s, id: %d, description: %s", name, id, description)
		if err := rows.Scan(&skill.Name, &skill.Id, &skill.Description); err != nil {
			log.Printf("Failed to load Skills: %v", err)
			return nil, err
		}
		log.Printf("Loaded Skill: %s", skill.Name)
		skills = append(skills, &skill)
	}
	return skills, nil
}
