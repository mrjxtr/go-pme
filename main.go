package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"sync"
	"time"
)

type Endpoint struct {
	URL    string            `json:"url"`
	Method string            `json:"method"`
	Params map[string]string `json:"params,omitempty"`
}

// poke makes a request to the endpoint with the specified method and logs the result
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
	default:
		return fmt.Errorf("unsupported HTTP method %q for %s", ep.Method, ep.URL)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status %d for %s", resp.StatusCode, ep.URL)
	}

	slog.Info("Succesfully poked", "endpoint", ep.URL)
	return nil
}

// loadEndpointsJSON loads the endpoints from a specified JSON file
// if no filename is provided, it will default to "endpoints.json"
func loadEndpointsJSON(filename ...string) ([]Endpoint, error) {
	jsonFile := "endpoints.json"
	if len(filename) > 0 && filename[0] != "" {
		jsonFile = filename[0]
	}

	data, err := os.ReadFile(jsonFile)
	if err != nil {
		return nil, err
	}

	var endpoints []Endpoint
	if err := json.Unmarshal(data, &endpoints); err != nil {
		return nil, err
	}

	for _, ep := range endpoints {
		slog.Info("url found", "url", ep.URL)
	}

	return endpoints, nil
}

func main() {
	endpoints, err := loadEndpointsJSON()
	if err != nil {
		slog.Error("error getting endpoints", "error", err)
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
