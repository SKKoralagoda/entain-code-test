// Package core contains the proto definitions of the Core Service
package core

import (
	"fmt"
	"time"

	"git.neds.sh/technology/pricekinetics/tools/codetest/model"
)

//go:generate ./gen-proto.sh

// ConvertFromModel converts a model.Event to a core.SportEvent
func (to *SportEvent) ConvertFromModel(model *model.Event) {
	to.ID = model.ID
	to.Name = model.GetName().GetValue()
	to.StartTime = time.Unix(0, model.StartTime.GetValue()).Format(time.RFC3339)
	to.BettingStatus = model.GetBettingStatus().GetValue().String()
	to.SportTypeID = model.GetEventTypeID().GetValue()
	to.Markets = model.Markets
	to.League = model.GetSportData().GetLeague().GetValue()
	to.SportName = model.GetSportData().GetName().GetValue()
	to.Round = model.GetSportData().GetRound().GetValue()
	to.Region = model.GetSportData().GetRegion().GetValue()
	to.Display = model.GetDisplay().GetValue()
}

// ConvertFromModel converts a model.Event into a core.RacingEvent.
func (to *RacingEvent) ConvertFromModel(event *model.Event) {
	to.ID = event.ID
	to.Name = event.GetName().GetValue()
	// Set StartTime only when provided
	if event.GetStartTime() != nil && !event.GetStartTime().GetDeleted() {
		to.StartTime = time.Unix(0, event.GetStartTime().GetValue()).Format(time.RFC3339)
	}
	to.BettingStatus = event.GetBettingStatus().GetValue().String()
	to.Markets = event.Markets
	to.SportTypeID = event.GetEventTypeID().GetValue()

	// Racing-specific information
	to.RacingType = event.GetRacingData().GetRacingType().GetValue()
	to.RaceName = event.GetRacingData().GetRaceName().GetValue()
	to.Distance = fmt.Sprintf("%d", event.GetRacingData().GetDistance().GetValue())
	to.Venue = event.GetRacingData().GetVenue().GetValue()

	// Event visibility
	to.Display = event.GetDisplay().GetValue()
}
