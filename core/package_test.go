package core

import (
	"fmt"
	"testing"

	"git.neds.sh/technology/pricekinetics/tools/codetest/model"
)

// Task 01 Tests
// TestConvertFromModelDisplay checks that the API response
// contains the correct Display value from the event model
func TestConvertFromModelDisplay(t *testing.T) {

	// Test different event visibility values
	tests := []struct {
		name     string
		display  *model.OptionalBool
		expected bool
	}{
		{"Visible event", &model.OptionalBool{Value: true}, true},
		{"Hidden event", &model.OptionalBool{Value: false}, false},
		{"Missing Display defaults to false", nil, false},
		{"Deleted Display with false value", &model.OptionalBool{Deleted: true}, false},
	}

	// Run each test scenario
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create an event with the test Display value
			event := &model.Event{
				ID:      "testEvent",
				Display: tt.display,
			}
			// Convert the event model into an API response
			result := &SportEvent{}
			result.ConvertFromModel(event)
			// Check if the API response has the expected Display value
			if result.Display != tt.expected {
				t.Errorf("Expected Display %v, got %v",
					tt.expected, result.Display)
			}
		})
	}
}

// Task 02 Tests
// TestConvertFromModelRacingEvent checks that horse, greyhound, and harness racing events are correctly converted into API responses
func TestConvertFromModelRacingEvent(t *testing.T) {
	// Test different racing types
	tests := []struct {
		name       string
		racingType string
		venue      string
		distance   int64
	}{
		{"Horse racing", "Horse", "Brisbane", 1200},
		{"Greyhound racing", "Greyhound", "Melbourne", 525},
		{"Harness racing", "Harness", "Sydney", 1609},
	}

	// Run each test scenario
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create an event with racing details
			event := &model.Event{
				ID: "race001",
				Name: &model.OptionalString{
					Value: tt.name,
				},
				EventTypeID: &model.OptionalString{
					Value: "racing",
				},
				Display: &model.OptionalBool{
					Value: true,
				},
				RacingData: &model.RacingEvent{
					RaceName:   &model.OptionalString{Value: tt.name},
					RacingType: &model.OptionalString{Value: tt.racingType},
					Distance:   &model.OptionalInt64{Value: tt.distance},
					Venue:      &model.OptionalString{Value: tt.venue},
				},
			}

			// Convert stored event data into the API response
			result := &RacingEvent{}
			result.ConvertFromModel(event)

			// Check racing details in the response
			if result.RacingType != tt.racingType {
				t.Errorf("Expected RacingType %s, got %s", tt.racingType, result.RacingType)
			}

			if result.RaceName != tt.name {
				t.Errorf("Expected RaceName %s, got %s", tt.name, result.RaceName)
			}

			if result.Venue != tt.venue {
				t.Errorf("Expected Venue %s, got %s", tt.venue, result.Venue)
			}

			if result.Distance != fmt.Sprintf("%d", tt.distance) {
				t.Errorf("Expected Distance %d, got %s", tt.distance, result.Distance)
			}

			if !result.Display {
				t.Error("Expected Display to be true")
			}
		})
	}
}
