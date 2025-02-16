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
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Printf("Failed to open database at path: %s", dbPath)
		log.Printf("Error: %s", err)
		return err
	}
	defer db.Close()
	return nil
}

func LoadBaseSkills() ([]*models.Skill, error) {
	db, err := sql.Open("sqlite3", getDatabasePath())
	if err != nil {
		log.Printf("Failed to open database at path: %s", getDatabasePath())
		log.Printf("Error: %s", err)
		return nil, err
	}
	defer db.Close()
	rows, err := db.Query("select id, name, description from skills")
	if err != nil {
		log.Printf("Failed to load Skills: %v", err)
		return nil, err
	}

	log.Printf("Database Connection open: %v, Connections: %v, %v", db.Stats().Idle, db.Stats().OpenConnections, rows.Err())

	defer rows.Close()
	var skills []*models.Skill
	for rows.Next() {
		var skill models.Skill
		if err := rows.Scan(&skill.Id, &skill.Name, &skill.Description); err != nil {
			log.Printf("Failed to load Skills: %v", err)
			return nil, err
		}
		log.Printf("Loaded Skill: %s", skill.Name)
		skills = append(skills, &skill)
	}
	return skills, nil
}
