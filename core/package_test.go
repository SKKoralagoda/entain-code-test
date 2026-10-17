package core

import (
	"testing"

	"git.neds.sh/technology/pricekinetics/tools/codetest/model"
)

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
