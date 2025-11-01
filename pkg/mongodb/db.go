package mongodb

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// connectMongoDB establishes a connection to MongoDB and returns the client handle.
func ConnectMongoDB(uri string) (*mongo.Client, error) {
	// Set a timeout for the connection attempt
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel() // Ensure the context is canceled when done

	// Parse the URI to check for authSource.
	// If it's missing, the driver defaults to "admin", but being explicit can solve auth issues.
	u, err := url.Parse(uri)
	if err != nil {
		return nil, fmt.Errorf("failed to parse MongoDB URI: %w", err)
	}

	// If authSource is not specified, explicitly set it to "admin".
	if u.Query().Get("authSource") == "" {
		uri += "?authSource=admin"
		log.Println("`authSource` not specified, appending `?authSource=admin` to the URI.")
	}

	// Set client options
	clientOptions := options.Client().ApplyURI(uri)

	// Connect to MongoDB
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("error connecting to MongoDB: %w", err)
	}

	// Ping the primary server to verify the connection
	// Using readpref.Primary() ensures it attempts to reach the main server
	if err = client.Ping(ctx, readpref.Primary()); err != nil {
		// Disconnect if ping fails (though defer handles the final closure)
		client.Disconnect(context.Background())
		return nil, fmt.Errorf("error pinging MongoDB: %w", err)
	}

	// NOTE: For a persistent application (like a web server), you'd typically manage
	// the client lifecycle differently, keeping it open globally and only closing
	// it on application shutdown. The defer above is simplified for this example.

	return client, nil
}
