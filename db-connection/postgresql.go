package db_connection

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

// DBConnection holds database configuration.
type DBConnection struct {
	Host     string
	Port     int
	Username string
	Password string
	Database string
	SSLMode  string
}

// Connect opens and verifies a PostgreSQL connection.
func Connect(conn DBConnection) (*sql.DB, error) {
	log.Printf("Connecting to postgresql://%s:%s@%s:%d/%s?sslmode=%s", conn.Username, conn.Password, conn.Host, conn.Port, conn.Database, conn.SSLMode)
	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		conn.Host,
		conn.Port,
		conn.Username,
		conn.Password,
		conn.Database,
		conn.SSLMode,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, err
	}

	return db, nil
}
