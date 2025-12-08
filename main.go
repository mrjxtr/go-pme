package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/joho/godotenv"
)

type Endpoint struct {
	URL     string            `json:"url"`
	Method  string            `json:"method"`
	Headers map[string]string `json:"headers,omitempty"`
}

// poke makes a request to the endpoint with the specified method and logs the result
func (ep *Endpoint) poke() error {
	req, err := http.NewRequest(ep.Method, ep.URL, nil)
	if err != nil {
		return fmt.Errorf(
			"failed to create %s request for %s: %w",
			ep.Method,
			ep.URL,
			err,
		)
	}

	for key, value := range ep.Headers {
		req.Header.Set(key, value)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("%s request error for %s: %w", ep.Method, ep.URL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status %d for %s", resp.StatusCode, ep.URL)
	}

	slog.Info("Succesfully poked", "endpoint", ep.URL, "status", resp.StatusCode)
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
	if err := godotenv.Load(); err != nil {
		slog.Error("error loading env", "error", err)
		os.Exit(1)
	}

	endpoints, err := loadEndpointsJSON()
	if err != nil {
		slog.Error("error getting endpoints", "error", err)
		os.Exit(1)
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
