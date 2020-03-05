package main

import (
	"fmt"

	"github.com/dubJay/slash-cmc/scraper"
)

func main() {
	// ScrapeCalendar takes about a minute to execute so that we don't trigger any anti-scraping actions.
	entries, err := scraper.ScrapeCalendarEntries()
	if err != nil {
		fmt.Println("Failed to scrape CMC calendar:", err)
		return
	}

	fmt.Println("Total Parsed Entries %d", len(entries))
	for _, entry := range entries {
		fmt.Printf("%+v\n\n", entry)
	}
	
}
