package merger

import (
	"context"
	"testing"

	"git.neds.sh/technology/pricekinetics/tools/codetest/model"
)

// TestMergeEventDisplay checks that Display updates correctly
// without losing existing event information
func TestMergeEventDisplay(t *testing.T) {

	// Test different Display update scenarios
	tests := []struct {
		name     string
		existing *model.OptionalBool
		incoming *model.OptionalBool
		expected bool
	}{
		{"Hide event", &model.OptionalBool{Value: true}, &model.OptionalBool{Value: false}, false},
		{"Show event", &model.OptionalBool{Value: false}, &model.OptionalBool{Value: true}, true},
		{"Keep existing value", &model.OptionalBool{Value: true}, nil, true},
		{"Default to false", nil, nil, false},
	}

	// Run each test scenario
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Existing event data (old data)
			existing := &model.Event{
				ID:      "testEvent",
				Name:    &model.OptionalString{Value: "Test Event"},
				Display: tt.existing,
			}
			// Incoming event update (new data)
			incoming := &model.Event{
				ID:      "testEvent",
				Display: tt.incoming,
			}
			// Merge the existing event with the new update
			result := MergeEvent(context.Background(), existing, incoming)
			// Check if Display matches the expected value
			if result.GetDisplay().GetValue() != tt.expected {
				t.Errorf("Expected Display %v, got %v",
					tt.expected, result.GetDisplay().GetValue())
			}
			// Check that the existing event name remains unchanged.
			if result.GetName().GetValue() != "Test Event" {
				t.Error("Existing event name was changed")
			}
		})
	}
}

// TestMergeEventDisplayNotProvided checks that a missing Display value remains nil during merging
func TestMergeEventDisplayNotProvided(t *testing.T) {
	// Create the existing event without a Display value
	existing := &model.Event{
		ID: "testEvent",
	}
	// Create the incoming event update without a Display value
	incoming := &model.Event{
		ID: "testEvent",
	}
	// Merge the existing and incoming events
	result := MergeEvent(context.Background(), existing, incoming)
	// Check that Display remains unset
	if result.Display != nil {
		t.Error("Display should remain unset")
	}
}
