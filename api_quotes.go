package main

import (
	"fmt"
	"strings"
)

type Quote struct {
	Text   string `json:"quote"`
	Author string `json:"author"`
}

var quotes = new(DataCache[Quote])

func fetchQuote(_ string) (Quote, error) {
	var result Quote

	err := fetchJSON("https://quote.shrt.day/api/quote", &result)
	if err != nil {
		return Quote{}, err
	}

	result.Text = strings.TrimSpace(result.Text)
	result.Author = strings.TrimSpace(result.Author)

	if result.Text == "" {
		return Quote{}, fmt.Errorf("quote.shrt.day returned no quote")
	}

	return result, nil
}
