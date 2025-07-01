package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

type NasaApod struct {
	Date  string `json:"date"`
	HdUrl string `json:"hdurl"`
	Title string `json:"title"`
}

type NasaStore struct {
	sync.RWMutex

	APOD NasaApod
}

var nasa = NewNasaStore()

func init() {
	nasa.UpdateAPOD()

	ticker := time.NewTicker(1 * time.Hour)

	go func() {
		defer ticker.Stop()

		for range ticker.C {
			nasa.UpdateAPOD()
		}
	}()
}

func NewNasaStore() *NasaStore {
	return &NasaStore{}
}

func (n *NasaStore) GetAPOD() NasaApod {
	n.RLock()
	defer n.RUnlock()

	return n.APOD
}

func (n *NasaStore) UpdateAPOD() {
	resp, err := http.Get(fmt.Sprintf("https://api.nasa.gov/planetary/apod?api_key=%s", NasaAPIKey))
	if err != nil {
		log.Warningf("unable to query api.nasa.gov: %v\n", err)

		return
	}

	defer resp.Body.Close()

	var result NasaApod

	if err = json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Warningf("unable to decode api.nasa.gov: %v\n", err)

		return
	}

	n.Lock()
	defer n.Unlock()

	n.APOD = result
}
