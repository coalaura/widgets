package main

import (
	"encoding/json"
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

	APOD    NasaApod
	expires time.Time
}

var nasa = NewNasaStore()

func NewNasaStore() *NasaStore {
	return &NasaStore{}
}

func (n *NasaStore) GetAPOD() NasaApod {
	n.RLock()

	if time.Now().Before(n.expires) {
		apod := n.APOD
		n.RUnlock()

		return apod
	}

	n.RUnlock()

	n.Lock()
	defer n.Unlock()

	if !time.Now().Before(n.expires) {
		if n.updateAPOD() {
			n.expires = time.Now().Add(apiCacheTTL)
		} else {
			n.expires = time.Now().Add(apiRetryDelay)
		}
	}

	return n.APOD
}

// updateAPOD runs while n is locked and preserves the previous image on failure.
func (n *NasaStore) updateAPOD() bool {
	resp, err := apiClient.Get("https://api.nasa.gov/planetary/apod?api_key=" + NasaAPIKey)
	if err != nil {
		log.Warnf("unable to query api.nasa.gov: %v\n", err)

		return false
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Warnf("api.nasa.gov returned %s\n", resp.Status)

		return false
	}

	var result NasaApod

	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		log.Warnf("unable to decode api.nasa.gov: %v\n", err)

		return false
	}

	if result.HdUrl == "" {
		log.Warnf("api.nasa.gov returned no image URL\n")

		return false
	}

	n.APOD = result

	return true
}
