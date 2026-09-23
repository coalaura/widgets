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
	expires time.Time

	Amount float64            `json:"amount"`
	Base   string             `json:"base"`
	Rates  map[string]float64 `json:"rates"`
}

var currencies = NewCurrencyStore()

func NewCurrencyStore() *CurrencyStore {
	return &CurrencyStore{
		Rates: make(map[string]float64),
	}
}

func (c *CurrencyStore) Enum() []string {
	c.ensureFresh()

	c.RLock()
	defer c.RUnlock()

	if c.Base == "" {
		return []string{"EUR", "USD"}
	}

	enum := make([]string, 0, len(c.Rates)+1)

	enum = append(enum, c.Base)

	for currency := range c.Rates {
		enum = append(enum, currency)
	}

	sort.Strings(enum)

	return enum
}

func (c *CurrencyStore) CalculateRate(from, to string) float64 {
	c.ensureFresh()

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

func (c *CurrencyStore) ensureFresh() {
	c.RLock()
	fresh := time.Now().Before(c.expires)
	c.RUnlock()

	if fresh {
		return
	}

	c.Lock()
	defer c.Unlock()

	if time.Now().Before(c.expires) {
		return
	}

	if c.update() {
		c.expires = time.Now().Add(apiCacheTTL)
	} else {
		c.expires = time.Now().Add(apiRetryDelay)
	}
}

// update runs while c is locked and preserves the previous rates on failure.
func (c *CurrencyStore) update() bool {
	resp, err := apiClient.Get("https://api.frankfurter.dev/v1/latest")
	if err != nil {
		log.Warnf("unable to query frankfurter.dev: %v\n", err)

		return false
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Warnf("frankfurter.dev returned %s\n", resp.Status)

		return false
	}

	var result CurrencyStore

	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		log.Warnf("unable to decode frankfurter.dev: %v\n", err)

		return false
	}

	if result.Base == "" || len(result.Rates) == 0 {
		log.Warnf("frankfurter.dev returned incomplete rates\n")

		return false
	}

	c.Amount = result.Amount
	c.Base = result.Base
	c.Rates = result.Rates

	return true
}
