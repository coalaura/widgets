package main

import (
	"encoding/json"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"
)

type CurrencyStore struct {
	sync.RWMutex

	Amount float64            `json:"amount"`
	Base   string             `json:"base"`
	Rates  map[string]float64 `json:"rates"`
}

var currencies = NewCurrencyStore()

func init() {
	currencies.Update()

	ticker := time.NewTicker(30 * time.Minute)

	go func() {
		defer ticker.Stop()

		for range ticker.C {
			currencies.Update()
		}
	}()
}

func NewCurrencyStore() *CurrencyStore {
	return &CurrencyStore{
		Rates: make(map[string]float64),
	}
}

func (c *CurrencyStore) Enum() []string {
	enum := []string{c.Base}

	c.RLock()

	for currency := range c.Rates {
		enum = append(enum, currency)
	}

	c.RUnlock()

	sort.Strings(enum)

	return enum
}

func (c *CurrencyStore) CalculateRate(from, to string) float64 {
	from = strings.ToUpper(from)
	to = strings.ToUpper(to)

	if from == to {
		return 1.0
	}

	c.RLock()
	defer c.RUnlock()

	if from == c.Base {
		rate, ok := c.Rates[to]
		if !ok {
			return 0.0
		}

		return rate
	}

	if to == c.Base {
		rate, ok := c.Rates[from]
		if !ok {
			return 0.0
		}

		return 1.0 / rate
	}

	fromRate, fromOk := c.Rates[from]
	toRate, toOk := c.Rates[to]

	if !fromOk || !toOk {
		return 0.0
	}

	return toRate / fromRate
}

func (c *CurrencyStore) Update() {
	resp, err := http.Get("https://api.frankfurter.dev/v1/latest")
	if err != nil {
		log.Warningf("unable to query frankfurter.dev: %v\n", err)

		return
	}

	defer resp.Body.Close()

	var result CurrencyStore

	if err = json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Warningf("unable to decode frankfurter.dev: %v\n", err)

		return
	}

	c.Lock()
	defer c.Unlock()

	c.Amount = result.Amount
	c.Base = result.Base
	c.Rates = result.Rates
}
