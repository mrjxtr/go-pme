package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/joho/godotenv"
)

const (
	DefaultConfigFile    string        = "endpoints.json"
	DefaultClientTimeout time.Duration = 10 * time.Second
	DefaultErrSnipSize   int64         = 512
)

// Endpoint represents a single API endpoint with its configuration.
type Endpoint struct {
	Name    string            `json:"name"`
	URL     string            `json:"url"`
	Method  string            `json:"method"`
	Headers map[string]string `json:"headers"`
	Payload map[string]string `json:"payload"`
}

// TODO: Add a "App" struct

// poke makes a request to the endpoint with the specified method and logs the result.
// if headers are provided, values will be replaced with environment variables
//
// e.g headers: { "apikey" : "API_KEY"}, make sure on your `.env` you've set your `API_KEY` variable
func (ep Endpoint) poke(ctx context.Context, client *http.Client) error {
	var body io.Reader

	if len(ep.Payload) > 0 {
		payload := make(map[string]string, len(ep.Payload))

		for key, envName := range ep.Payload {
			v, ok := os.LookupEnv(envName)
			if !ok || v == "" {
				return fmt.Errorf(
					"env var %#q (referenced by payload key %#q for %#q) is unset or empty",
					envName,
					key,
					ep.Name,
				)
			}
			payload[key] = v
		}

		jsonData, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf(
				"failed to marshal json payload for %s: %w",
				ep.URL,
				err,
			)
		}
		body = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequestWithContext(ctx, ep.Method, ep.URL, body)
	if err != nil {
		return fmt.Errorf(
			"failed to create %s request for %s: %w",
			ep.Method,
			ep.URL,
			err,
		)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	for key, envName := range ep.Headers {
		v, ok := os.LookupEnv(envName)
		if !ok || v == "" {
			return fmt.Errorf(
				"env var %#q (referenced by header key %#q for %#q) is unset or empty",
				envName,
				key,
				ep.Name,
			)
		}
		req.Header.Set(key, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf(
			"%s request error for %s: %w",
			ep.Method,
			ep.Name,
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
		"status",
		resp.StatusCode,
	)
	return nil
}

// loadEndpointsJSON loads the endpoints from a specified JSON file
// if no filename is provided, it will default to "endpoints.json"
func loadEndpointsJSON(filename ...string) ([]Endpoint, error) {
	jsonFile := DefaultConfigFile
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

	var errs []error
	for i, ep := range endpoints {
		if ep.Name == "" {
			errs = append(
				errs,
				fmt.Errorf("invalid endpoint: index %d missing %q", i, "name"),
			)
		}
		if ep.URL == "" {
			errs = append(
				errs,
				fmt.Errorf(
					"invalid endpoint: %q (index %d) missing %q",
					ep.Name,
					i,
					"url",
				),
			)
		}
		if ep.Method == "" {
			errs = append(
				errs,
				fmt.Errorf(
					"invalid endpoint: %q (index %d) missing %q",
					ep.Name,
					i,
					"method",
				),
			)
		}
	}
	if err := errors.Join(errs...); err != nil {
		return nil, err
	}

	for _, ep := range endpoints {
		slog.Info("endpoint loaded", "name", ep.Name, "url", ep.URL)
	}

	return endpoints, nil
}

func main() {
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		slog.Error("error loading env", "error", err)
		os.Exit(1)
	}

	endpoints, err := loadEndpointsJSON()
	if err != nil {
		if joined, ok := err.(interface{ Unwrap() []error }); ok {
			for _, e := range joined.Unwrap() {
				slog.Error("error loading endpoints", "error", e)
			}
		} else {
			slog.Error("error loading endpoints", "error", err)
		}
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	client := &http.Client{Timeout: DefaultClientTimeout}

	var wg sync.WaitGroup

	slog.Info("Poking endpoints", "count", len(endpoints))

	currTime := time.Now()

	for _, endpoint := range endpoints {
		ep := endpoint
		wg.Go(func() {
			if err := ep.poke(ctx, client); err != nil {
				slog.Error("poke failed", "error", err)
			}
		})
	}

	wg.Wait()

	slog.Info("Time elapsed", "time", time.Since(currTime))
}
