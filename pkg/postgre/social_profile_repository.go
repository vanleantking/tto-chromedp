package postgre

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"
	"tto_chromedp/pkg/models"
)

type SocialProfileRepository interface {
	UpsertContentInterestsAndGetIDs(contentInterests []string, userID int) (map[string]int, error)
	UpsertBrandsAndGetIDs(brandNames []string, userID int, clientID int) (map[string]int, error)
	UpdateTTOUser(userID int, identityData map[string]interface{}) error
	GetSocialProfileCrawlTTO() ([]models.SocialProfile, error)
	Close() error
}

type socialProfileRepository struct {
	db *sql.DB
}

func NewSocialProfileRepository(db *sql.DB) SocialProfileRepository {
	return &socialProfileRepository{db: db}
}

// UpsertContentInterestsAndGetIDs inserts new content interests or does nothing if they exist,
// then retrieves the IDs for all given names.
func (sp *socialProfileRepository) UpsertContentInterestsAndGetIDs(contentInterests []string, userID int) (map[string]int, error) {
	if len(contentInterests) == 0 {
		return map[string]int{}, nil
	}

	// Start a transaction for atomicity
	tx, err := sp.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	// Defer a rollback in case of error. If commit succeeds, this is ignored.
	defer tx.Rollback()

	currentTime := time.Now()
	statusValue := 1
	numFields := 6 // name, created_by, updated_by, created_at, updated_at, status

	// --- 1. Build and Execute Batched INSERT ... ON CONFLICT ---

	// Dynamically build the VALUES clause placeholders ($1, $2, $3, ...), e.g., ($1, $2, $3), ($4, $5, $6)
	var valueStrings []string
	var valueArgs []interface{}
	placeholderIdx := 1

	for _, name := range contentInterests {
		// Construct the placeholder group for one row: ($N, $N+1, $N+2, ...)
		valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d)",
			placeholderIdx, placeholderIdx+1, placeholderIdx+2, placeholderIdx+3, placeholderIdx+4, placeholderIdx+5))

		// Collect the actual values in order
		valueArgs = append(valueArgs, name, userID, userID, currentTime, currentTime, statusValue)

		placeholderIdx += numFields
	}

	insertQuery := fmt.Sprintf(`
		INSERT INTO cms.content_interest (
			name, created_by, updated_by, created_at, updated_at, status
		)
		VALUES %s
		ON CONFLICT (name) DO NOTHING;
	`, strings.Join(valueStrings, ","))

	// Execute the batched insert within the transaction
	if _, err := tx.Exec(insertQuery, valueArgs...); err != nil {
		return nil, fmt.Errorf("batched upsert failed: %w", err)
	}

	// --- 2. SELECT IDs for all given names ---

	contentMap := make(map[string]int)

	// Use PostgreSQL's ANY() clause with the contentInterests slice (passed as $1)
	// The lib/pq driver automatically handles passing a Go slice as a PostgreSQL array.
	selectQuery := "SELECT id, name FROM cms.content_interest WHERE name = ANY($1);"

	rows, err := tx.Query(selectQuery, contentInterests)
	if err != nil {
		return nil, fmt.Errorf("select IDs query failed: %w", err)
	}
	defer rows.Close() // Ensure rows are closed immediately after function return

	for rows.Next() {
		var id int
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		contentMap[name] = id
	}

	// Check for any error encountered during iteration
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during row iteration: %w", err)
	}

	// --- 3. Commit Transaction ---
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return contentMap, nil
}

// UpsertBrandsAndGetIDs inserts new brand names or does nothing if they exist,
// then retrieves the IDs for all given names.
func (sp *socialProfileRepository) UpsertBrandsAndGetIDs(brandNames []string, userID int, clientID int) (map[string]int, error) {
	if len(brandNames) == 0 {
		return map[string]int{}, nil
	}

	tx, err := sp.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	currentTime := time.Now()
	positioningValue := 1
	statusValue := 1
	// title, positioning, created_by, updated_by, created_at, updated_at, client_id, status
	numFields := 8

	// --- 1. Build and Execute Batched INSERT ... ON CONFLICT ---
	var valueStrings []string
	var valueArgs []interface{}
	placeholderIdx := 1

	for _, name := range brandNames {
		valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d)",
			placeholderIdx, placeholderIdx+1, placeholderIdx+2, placeholderIdx+3, placeholderIdx+4, placeholderIdx+5, placeholderIdx+6, placeholderIdx+7))

		// Collect the actual values in order
		valueArgs = append(valueArgs, name, positioningValue, userID, userID, currentTime, currentTime, clientID, statusValue)

		placeholderIdx += numFields
	}

	insertQuery := fmt.Sprintf(`
		INSERT INTO cms.brands (
			title, positioning, created_by, updated_by, created_at, updated_at, client_id, status
		)
		VALUES %s
		ON CONFLICT (title) DO NOTHING;
	`, strings.Join(valueStrings, ","))

	if _, err := tx.Exec(insertQuery, valueArgs...); err != nil {
		return nil, fmt.Errorf("batched brand upsert failed: %w", err)
	}

	// --- 2. SELECT IDs for all given names ---

	brandMap := make(map[string]int)

	selectQuery := "SELECT id, title FROM cms.brands WHERE title = ANY($1);"

	rows, err := tx.Query(selectQuery, brandNames)
	if err != nil {
		return nil, fmt.Errorf("select brand IDs query failed: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var title string
		if err := rows.Scan(&id, &title); err != nil {
			return nil, fmt.Errorf("failed to scan brand row: %w", err)
		}
		brandMap[title] = id
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during brand row iteration: %w", err)
	}

	// --- 3. Commit Transaction ---
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit brand transaction: %w", err)
	}

	return brandMap, nil
}

