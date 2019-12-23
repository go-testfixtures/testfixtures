package testfixtures

import (
	"fmt"
	"time"
)

var timeFormats = []string{
	"2006-01-02",
	"2006-01-02 15:04",
	"2006-01-02 15:04:05",
	"20060102",
	"20060102 15:04",
	"20060102 15:04:05",
	"02/01/2006",
	"02/01/2006 15:04",
	"02/01/2006 15:04:05",
	"2006-01-02T15:04-07:00",
	"2006-01-02T15:04:05-07:00",
}

func tryStrToDate(s string) (time.Time, error) {
	for _, f := range timeFormats {
		t, err := time.ParseInLocation(f, s, time.Local)
		if err != nil {
			continue
		}
		return t, nil
	}
	return time.Time{}, fmt.Errorf(`testfixtures: could not convert string "%s" to time`, s)
}
