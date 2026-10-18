package markettransform

import (
	"context"
	"testing"
	"time"

	"git.neds.sh/technology/pricekinetics/tools/codetest/model"
)

// TestMarketTransform checks that ClosedAt is recorded only once
func TestMarketTransform(t *testing.T) {
	// Check that the original timestamp is kept
	existingClosedAt := time.Now().Add(-time.Hour).UnixNano()

	tests := []struct {
		name          string
		status        model.BettingStatus
		closedAt      *model.OptionalInt64
		expectUpdated bool
	}{
		{
			name:          "Open market",
			status:        model.BettingStatus_BettingOpen,
			expectUpdated: false,
		},
		{
			name:          "Newly closed market",
			status:        model.BettingStatus_BettingClosed,
			expectUpdated: true,
		},
		{
			name:   "Already closed market",
			status: model.BettingStatus_BettingClosed,
			closedAt: &model.OptionalInt64{
				Value: existingClosedAt,
			},
			expectUpdated: false,
		},
	}

	transform := NewMarketTransformClient()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create an event containing one market
			event := &model.Event{
				ID: "task03TestEvent",
				Markets: []*model.Market{
					{
						ID: "market001",
						BettingStatus: &model.OptionalBettingStatus{
							Value: tt.status,
						},
						ClosedAt: tt.closedAt,
					},
				},
			}

			// Run the market transform
			before := time.Now().UnixNano()
			result, err := transform.TransformEvent(
				context.Background(), event, event,
			)
			after := time.Now().UnixNano()

			if err != nil {
				t.Fatalf("TransformEvent failed: %v", err)
			}

			// Open and already closed markets need no update
			if !tt.expectUpdated {
				if result != nil {
					t.Fatal("Expected no update")
				}

				// Confirm the original timestamp was not changed
				if tt.closedAt != nil &&
					event.Markets[0].GetClosedAt().GetValue() != existingClosedAt {
					t.Error("Existing ClosedAt was changed")
				}
				return
			}

			// A newly closed market should receive a timestamp
			if result == nil || len(result.Markets) != 1 {
				t.Fatal("Expected one updated market")
			}

			closedAt := result.Markets[0].GetClosedAt().GetValue()

			// Check that ClosedAt contains the current timestamp
			if closedAt < before || closedAt > after {
				t.Errorf("ClosedAt %d is not within the expected time range", closedAt)
			}
		})
	}
}
