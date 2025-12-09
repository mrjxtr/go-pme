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

// Endpoint represents a single API endpoint with its configuration.
type Endpoint struct {
	Name    string            `json:"name"`
	URL     string            `json:"url"`
	Method  string            `json:"method"`
	Headers map[string]string `json:"headers"`
}

// poke makes a request to the endpoint with the specified method and logs the result.
// if headers are provided, values will be replaced with environment variables
//
// e.g headers: { "apikey" : "API_KEY"}, make sure on your `.env` you've set your `API_KEY` variable
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
		value = os.Getenv(value)
		req.Header.Set(key, value)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf(
			"%s request error for %s %s: %w",
			ep.Method,
			ep.Name,
			ep.URL,
			err,
		)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status %d for %s %s", resp.StatusCode, ep.Name, ep.URL)
	}

	slog.Info(
		"Succesfully poked",
		"name",
		ep.Name,
		"endpoint",
		ep.URL,
		"status",
		resp.StatusCode,
	)
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
		slog.Info("endpoints found", "name", ep.Name, "url", ep.URL)
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
}
