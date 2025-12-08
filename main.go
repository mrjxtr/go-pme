package main

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sync"
	"time"
)

type Endpoint struct {
	URL    string
	Method string
	Params map[string]string
}

func (ep *Endpoint) poke() error {
	var resp *http.Response

	switch ep.Method {
	case "GET":
		r, err := http.Get(ep.URL)
		if err != nil {
			return fmt.Errorf("GET request error for %s: %w", ep.URL, err)
		}
		resp = r
	case "POST":
		// NOTE: Nil for now for quick testing
		var body io.Reader = nil

		r, err := http.Post(ep.URL, "application/json", body)
		if err != nil {
			return fmt.Errorf("POST request error for %s: %w", ep.URL, err)
		}
		resp = r
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status %d for %s", resp.StatusCode, ep.URL)
	}

	slog.Info("Succesfully poked", "endpoint", ep.URL)
	return nil
}

func main() {
	endpoints := []Endpoint{
		{
			URL:    "https://randompinoy.xyz/api/v1/pinoys",
			Method: "GET",
		},
		{
			URL:    "https://randompinoy.xyz/api/v1/pinoys?results=3",
			Method: "GET",
		},
		{
			URL:    "https://randompinoy.xyz/api/v1/pinoys?results=31",
			Method: "GET",
		},
	}

	var wg sync.WaitGroup

	slog.Info("Poking endpoints", "count", len(endpoints))

	currTime := time.Now()

	for _, endpoint := range endpoints {
		ep := endpoint
		wg.Go(func() {
			if err := ep.poke(); err != nil {
				slog.Error("poke failed", "error", err)
			}
		})
	}

	wg.Wait()

	slog.Info("Time elapsed", "time", time.Since(currTime))
	slog.Info("Poked endpoints", "count", len(endpoints))
}
