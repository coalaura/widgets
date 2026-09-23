package main

import (
	"net/http"
	"time"
)

const (
	apiCacheTTL   = time.Hour
	apiRetryDelay = time.Minute
)

var apiClient = &http.Client{Timeout: 5 * time.Second}
