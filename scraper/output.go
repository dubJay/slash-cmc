package scraper

import (
	"fmt"
	"time"
	
	"github.com/arran4/golang-ical"
)

const eventPage = "https://www.cmc.org/EventDetails.aspx?ID=%s"

func CalendarEntriesToICS(entries CalendarEntries) string {
	cal := ics.NewCalendar()
	cal.SetMethod(ics.MethodRequest)

	now := time.Now()
	for _, entry := range entries {
		event := cal.AddEvent(entry.EventID)
		event.SetCreatedTime(now)
		event.SetModifiedAt(now)
		event.SetAllDayStartAt(entry.Date)
		event.SetAllDayEndAt(entry.Date)
		event.SetSummary(entry.Title)
		event.SetDescription(fmt.Sprintf("Type: %s\nSpots remaining: %d", entry.TripType, entry.Remaining))
		event.SetURL(fmt.Sprintf(eventPage, entry.EventID))
	}

	return cal.Serialize()
}
