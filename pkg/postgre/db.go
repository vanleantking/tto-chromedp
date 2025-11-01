package postgre

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq" // PostgreSQL driver
)

// Define the structure for the database credentials
type DBCredentials struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// initDB creates a DSN (Data Source Name) string and establishes the database connection.
func InitDB(creds DBCredentials) (*sql.DB, error) {
	fmt.Println("enter here")
	// DSN format: "user=USER password=PASSWORD host=HOST port=PORT dbname=DBNAME sslmode=disable"
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		creds.Host, creds.Port, creds.User, creds.Password, creds.DBName, creds.SSLMode)

	fmt.Println("Attempting to connect to PostgreSQL...", creds.Host, creds.Port, creds.DBName, creds.User, creds.Password, creds.SSLMode)

	// Open the connection. The database connection is not established immediately here.
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %w", err)
	}

	// Ping the database to verify the connection is active and credentials are valid.
	if err := db.Ping(); err != nil {
		db.Close() // Close connection if ping fails
		return nil, fmt.Errorf("error connecting to database (ping failed): %w", err)
	}

	// Set connection pool parameters (recommended for production)
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	log.Println("Successfully connected to PostgreSQL database!")
	return db, nil
}

func main() {
	// --- Load Environment Variables ---
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: Could not load .env file: %v", err)
	}

	// --- 1. Get Credentials from Environment ---
	creds := DBCredentials{
		Host:     os.Getenv("PG_HOST"),
		Port:     os.Getenv("PG_PORT"),
		User:     os.Getenv("PG_USER"),
		Password: os.Getenv("PG_PASSWORD"),
		DBName:   os.Getenv("PG_DBNAME"),
		SSLMode:  os.Getenv("PG_SSLMODE"), // e.g., "disable", "require", "verify-full"
	}

	// Simple check for required credentials
	if creds.User == "" || creds.DBName == "" {
		log.Fatalf("Fatal: PG_USER and PG_DBNAME must be set in the environment or .env file.")
	}

	// --- 2. Initialize Database Connection ---
	db, err := InitDB(creds)
	if err != nil {
		log.Fatalf("Fatal: Database initialization failed: %v", err)
	}
	defer db.Close() // Ensure the connection is closed when main exits

	// --- 3. Execute a Simple Query ---
	var serverTime time.Time

	// Query the PostgreSQL server for its current time
	query := "SELECT now()"

	err = db.QueryRow(query).Scan(&serverTime)
	if err != nil {
		log.Fatalf("Fatal: Failed to execute query: %v", err)
	}

	log.Printf("PostgreSQL Server Time: %s", serverTime.Format(time.RFC3339))
}
