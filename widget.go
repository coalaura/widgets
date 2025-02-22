package main

import (
	"encoding/json"
	"sort"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/exp/maps"
)

type Opts map[string]string

type Handler func(*fiber.Ctx, Opts)

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

func NewWidget(name, description string, options Opts, handler Handler) *Widget {
	if options == nil {
		options = make(Opts)
	}

	options["font"] = "Roboto"
	options["size"] = "100%"
	options["color"] = "#cad3f5"
	options["weight"] = "700"
	options["align"] = "left"

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
	opts := maps.Clone(w.Options)

	for key, def := range opts {
		opts[key] = c.Query(key, def)
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
	year := now.Format("2006")

	// Display current IP
	m.Register(
		"ip",
		"Displays the current date/time in a specified format (using dayjs). Handy for quick references or scheduling.",
		Opts{
			"prefix": "IP is ",
		},
		func(c *fiber.Ctx, options Opts) {
			options["ip"] = c.IP()
		},
	)

	// Display a date progress
	m.Register(
		"progress",
		"Shows your current IP address, optionally prefixed with custom text. Useful for network status or location info.",
		Opts{
			"from": "01-01-" + year,
			"to":   "12-31-" + year,
			"bg":   "#b7bdf8",
		},
		nil,
	)

	// Display the date and/or time
	m.Register(
		"date",
		"Calculates and displays the elapsed time between two dates as a percentage, giving a visual sense of completion.",
		Opts{
			"format": "DD/MM/YYYY",
		},
		nil,
	)
}
