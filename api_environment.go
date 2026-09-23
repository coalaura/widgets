package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	locationCacheTTL = 24 * time.Hour
	forecastCacheTTL = 15 * time.Minute
	airCacheTTL      = 30 * time.Minute
)

type CacheEntry[T any] struct {
	sync.Mutex
	value   T
	expires time.Time
	ready   bool
}

type DataCache[T any] struct {
	sync.Mutex
	entries map[string]*CacheEntry[T]
}

type Location struct {
	Name      string  `json:"name"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type GeocodingResult struct {
	Results []Location `json:"results"`
}

type CurrentWeather struct {
	Temperature *float64 `json:"temperature_2m"`
	WeatherCode *int     `json:"weather_code"`
}

type DailySun struct {
	Sunrise          []string  `json:"sunrise"`
	Sunset           []string  `json:"sunset"`
	DaylightDuration []float64 `json:"daylight_duration"`
}

type Forecast struct {
	Location Location       `json:"-"`
	Timezone string         `json:"timezone"`
	Current  CurrentWeather `json:"current"`
	Daily    DailySun       `json:"daily"`
}

type CurrentAir struct {
	AQI *int `json:"us_aqi"`
}

type AirConditions struct {
	Location Location   `json:"-"`
	Current  CurrentAir `json:"current"`
}

var (
	locations = new(DataCache[Location])
	forecasts = new(DataCache[Forecast])
	airData   = new(DataCache[AirConditions])
)

func (c *DataCache[T]) get(key string, ttl time.Duration, fetch func(string) (T, error)) (T, bool) {
	c.Lock()

	if c.entries == nil {
		c.entries = make(map[string]*CacheEntry[T])
	}

	entry := c.entries[key]
	if entry == nil {
		// Bound the number of distinct user-supplied locations kept in memory.
		if len(c.entries) >= 128 {
			for oldKey := range c.entries {
				delete(c.entries, oldKey)

				break
			}
		}

		entry = new(CacheEntry[T])
		c.entries[key] = entry
	}

	c.Unlock()

	entry.Lock()
	defer entry.Unlock()

	if time.Now().Before(entry.expires) {
		return entry.value, entry.ready
	}

	value, err := fetch(key)
	if err != nil {
		log.Warnf("unable to refresh widget data: %v\n", err)
		entry.expires = time.Now().Add(apiRetryDelay)

		return entry.value, entry.ready
	}

	entry.value = value
	entry.ready = true
	entry.expires = time.Now().Add(ttl)

	return entry.value, true
}

func getForecast(city string) (Forecast, bool) {
	return forecasts.get(normalizeCity(city), forecastCacheTTL, fetchForecast)
}

func getAir(city string) (AirConditions, bool) {
	return airData.get(normalizeCity(city), airCacheTTL, fetchAir)
}

func fetchLocation(city string) (Location, error) {
	var result GeocodingResult

	endpoint := "https://geocoding-api.open-meteo.com/v1/search?count=1&language=en&name=" + url.QueryEscape(city)

	err := fetchJSON(endpoint, &result)
	if err != nil {
		return Location{}, err
	}

	if len(result.Results) == 0 {
		return Location{}, fmt.Errorf("no location found for %q", city)
	}

	return result.Results[0], nil
}

func fetchForecast(city string) (Forecast, error) {
	place, ok := locations.get(city, locationCacheTTL, fetchLocation)
	if !ok {
		return Forecast{}, fmt.Errorf("location unavailable for %q", city)
	}

	var result Forecast

	endpoint := "https://api.open-meteo.com/v1/forecast?latitude=" + strconv.FormatFloat(place.Latitude, 'f', -1, 64) + "&longitude=" + strconv.FormatFloat(place.Longitude, 'f', -1, 64) + "&current=temperature_2m,weather_code&daily=sunrise,sunset,daylight_duration&past_days=1&forecast_days=2&timezone=auto"

	err := fetchJSON(endpoint, &result)
	if err != nil {
		return Forecast{}, err
	}

	if result.Current.Temperature == nil || result.Current.WeatherCode == nil || result.Timezone == "" ||
		len(result.Daily.Sunrise) < 3 || len(result.Daily.Sunset) < 3 || len(result.Daily.DaylightDuration) < 3 {
		return Forecast{}, fmt.Errorf("incomplete forecast for %q", city)
	}

	if !strings.Contains(result.Daily.Sunrise[1], "T") || !strings.Contains(result.Daily.Sunset[1], "T") {
		return Forecast{}, fmt.Errorf("invalid sunrise or sunset for %q", city)
	}

	result.Location = place

	return result, nil
}

func fetchAir(city string) (AirConditions, error) {
	place, ok := locations.get(city, locationCacheTTL, fetchLocation)
	if !ok {
		return AirConditions{}, fmt.Errorf("location unavailable for %q", city)
	}

	var result AirConditions

	endpoint := "https://air-quality-api.open-meteo.com/v1/air-quality?latitude=" + strconv.FormatFloat(place.Latitude, 'f', -1, 64) + "&longitude=" + strconv.FormatFloat(place.Longitude, 'f', -1, 64) + "&current=us_aqi&timezone=auto"

	err := fetchJSON(endpoint, &result)
	if err != nil {
		return AirConditions{}, err
	}

	if result.Current.AQI == nil || *result.Current.AQI < 0 {
		return AirConditions{}, fmt.Errorf("missing air quality for %q", city)
	}

	result.Location = place

	return result, nil
}

func fetchJSON(endpoint string, destination any) error {
	response, err := apiClient.Get(endpoint)
	if err != nil {
		return err
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("%s returned %s", response.Request.URL.Host, response.Status)
	}

	err = json.NewDecoder(io.LimitReader(response.Body, 1<<20)).Decode(destination)
	if err != nil {
		return fmt.Errorf("unable to decode %s: %w", response.Request.URL.Host, err)
	}

	return nil
}

func normalizeCity(city string) string {
	city = strings.ToLower(strings.TrimSpace(city))

	if len(city) > 100 {
		city = city[:100]
	}

	return city
}
