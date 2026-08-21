package helpers

import (
	"time"
	"errors"
)

// ParseDate parses a date string in the format "2006-01-02" or RFC3339 and returns a time.Time object.
func ParseDate(dateStr string) (*time.Time, error) {
	if dateStr != "" {
		parsedDate, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			parsedDate, err = time.Parse(time.RFC3339, dateStr)
		}
		if err != nil {
			return nil, err
		}
		return &parsedDate, nil
	} else {
		return nil, errors.New("date string is empty")
	}
}