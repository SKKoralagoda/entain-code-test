package service_test

import (
	"context"
	"testing"

	"git.neds.sh/technology/pricekinetics/tools/codetest/core"
	"git.neds.sh/technology/pricekinetics/tools/codetest/core/repository"
	"git.neds.sh/technology/pricekinetics/tools/codetest/core/service"
	"git.neds.sh/technology/pricekinetics/tools/codetest/core/transforms"
	"git.neds.sh/technology/pricekinetics/tools/codetest/core/transforms/sporttransform"
	"git.neds.sh/technology/pricekinetics/tools/codetest/merger"
	"git.neds.sh/technology/pricekinetics/tools/codetest/model"
	"github.com/stretchr/testify/assert"
)

func TestService_IntegrationTest_NewEvent(t *testing.T) {
	repo, err := repository.NewMongoRepository(context.Background(), "mongodb://localhost:27018")
	assert.NoError(t, err)
	defer repo.DeleteEventByID(context.Background(), "integration-test-1")
	host := &service.Service{
		Upstreams: &service.Upstreams{
			MergerClient: merger.NewInlineMergerClient(),
			Repo:         repo,
			Transforms: []transforms.TransformClient{
				sporttransform.NewSportTransformClient(),
			},
		},
	}

	output, err := host.Update(context.Background(), &core.UpdateRequest{
		Event: &model.Event{
			ID:          "integration-test-1",
			Name:        &model.OptionalString{Value: "Test event"},
			EventTypeID: &model.OptionalString{Value: "soccer"},
			StartTime:   &model.OptionalInt64{Value: 1758244443000000000}, // Friday, September 19, 2025 11:14:03 AM GMT+10:00
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, "New Event born integration-test-1", output.Message)

	output, err = host.Update(context.Background(), &core.UpdateRequest{
		Event: &model.Event{
			ID: "integration-test-1",
			Markets: []*model.Market{
				{
					ID:   "mkt01",
					Name: &model.OptionalString{Value: "New Market"},
				},
			},
		},
	})

	assert.NoError(t, err)
	assert.Equal(t, "Success", output.Message)

	final, err := host.GetSportEvent(context.Background(), &core.GetSportEventRequest{EventID: "integration-test-1"})
	assert.NoError(t, err)
	assert.Equal(t, "Test event", final.Event.Name)
	assert.Contains(t, final.Event.StartTime, "2025-09-19")
	assert.Equal(t, "soccer", final.Event.SportTypeID)
	assert.Equal(t, "Soccer", final.Event.SportName)
	assert.Equal(t, "New Market", final.Event.Markets[0].Name.Value)
}

// TestSearchEvents checks searching with optional filters.
func TestSearchEvents(t *testing.T) {
	ctx := context.Background()

	// Connect to the test database.
	repo, err := repository.NewMongoRepository(ctx, "mongodb://localhost:27018")
	if !assert.NoError(t, err) {
		return
	}

	// Create two test events.
	events := []*model.Event{
		{
			ID:            "search-test-open",
			StartTime:     &model.OptionalInt64{Value: 1792454400000000000},
			BettingStatus: &model.OptionalBettingStatus{Value: model.BettingStatus_BettingOpen},
			Display:       &model.OptionalBool{Value: true},
		},
		{
			ID:            "search-test-closed",
			StartTime:     &model.OptionalInt64{Value: 1792540800000000000},
			BettingStatus: &model.OptionalBettingStatus{Value: model.BettingStatus_BettingClosed},
			Display:       &model.OptionalBool{Value: false},
		},
	}

	// Remove test events after testing.
	t.Cleanup(func() {
		for _, event := range events {
			_ = repo.DeleteEventByID(context.Background(), event.ID)
		}
	})

	for _, event := range events {
		if !assert.NoError(t, repo.UpdateEvent(ctx, event)) {
			return
		}
	}

	host := &service.Service{
		Upstreams: &service.Upstreams{
			Repo: repo,
		},
	}

	// Test individual and combined filters.
	tests := []struct {
		name     string
		request  *core.SearchEventsRequest
		expected []string
	}{
		{"Date", &core.SearchEventsRequest{Date: strPtr("2026-10-20")}, []string{"search-test-open"}},
		{"BettingStatus", &core.SearchEventsRequest{BettingStatus: statusPtr(model.BettingStatus_BettingClosed)}, []string{"search-test-closed"}},
		{"Display true", &core.SearchEventsRequest{Display: boolPtr(true)}, []string{"search-test-open"}},
		{"Display false", &core.SearchEventsRequest{Display: boolPtr(false)}, []string{"search-test-closed"}},
		{"Date and Status", &core.SearchEventsRequest{Date: strPtr("2026-10-20"), BettingStatus: statusPtr(model.BettingStatus_BettingOpen)}, []string{"search-test-open"}},
		{"Date and Display", &core.SearchEventsRequest{Date: strPtr("2026-10-20"), Display: boolPtr(true)}, []string{"search-test-open"}},
		{"Status and Display", &core.SearchEventsRequest{BettingStatus: statusPtr(model.BettingStatus_BettingOpen), Display: boolPtr(true)}, []string{"search-test-open"}},
		{"All filters", &core.SearchEventsRequest{Date: strPtr("2026-10-20"), BettingStatus: statusPtr(model.BettingStatus_BettingOpen), Display: boolPtr(true)}, []string{"search-test-open"}},
		{"No match", &core.SearchEventsRequest{Date: strPtr("2026-10-20"), BettingStatus: statusPtr(model.BettingStatus_BettingClosed)}, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := host.SearchEvents(ctx, tt.request)
			if !assert.NoError(t, err) {
				return
			}

			// Check that expected events are returned.
			var ids []string
			for _, event := range result.Events {
				if event.ID == "search-test-open" || event.ID == "search-test-closed" {
					ids = append(ids, event.ID)
				}
			}

			assert.ElementsMatch(t, tt.expected, ids)
		})
	}

	// Invalid dates should return an error.
	t.Run("Invalid date", func(t *testing.T) {
		_, err := host.SearchEvents(ctx, &core.SearchEventsRequest{
			Date: strPtr("20-10-2026"),
		})
		assert.Error(t, err)
	})
}

// Helper functions for optional protobuf fields.
func strPtr(value string) *string {
	return &value
}

func boolPtr(value bool) *bool {
	return &value
}

func statusPtr(value model.BettingStatus) *model.BettingStatus {
	return &value
}
