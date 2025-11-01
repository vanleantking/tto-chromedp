package mongodb

import (
	"context"
	"fmt"
	"log"
	"strings"

	"tto_chromedp/pkg/utils"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type CountryDetailRepository interface {
	// Implementation details would go here
	GetCountryCodes(ctx context.Context) (map[string]string, error)
}

type countryRepository struct {
	client     *mongo.Client
	Collection string
}

// Document structure matching your MongoDB data
type CountryData struct {
	IsoCode2        string `bson:"iso_code_2"`
	OfficialCountry string `bson:"official_country"`
}

func NewCountryDetailRepository(db *mongo.Client, collecttion string) CountryDetailRepository {
	return &countryRepository{
		client:     db,
		Collection: collecttion,
	}
}

func (cdr *countryRepository) GetCountryCodes(ctx context.Context) (map[string]string, error) {

	collection := cdr.client.Database("adserver").Collection(cdr.Collection)

	options := options.Find().SetProjection(bson.M{
		"iso_code_2":       1,
		"_id":              0,
		"official_country": 1,
	})

	findCursor, err := collection.Find(ctx, bson.M{}, options)
	if err != nil {
		return nil, fmt.Errorf("failed to execute find query for country codes: %w", err)
	}
	defer findCursor.Close(ctx)

	countryCodes := make(map[string]string)
	for findCursor.Next(ctx) {
		var result CountryData
		if err := findCursor.Decode(&result); err != nil {
			// Log the error but continue processing other documents
			log.Printf("Warning: Failed to decode country data document: %v", err)
			continue // Or return nil, fmt.Errorf(...) if you want to fail the whole operation
		}

		countryName := result.OfficialCountry
		if countryName == "" {
			countryName = utils.COUNTRY_UNKNOWN
		}

		if result.IsoCode2 != "" {
			countryCodes[strings.ToUpper(result.IsoCode2)] = countryName
		}
	}

	if err := findCursor.Err(); err != nil {
		return countryCodes, fmt.Errorf("cursor error after iteration: %w", err)
	}

	return countryCodes, nil
}

func processField(field interface{}) (bool, string) {
	fmt.Println(field)
	// The 'type' keyword inside the switch statement makes it a type switch.
	switch field.(type) {
	case string:
		return true, field.(string)
	default:
		return false, ""
	}
}