// UpdateTTSUser updates the social_profiles table with the processed identity data for a user.
func (sp *socialProfileRepository) UpdateTTOUser(userID int, identityData map[string]interface{}) error {
	// Map identity_data keys to database column names
	dbMapping := map[string]string{
		"content_interest":          "content_interest",
		"audience_age":              "content_interest",
		"audience_location":         "content_interest",
		"audience_gender":           "content_interest",
		"kol_growth":                "content_interest",
		"tiktokshop_updated_at":     "tiktokshop_updated_at",
		"tiktokshop_creator_status": "tiktokshop_creator_status",
	}

	// Build the SET part of the query dynamically
	var setClauses []string
	var params []interface{}
	placeholderIdx := 1

	for dataKey, dbCol := range dbMapping {
		value, ok := identityData[dataKey]
		if !ok || value == nil {
			continue // Skip if key is missing or value is nil
		}

		var paramValue interface{}

		// Check if the value is a complex type (map or slice) that needs JSON marshaling
		if _, isMap := value.(map[string]interface{}); isMap {
			// Convert complex types to JSON byte slice
			jsonValue, err := json.Marshal(value)
			if err != nil {
				log.Printf("Warning: Failed to marshal key %s to JSON: %v. Skipping.", dataKey, err)
				continue
			}
			paramValue = jsonValue
		} else if _, isSlice := value.([]interface{}); isSlice {
			// Convert complex types to JSON byte slice
			jsonValue, err := json.Marshal(value)
			if err != nil {
				log.Printf("Warning: Failed to marshal key %s to JSON: %v. Skipping.", dataKey, err)
				continue
			}
			paramValue = jsonValue
		} else {
			// Use standard value for simple types, including time.Time objects
			paramValue = value
		}

		params = append(params, paramValue)
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", dbCol, placeholderIdx))
		placeholderIdx++
	}

	if len(setClauses) == 0 {
		return nil // Nothing to update
	}

	// Append the WHERE clause parameter (user_id) and updated_at
	setClauses = append(setClauses, fmt.Sprintf("updated_at = $%d", placeholderIdx))
	params = append(params, time.Now())
	placeholderIdx++

	updateQuery := fmt.Sprintf(`
		UPDATE crawler.social_profiles
		SET %s
		WHERE id = $%d;`, strings.Join(setClauses, ", "), placeholderIdx)
	params = append(params, userID)
	log.Printf("Executing update query: %s with params (excluding userID): %v", updateQuery, params[:len(params)-1]) // Log query for debugging

	// Execute the query within a transaction for safety
	tx, err := sp.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start update transaction: %w", err)
	}
	// Use an anonymous function in defer to ensure tx.Rollback() only runs if Commit fails or an error occurs before it.
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r) // Re-throw the panic
		} else if err != nil { // Check if err was set before the commit
			tx.Rollback()
		}
	}()

	_, err = tx.Exec(updateQuery, params...)
	if err != nil {
		// The defer function will handle rollback since err is set here
		return fmt.Errorf("failed to execute update query: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		// The defer function will handle rollback since err is set here
		return fmt.Errorf("failed to commit update transaction: %w", err)
	}

	return nil
}

func (sp *socialProfileRepository) GetSocialProfileCrawlTTO() ([]models.SocialProfile, error) {
	sqlQuery := `SELECT id, username FROM crawler.social_profiles WHERE tiktokshop_creator_status = $1 LIMIT $2;`
	rows, err := sp.db.QueryContext(context.Background(), sqlQuery, -1, 200)
	if err != nil {
		return nil, fmt.Errorf("failed to query social profiles: %w", err)
	}
	defer rows.Close()

	var profiles []models.SocialProfile
	for rows.Next() {
		var profile models.SocialProfile
		if err := rows.Scan(&profile.ID, &profile.UserName); err != nil {
			// Return profiles found so far along with the error
			return profiles, fmt.Errorf("failed to scan social profile row: %w", err)
		}
		profiles = append(profiles, profile)
	}

	if err := rows.Err(); err != nil {
		return profiles, fmt.Errorf("error during rows iteration: %w", err)
	}

	log.Printf("Fetched %d social profiles for TTO crawling.", len(profiles))
	return profiles, nil
}

func (sp *socialProfileRepository) Close() error {
	return sp.db.Close()
}
