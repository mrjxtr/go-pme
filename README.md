# go-pme

Poke My Endpoints - a lightweight tool to hit multiple HTTP endpoints concurrently. Useful for keeping services warm or triggering scheduled jobs.

## Setup

1. Install dependencies:

   ```bash
   go mod tidy
   ```

2. Create a `.env` file for any API keys or secrets. `.env.example` available for reference:

   ```bash
   cp .env.example .env
   ```

   Then edit `.env` with your actual API keys.

3. Create an `endpoints.json` file. `example_endpoints.json` available for reference:

   ```json
   [
     {
       "name": "Endpoint 1",
       "url": "https://randomapi.xyz/api/v1/endpoint",
       "method": "GET"
     },
     {
       "name": "Endpoint 2",
       "url": "https://randomapi.io/api/v2/endpoint",
       "method": "GET",
       "headers": { "apikey": "API_KEY" }
     }
     {
       "name": "Endpoint 3",
       "url": "https://randomapi.dev/api/v3/endpoint",
       "method": "POST",
       "headers": { "apikey" : "API_KEY"},
       "payload": { "email": "EMAIL", "password": "PASSWORD"}
     }
   ]
   ```

## Endpoint Config

| Field     | Required | Description                                       |
| --------- | -------- | ------------------------------------------------- |
| `name`    | yes      | Label for logging                                 |
| `url`     | yes      | Full URL to poke                                  |
| `method`  | yes      | HTTP method (GET, POST, etc.)                     |
| `headers` | no       | Key-value pairs where values are `.env` var names |
| `payload` | no       | Key-value pairs where values are `.env` var names |

> Header values reference environment variable names, not the actual secrets. This keeps your `endpoints.json` safe to commit if desired.

## Usage

```bash
go run main.go
```

Or build and run:

```bash
go build -o go-pme && ./go-pme
```
