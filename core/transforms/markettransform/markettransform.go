package markettransform

import (
	"context"
	"time"

	"git.neds.sh/technology/pricekinetics/tools/codetest/core/transforms"
	"git.neds.sh/technology/pricekinetics/tools/codetest/model"
)

// markettransformClient handles market-related transformations
type markettransformClient struct{}

// NewMarketTransformClient creates a new market transform client
func NewMarketTransformClient() transforms.TransformClient {
	return &markettransformClient{}
}

// TransformEvent records when a market closes for the first time
func (t *markettransformClient) TransformEvent(_ context.Context, partialUpdate, fullModel *model.Event) (*model.Event, error) {

	// Store markets that need a ClosedAt timestamp
	var updatedMarkets []*model.Market

	// Check each market in the event.
	for _, market := range fullModel.GetMarkets() {

		// Skip markets that are not closed
		if market.GetBettingStatus().GetValue() != model.BettingStatus_BettingClosed {
			continue
		}

		// Keep the existing timestamp if already recorded
		if market.GetClosedAt() != nil && !market.GetClosedAt().GetDeleted() {
			continue
		}

		// Record the current time for a closed market
		updatedMarkets = append(updatedMarkets, &model.Market{
			ID:       market.ID,
			ClosedAt: &model.OptionalInt64{Value: time.Now().UnixNano()},
		})
	}

	// No markets need updating
	if len(updatedMarkets) == 0 {
		return nil, nil
	}

	// Return only the market changes to be merged with the event
	return &model.Event{
		ID:      fullModel.ID,
		Markets: updatedMarkets,
	}, nil
}

// GetName returns the transform name
func (t *markettransformClient) GetName() string {
	return "MarketTransform"
}
