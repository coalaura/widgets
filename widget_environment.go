package main

import (
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func (m *WidgetManager) RegisterEnvironment() {
	m.Register(
		"moon",
		"The current moon phase and approximate illumination, updated throughout the day.",
		nil,
		nil,
	)

	m.Register(
		"sun",
		"Today's sunrise and sunset times for a city.",
		Options{
			"city":   cityOption(),
			"clock":  NewEnum("24", slice("24", "12"), "Show times using a 24-hour or 12-hour clock."),
			"format": NewString("{sunrise} → {sunset}", "Text to display. Use {sunrise} and {sunset} for the local times."),
		},
		func(_ http.ResponseWriter, _ *http.Request, options map[string]any) {
			data, ok := getForecast(options["city"].(string))
			options["available"] = ok

			if ok {
				_, sunrise, _ := strings.Cut(data.Daily.Sunrise[1], "T")
				_, sunset, _ := strings.Cut(data.Daily.Sunset[1], "T")

				clock := options["clock"].(string)
				sunrise = formatSolarTime(sunrise, clock)
				sunset = formatSolarTime(sunset, clock)

				text := strings.ReplaceAll(options["format"].(string), "{sunrise}", sunrise)
				options["times"] = strings.ReplaceAll(text, "{sunset}", sunset)
				options["place"] = data.Location.Name
			}
		},
	)

	m.Register(
		"weather",
		"Current temperature and conditions in a city.",
		Options{
			"city": cityOption(),
			"unit": NewEnum("C", slice("C", "F"), "Temperature in Celsius or Fahrenheit."),
		},
		func(_ http.ResponseWriter, _ *http.Request, options map[string]any) {
			data, ok := getForecast(options["city"].(string))
			options["available"] = ok

			if ok {
				temperature := *data.Current.Temperature

				if options["unit"] == "F" {
					temperature = temperature*9/5 + 32
				}

				options["temperature"] = strconv.FormatFloat(temperature, 'f', 0, 64)
				options["conditions"] = weatherDescription(*data.Current.WeatherCode)
				options["place"] = data.Location.Name
			}
		},
	)

	m.Register(
		"holiday",
		"The next public holiday in a country, optionally including a local region.",
		Options{
			"country":     NewString("DE", "Two-letter country code (for example DE, US or GB)."),
			"region":      NewString("", "Optional subdivision code for regional public holidays (for example DE-BE or US-CA)."),
			"display":     NewEnum("relative", slice("relative", "date", "both"), "Show days remaining, the calendar date, or both."),
			"date_format": NewString("MMM Do", "Date pattern, for example MMM Do, DD.MM.YYYY or dddd, MMMM D. Wrap literal text in [brackets]."),
		},
		func(_ http.ResponseWriter, _ *http.Request, options map[string]any) {
			options["holidays"] = getHolidays(options["country"].(string), options["region"].(string))
		},
	)

	m.Register(
		"week",
		"The current ISO week number and how far away the weekend is.",
		nil,
		nil,
	)

	m.Register(
		"daylight",
		"Compare today's hours of daylight with yesterday's for a city.",
		cityOptions(),
		func(_ http.ResponseWriter, _ *http.Request, options map[string]any) {
			data, ok := getForecast(options["city"].(string))
			options["available"] = ok

			if ok {
				options["minutes"] = int(math.Round((data.Daily.DaylightDuration[1] - data.Daily.DaylightDuration[0]) / 60))
				options["hours"] = int(data.Daily.DaylightDuration[1] / 3600)
				options["remainder"] = int(data.Daily.DaylightDuration[1]/60) % 60
				options["place"] = data.Location.Name
			}
		},
	)

	m.Register(
		"air",
		"Current US air quality index and its health category for a city.",
		cityOptions(),
		func(_ http.ResponseWriter, _ *http.Request, options map[string]any) {
			data, ok := getAir(options["city"].(string))
			options["available"] = ok

			if ok {
				options["aqi"] = *data.Current.AQI
				options["category"] = airQualityCategory(*data.Current.AQI)
				options["place"] = data.Location.Name
			}
		},
	)

	m.Register(
		"curiosity",
		"A small science or nature fact that changes every day.",
		nil,
		nil,
	)
}

func cityOptions() Options {
	return Options{"city": cityOption()}
}

func cityOption() Option {
	return NewString("Berlin", "City or postal code. Add a country after a comma to narrow the search.")
}

func formatSolarTime(value, clock string) string {
	if clock == "24" {
		return value
	}

	parsed, err := time.Parse("15:04", value)
	if err != nil {
		return value
	}

	return parsed.Format("3:04 PM")
}

func weatherDescription(code int) string {
	switch code {
	case 0:
		return "Clear sky"
	case 1, 2:
		return "Partly cloudy"
	case 3:
		return "Overcast"
	case 45, 48:
		return "Fog"
	case 51, 53, 55, 56, 57:
		return "Drizzle"
	case 61, 63, 65, 66, 67, 80, 81, 82:
		return "Rain"
	case 71, 73, 75, 77, 85, 86:
		return "Snow"
	case 95, 96, 99:
		return "Thunderstorms"
	default:
		return "Conditions unavailable"
	}
}

func airQualityCategory(aqi int) string {
	switch {
	case aqi <= 50:
		return "Good"
	case aqi <= 100:
		return "Moderate"
	case aqi <= 150:
		return "Sensitive groups"
	case aqi <= 200:
		return "Unhealthy"
	case aqi <= 300:
		return "Very unhealthy"
	default:
		return "Hazardous"
	}
}
