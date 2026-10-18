package repository

import (
	"context"
	"fmt"

	"encoding/json"
	"git.neds.sh/technology/pricekinetics/tools/codetest/model"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

// mongoRepo stores events in MongoDB
type mongoRepo struct {
	client     *mongo.Client
	collection *mongo.Collection
}

// NewMongoRepository connects to MongoDB
func NewMongoRepository(ctx context.Context, uri string) (Repository, error) {
	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}

	repo := &mongoRepo{
		client:     client,
		collection: client.Database("codetest").Collection("events"),
	}

	if !repo.HealthCheck(ctx) {
		_ = client.Disconnect(ctx)
		return nil, fmt.Errorf("failed to connect to MongoDB")
	}

	return repo, nil
}

// HealthCheck checks the MongoDB connection
func (r *mongoRepo) HealthCheck(ctx context.Context) bool {
	return r.client.Ping(ctx, readpref.Primary()) == nil
}

// UpdateEvent saves or replaces an event in MongoDB.
func (r *mongoRepo) UpdateEvent(ctx context.Context, event *model.Event) error {
	// Convert the event into JSON.
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	// Save the event using its ID.
	filter := bson.M{"_id": event.ID}
	document := bson.M{
		"_id":  event.ID,
		"data": string(data),
	}

	_, err = r.collection.ReplaceOne(
		ctx,
		filter,
		document,
		options.Replace().SetUpsert(true),
	)
	return err
}

// GetEventByID retrieves an event from MongoDB
func (r *mongoRepo) GetEventByID(ctx context.Context, id string) (*model.Event, error) {
	var document struct {
		Data string `bson:"data"`
	}

	// Find the event by ID
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&document)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	// Convert the stored JSON back into an event
	event := &model.Event{}
	if err := json.Unmarshal([]byte(document.Data), event); err != nil {
		return nil, err
	}

	return event, nil
}

// DeleteEventByID removes an event from MongoDB
func (r *mongoRepo) DeleteEventByID(ctx context.Context, id string) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

// SearchEvents retrieves all stored events from MongoDB.
func (r *mongoRepo) SearchEvents(ctx context.Context) ([]*model.Event, error) {
	// Find all documents in the events collection.
	// An empty filter means retrieve every document.
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}

	// Close the cursor when the function finishes.
	defer cursor.Close(ctx)

	// Store the events retrieved from MongoDB.
	var events []*model.Event

	// Loop through each document returned by MongoDB.
	for cursor.Next(ctx) {
		// Define the structure of the stored JSON data.
		var document struct {
			Data string `bson:"data"`
		}
		// Read the current MongoDB document.
		if err := cursor.Decode(&document); err != nil {
			return nil, err
		}
		// Create an empty event.
		event := &model.Event{}
		// Convert the stored JSON data into a Go event model.
		if err := json.Unmarshal([]byte(document.Data), event); err != nil {
			return nil, err
		}
		// Add the converted event to the results.
		events = append(events, event)
	}
	// Check if any errors occurred while reading documents.
	if err := cursor.Err(); err != nil {
		return nil, err
	}
	// Return all events to the service for filtering.
	return events, nil
}
