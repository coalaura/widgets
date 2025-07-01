package main

import (
	"encoding/json"
	"math/rand"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
)

type Handler func(*fiber.Ctx, map[string]any)

type Widget struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Options     Options `json:"options"`
	Handler     Handler `json:"-"`
}

type WidgetManager struct {
	sync.RWMutex
	json    []byte
	widgets map[string]*Widget
}

func NewWidget(name, description string, options Options, handler Handler) *Widget {
	if options == nil {
		options = make(Options)
	}

	options["font"] = NewString("JetBrains Mono", "The font family for the widget text. Accepts any valid CSS font-family value.")
	options["size"] = NewSize("100%", "The font size of the widget text. Accepts any valid CSS font-size value.")
	options["color"] = NewColor("#cad3f5", "The color of the widget text. Accepts any valid CSS color value.")
	options["weight"] = NewEnum("700", slice(100, 200, 300, 400, 500, 600, 700, 800, 900), "The font weight of the widget text. Accepts CSS font-weight numeric values (100-900).")
	options["align"] = NewEnum("left", slice("left", "center", "right"), "The horizontal alignment of the widget text. Accepts 'left', 'center' or 'right'.")

	return &Widget{
		Name:        name,
		Description: description,
		Options:     options,
		Handler:     handler,
	}
}

func NewWidgetManager() *WidgetManager {
	return &WidgetManager{
		widgets: make(map[string]*Widget),
	}
}

func (w *Widget) Render(c *fiber.Ctx) error {
	opts := make(map[string]any)

	for key, opt := range w.Options {
		opts[key] = opt.Value(c.Query(key))
	}

	if w.Handler != nil {
		w.Handler(c, opts)
	}

	return c.Render("widgets/"+w.Name, opts, "layout")
}

func (m *WidgetManager) Register(name, description string, options Options, handler Handler) {
	m.Lock()
	defer m.Unlock()

	m.widgets[name] = NewWidget(name, description, options, handler)
}

func (m *WidgetManager) Get(name string) *Widget {
	m.RLock()
	defer m.RUnlock()

	if wg, ok := m.widgets[name]; ok {
		return wg
	}

	return nil
}

func (m *WidgetManager) JSON() []byte {
	m.RLock()
	defer m.RUnlock()

	if m.json == nil {
		var list []*Widget

		for _, widget := range m.widgets {
			list = append(list, widget)
		}

		sort.Slice(list, func(a, b int) bool {
			return list[a].Name < list[b].Name
		})

		m.json, _ = json.Marshal(list)
	}

	return m.json
}

func (m *WidgetManager) Render(c *fiber.Ctx, name string) error {
	wg := m.Get(name)
	if wg == nil {
		return abort(c, 404)
	}

	return wg.Render(c)
}

func (m *WidgetManager) RegisterDefault() {
	now := time.Now()
	ago := now.Add(-time.Duration(rand.Intn(12)+18) * 365 * 24 * time.Hour)

	yearNow := now.Format("2006")
	yearAgo := ago.Format("2006")

	// Display current IP
	m.Register(
		"ip",
		"Shows your current IP address, optionally prefixed with custom text. Useful for network status or location info.",
		Options{
			"prefix": NewString("IP is ", "A string to display before the IP address. Can be left empty."),
		},
		func(c *fiber.Ctx, options map[string]any) {
			ip := c.IP()

			if ip == "" {
				ip = "n/a"
			}

			options["ip"] = ip
		},
	)

	// Display a date progress
	m.Register(
		"progress",
		"Calculates and displays the elapsed time between two dates as a percentage, giving a visual sense of completion.",
		Options{
			"from": NewString("01-01-"+yearNow, "The start date of the progress period. Accepts any valid date/time format."),
			"to":   NewString("12-31-"+yearNow, "The end date of the progress period. Accepts any valid date/time format."),
			"bg":   NewColor("#b7bdf8", "The background color for the progress bar. Accepts any valid CSS color value."),
		},
		nil,
	)

	// Display the date and/or time
	m.Register(
		"date",
		"Displays the current date/time in a specified format (using dayjs). Handy for quick references or scheduling.",
		Options{
			"format": NewString("dddd, MMMM D, h:mm A", "The display format for the date/time, using dayjs.js tokens. Accepts any valid date/time format supported by day.js."),
		},
		nil,
	)

	// Display your precise age
	m.Register(
		"age",
		"Displays your current age as a floating-point number. For extra precision, you can add your time of birth.",
		Options{
			"date": NewString("01-01-"+yearAgo, "Your date of birth (optionally include time of birth for extra precision). Accepts any valid date/time format."),
		},
		nil,
	)

	// Countdown to a specific event
	m.Register(
		"countdown",
		"Displays a countdown to a specific date and time.",
		Options{
			"event": NewString("A cool thing", "The name of the event being counted down to."),
			"to":    NewString("2026-02-10T09:00:00", "The target date of the event. Accepts any valid date/time format."),
		},
		nil,
	)

	// Display a binary clock
	m.Register(
		"binary",
		"A clock for those who think in 0s and 1s.",
		Options{
			"rule": NewString(" : ", "The character(s) used to separate the binary hours, minutes, and seconds."),
		},
		nil,
	)

	// Currency conversion widget
	m.Register(
		"currency",
		"Converts an amount from one currency to another using up-to-date exchange rates.",
		Options{
			"from":   NewEnum("EUR", currencies.Enum(), "The currency code to convert from. Accepts any valid (and supported) 3 letter currency code."),
			"to":     NewEnum("USD", currencies.Enum(), "The currency code to convert to. Accepts any valid (and supported) 3 letter currency code."),
			"round":  NewInt(3, "The maximum decimal precision of the conversion."),
			"amount": NewFloat(1.0, "The amount of the 'from' currency to convert."),
			"format": NewString("{amount} {from} = {rate} {to}", "How to format the resulting conversion rate."),
		},
		func(c *fiber.Ctx, options map[string]any) {
			from := options["from"].(string)
			to := options["to"].(string)

			round := options["round"].(int)
			amount := options["amount"].(float64)

			rate := currencies.CalculateRate(from, to) * amount

			options["rate"] = strconv.FormatFloat(rate, 'f', round, 64)
		},
	)
}
