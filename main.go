package main

import (
	"io"
	"log/slog"
	"net/http"
	"os"
	"sync"
	"time"
)

type Endpoint struct {
	URL    string
	Method string
	Params map[string]string
}

func (ep *Endpoint) poke() {
	var resp *http.Response

	switch ep.Method {
	case "GET":
		r, err := http.Get(ep.URL)
		if err != nil {
			slog.Error("GET request error", "endpoint", ep.URL, "error", err)
			os.Exit(1)
		}
		resp = r
	case "POST":
		// NOTE: Nil for now for quick testing
		var body io.Reader = nil

		r, err := http.Post(ep.URL, "application/json", body)
		if err != nil {
			slog.Error("POST request error", "endpoint", ep.URL, "error", err)
			os.Exit(1)
		}
		resp = r
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		slog.Error("bad status", "status", resp.StatusCode)
		os.Exit(1)
	}

	slog.Info("Succesfully poked", "endpoint", ep.URL)
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
			ep.poke()
		})
	}

	wg.Wait()

	slog.Info("Time elapsed", "time", time.Since(currTime))
	slog.Info("Poked endpoints", "count", len(endpoints))
}
