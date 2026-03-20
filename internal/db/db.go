package db

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

func InitDB() *sql.DB {
	database, err := sql.Open("sqlite3", "./momentum.db")
	if err != nil {
		log.Fatal(err)
	}

	createTableSQL := `
	CREATE TABLE IF NOT EXISTS tasks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		description TEXT,
		completed BOOLEAN NOT NULL DEFAULT 0,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);
	`

	_, err = database.Exec(createTableSQL)
	if err != nil {
		log.Fatal(err)
	}

	columns, err := database.Query("PRAGMA table_info(tasks)")
	if err != nil {
		log.Fatal(err)
	}
	defer columns.Close()

	hasCreatedAt := false
	for columns.Next() {
		var cid int
		var name string
		var ctype string
		var notNull int
		var defaultValue any
		var pk int

		err = columns.Scan(&cid, &name, &ctype, &notNull, &defaultValue, &pk)
		if err != nil {
			log.Fatal(err)
		}

		if name == "created_at" {
			hasCreatedAt = true
			break
		}
	}

	if !hasCreatedAt {
		_, err = database.Exec(`
			ALTER TABLE tasks ADD COLUMN created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP;
		`)
		if err != nil {
			log.Fatal(err)
		}
	}

	log.Println("Database initialized")

	return database
}
