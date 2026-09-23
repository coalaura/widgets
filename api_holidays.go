package main

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"
)

type Holiday struct {
	Date             string   `json:"date"`
	Name             string   `json:"name"`
	NationalHoliday  bool     `json:"nationalHoliday"`
	SubdivisionCodes []string `json:"subdivisionCodes"`
	HolidayTypes     []string `json:"holidayTypes"`
}

type HolidayDisplay struct {
	Date string `json:"date"`
	Name string `json:"name"`
}

var holidays = new(DataCache[[]Holiday])

func getHolidays(country, region string) []HolidayDisplay {
	country = strings.ToUpper(strings.TrimSpace(country))
	region = strings.ToUpper(strings.TrimSpace(region))

	if len(country) != 2 || country[0] < 'A' || country[0] > 'Z' || country[1] < 'A' || country[1] > 'Z' {
		return nil
	}

	year := time.Now().UTC().Year()
	result := make([]HolidayDisplay, 0, 32)

	for currentYear := year; currentYear <= year+1; currentYear++ {
		key := country + ":" + strconv.Itoa(currentYear)

		list, ok := holidays.get(key, locationCacheTTL, fetchHolidays)
		if !ok {
			continue
		}

		for _, entry := range list {
			if !entry.NationalHoliday && !slices.Contains(entry.SubdivisionCodes, region) {
				continue
			}

			if !slices.Contains(entry.HolidayTypes, "Public") {
				continue
			}

			result = append(result, HolidayDisplay{Date: entry.Date, Name: entry.Name})
		}
	}

	return result
}

func fetchHolidays(key string) ([]Holiday, error) {
	country, year, ok := strings.Cut(key, ":")
	if !ok {
		return nil, fmt.Errorf("invalid holiday cache key")
	}

	var result []Holiday

	endpoint := "https://nagerholidays.com/api/v4/Holidays/" + country + "/" + year

	err := fetchJSON(endpoint, &result)
	if err != nil {
		return nil, err
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("no holidays returned for %s", key)
	}

	return result, nil
}
