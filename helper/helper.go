package helper

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

func CreateDB(db *sql.DB, dbName string) error {
	// Check if the database exists
	var exists bool
	query := "SELECT EXISTS(SELECT datname FROM pg_catalog.pg_database WHERE datname = $1)"
	err := db.QueryRow(query, dbName).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check if database exists: %w", err)
	}

	if !exists {
		//create the new database
		createQuery := fmt.Sprintf("CREATE DATABASE %s", dbName)
		_, err = db.Exec(createQuery)
		if err != nil {
			return fmt.Errorf("failed to create database: %w", err)
		}

		fmt.Printf("Created database %s successfully\n", dbName)
	} else {
		fmt.Printf("Database %s already exists\n", dbName)
	}

	return nil
}

func CreateTasksTable(db *sql.DB) error {
	query1 := `
		CREATE TABLE IF NOT EXISTS tasks(
		id SERIAL PRIMARY KEY,
		title VARCHAR(100) NOT NULL,
		description VARCHAR(300) NOT NULL,
		completed BOOLEAN DEFAULT FALSE,
		created_at TIMESTAMP DEFAULT NOW(),
		updated_at TIMESTAMP DEFAULT NOW()
		);
	`
	_, err := db.Exec(query1)
	if err != nil {
		return fmt.Errorf("failed to create tasks table: %w", err)
	}

	query2 := `	
		CREATE OR REPLACE FUNCTION set_updated_at()
		RETURNS TRIGGER AS $$
		BEGIN
		NEW.updated_at = NOW();
		RETURN NEW;
		END;
		$$ LANGUAGE plpgsql;
	`
	_, err2 := db.Exec(query2)
	if err2 != nil {
		return err2
	}

	query3 := `
	DROP TRIGGER IF EXISTS trigger_set_updated_at ON tasks;
	`
	_, err3 := db.Exec(query3)
	if err3 != nil {
		return err3
	}

	query4 := `
		CREATE TRIGGER trigger_set_updated_at
		BEFORE UPDATE ON tasks
		FOR EACH ROW
		EXECUTE FUNCTION set_updated_at();
	`
	_, err4 := db.Exec(query4)
	if err4 != nil {
		return err4
	}
	fmt.Println("Tasks table created successfully")

	return nil
}
