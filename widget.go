package main

import (
	"encoding/json"
	"math/rand"
	"sort"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
)

type Opts map[string]*Opt

type Handler func(*fiber.Ctx, map[string]string)

type Opt struct {
	Default     string `json:"default"`
	Description string `json:"description"`
}

type Widget struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Options     Opts    `json:"options"`
	Handler     Handler `json:"-"`
}

type WidgetManager struct {
	sync.RWMutex
	json    []byte
	widgets map[string]*Widget
}

func NewOpt(def, description string) *Opt {
	return &Opt{
		Default:     def,
		Description: description,
	}
}

func NewWidget(name, description string, options Opts, handler Handler) *Widget {
	if options == nil {
		options = make(Opts)
	}

	options.RegisterDefault("font", "JetBrains Mono", "The font family for the widget text. Accepts any valid CSS font-family value.")
	options.RegisterDefault("size", "100%", "The font size of the widget text. Accepts any valid CSS font-size value.")
	options.RegisterDefault("color", "#cad3f5", "The color of the widget text. Accepts any valid CSS color value.")
	options.RegisterDefault("weight", "700", "The font weight of the widget text. Accepts CSS font-weight numeric values (100-900).")
	options.RegisterDefault("align", "left", "The horizontal alignment of the widget text. Accepts 'left', 'center' or 'right'.")

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

func (o *Opts) RegisterDefault(name, def, description string) {
	if _, ok := (*o)[name]; ok {
		return
	}

	(*o)[name] = NewOpt(def, description)
}

func (w *Widget) Render(c *fiber.Ctx) error {
	opts := make(map[string]string)

	for key, opt := range w.Options {
		value := c.Query(key)

		if value == "" {
			value = opt.Default
		}

		opts[key] = value
	}

	if w.Handler != nil {
		w.Handler(c, opts)
	}

	return c.Render("widgets/"+w.Name, opts, "layout")
}

func (m *WidgetManager) Register(name, description string, options Opts, handler Handler) {
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
		Opts{
			"prefix": NewOpt("IP is ", "A string to display before the IP address. Can be left empty."),
		},
		func(c *fiber.Ctx, options map[string]string) {
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
		Opts{
			"from": NewOpt("01-01-"+yearNow, "The start date of the progress period. Format: DD-MM-YYYY."),
			"to":   NewOpt("12-31-"+yearNow, "The end date of the progress period. Format: DD-MM-YYYY."),
			"bg":   NewOpt("#b7bdf8", "The background color for the progress bar. Accepts any valid CSS color value."),
		},
		nil,
	)

	// Display the date and/or time
	m.Register(
		"date",
		"Displays the current date/time in a specified format (using dayjs). Handy for quick references or scheduling.",
		Opts{
			"format": NewOpt("DD/MM/YYYY", "The display format for the date/time, using dayjs.js tokens. Example: 'dddd, MMMM D, h:mm A'."),
		},
		nil,
	)

	// Display your precise age
	m.Register(
		"age",
		"Displays your current age as a floating-point number. For extra precision, you can add your time of birth.",
		Opts{
			"birthday": NewOpt("01-01-"+yearAgo, "Your date of birth. Required. Format: YYYY-MM-DD."),
			"time":     NewOpt("", "Your time of birth for extra precision. Optional. Format: HH:mm:ss."),
		},
		nil,
	)

	// Countdown to a specific event
	m.Register(
		"countdown",
		"Displays a countdown to a specific date and time.",
		Opts{
			"event": NewOpt("A cool thing", "The name of the event being counted down to."),
			"to":    NewOpt("2026-02-10T09:00:00", "The target date and time in ISO 8601 format (YYYY-MM-DDTHH:mm:ss)."),
		},
		nil,
	)

	// Display a binary clock
	m.Register(
		"binary",
		"A clock for those who think in 0s and 1s.",
		Opts{
			"rule": NewOpt(" : ", "The character(s) used to separate the binary hours, minutes, and seconds."),
		},
		nil,
	)
}
