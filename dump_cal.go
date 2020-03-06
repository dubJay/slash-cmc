package main

import (
	"fmt"
	"io/ioutil"

	"github.com/dubJay/slash-cmc/scraper"
)

func main() {
	// ScrapeCalendar takes about a minute to execute so that we don't trigger any anti-scraping actions.
	entries, err := scraper.ScrapeCalendarEntries()
	if err != nil {
		panic(err)
	}

	fmt.Println("Total Parsed Entries %d", len(entries))
	ics := scraper.CalendarEntriesToICS(entries)
	err = ioutil.WriteFile("./cmc-cal.ics", []byte(ics), 0644)
	if err != nil {
		panic(err)
	}
}
