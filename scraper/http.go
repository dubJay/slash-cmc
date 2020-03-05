package scraper

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/gocolly/colly"
)

var calendarURL = "https://cmc.org/calendar"
var eventBase = "EventDetails.aspx?ID="

// Metadata associated with an entry in the CMC calendar.
type CalendarEntry struct {
	Title string
	EventID string
	Date time.Time
	Remaining int
	TripType string
}

type CalendarEntries []CalendarEntry

// Builds a map of all required form parameters to load the list view from the cmc calendar. I tried a wide
// array of permutations to widdle it down to these key value pairs.
func buildMultipartFormWithViewState(viewState, generator, eventValidation string) map[string][]byte {
	empty := []byte("")
	return map[string][]byte {
		// All of the __ prefixed keys are required by ASP.NET, even the empty encrypted one.
		"__VIEWSTATE": []byte(viewState),
		"__VIEWSTATEGENERATOR": []byte(generator),
		"__EVENTVALIDATION": []byte(eventValidation),
		"__VIEWSTATEENCRYPTED": empty,
		// I have no idea what these two parameters control but if they're not set the aspx page fails to
		// load. Changing their values doesn't seem to have any effect.
		"dnn$ctr907$EventList$btnListView.x": []byte("0"),
		"dnn$ctr907$EventList$btnListView.y": []byte("100"),
	}
	
}

func calendarEntryFrom(row *colly.HTMLElement) (CalendarEntry, error) {
	rawTitle := row.ChildText("td a")
	rawEventID := row.ChildAttr("td a", "href")
	rawDate := row.ChildText("td:nth-child(2)")
	rawRemaining:= row.ChildText("td:nth-child(3)")
	for _, value := range []string{rawTitle, rawEventID, rawDate, rawRemaining} {
		if value == "" {
			return CalendarEntry{}, fmt.Errorf("Required metadata missing. Title: %s, EventId: %s, Date: %s, Remaining: %s", rawTitle, rawEventID, rawDate, rawRemaining)
		}
	}

	remaining, err := strconv.Atoi(rawRemaining)
	// If this fails there is a string message in place of a number. We can assume class is full.
	if err != nil {
		remaining = 0
	}

	// Because I know somebody will have questions eventually:
	// https://stackoverflow.com/questions/14106541/parsing-date-time-strings-which-are-not-standard-formats/14106561
	const form = "1/2/2006"
	loc, _ := time.LoadLocation("America/Denver")
	date, err := time.ParseInLocation(form, rawDate, loc)
	if err != nil {
		return CalendarEntry{}, err
	}
	
	return CalendarEntry{
		Title: rawTitle,
		EventID: strings.TrimPrefix(rawEventID, eventBase),
		Date: date,
		Remaining: remaining,
	}, nil
}

// Retreives calendar metadata from Colorado Mountain Club website.
func ScrapeCalendarEntries() (CalendarEntries, error) {
	var entries CalendarEntries
	idToTripType := make(map[string]string)
	
	// ASP.NET pages are notoriously difficult to scrape. We visit the calendar page twice so that we can
	// collect "hidden" form values first before requesting the data we really want. After obtaining
	// the event ids we can move on to collecting event type by drilling down on each event.
	firstVisit := colly.NewCollector()

	var viewState, generator, eventValidation string
	firstVisit.OnHTML("body", func(e *colly.HTMLElement) {
		want := "value"
		viewState = e.ChildAttr("#__VIEWSTATE", want)
		generator = e.ChildAttr("#__VIEWSTATEGENERATOR", want)
		eventValidation = e.ChildAttr("#__EVENTVALIDATION", want)
	})

	var failed error
	firstVisit.OnError(func(r *colly.Response, err error) {
		failed = fmt.Errorf("First visit request:", r.Request.URL, "failed with response:", r, "\nError:", err)
	})
	firstVisit.Visit(calendarURL)
	if failed != nil {
		return entries, failed
	}

	
	secondVisit := colly.NewCollector(
		colly.MaxDepth(2),
	)
	// CMC doesn't have very high tolerance for scraping.
	secondVisit.Limit(&colly.LimitRule{
		RandomDelay: 3 * time.Second,
	})
	
	secondVisit.OnHTML("#dnn_ctr907_EventList_gridList", func(e *colly.HTMLElement) {
		e.ForEach("tbody tr", func(_ int, row *colly.HTMLElement) {
			entry, err := calendarEntryFrom(row)
			if err != nil {
				log.Printf("Failed to parse calendar entry: %+v", row)
				return
			}
			entries = append(entries, entry)

			e.Request.Visit(row.ChildAttr("td a", "href"))
		})
	})
	secondVisit.OnHTML("tr td:nth-child(2)", func(e *colly.HTMLElement) {
		wantElement := "span"
		if strings.Contains(e.ChildAttr(wantElement, "id"), "lblType") {
			idToTripType[strings.TrimPrefix(e.Request.URL.RawQuery, "ID=")] = e.ChildText(wantElement)
		}
	})

	secondVisit.OnError(func(r *colly.Response, err error) {
		failed = fmt.Errorf("Second visit request:", r.Request.URL, "failed with response:", r, "\nError:", err)
	})
	secondVisit.PostMultipart(calendarURL, buildMultipartFormWithViewState(viewState, generator, eventValidation))

	for i := range entries {
		if tripType, ok := idToTripType[entries[i].EventID]; ok {
			entries[i].TripType = tripType
		}
	}

	return entries, failed
}
