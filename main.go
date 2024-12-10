package main

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"time"
)

var (
	status          bool
	lastMonitoredAt time.Time
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	addr := ":8080"
	if os.Getenv("PORT") != "" {
		addr = ":" + os.Getenv("PORT")
	}

	endpoint := os.Getenv("MONITOR_ZEABUR_ENDPOINT")
	if endpoint == "" {
		endpoint = "https://gateway.zeabur.com/graphql"
	}

	serviceID := os.Getenv("MONITOR_SERVICE_ID")
	environmentID := os.Getenv("MONITOR_ENVIRONMENT_ID")
	zeaburToken := os.Getenv("MONITOR_ZEABUR_TOKEN")

	if serviceID == "" || environmentID == "" || zeaburToken == "" {
		slog.Error("missing required environment variables")
		os.Exit(1)
	}

	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	// start a goroutine to get status
	go func() {
		for {
			var output struct {
				Data struct {
					Service struct {
						Status string `json:"status"`
					} `json:"service"`
				} `json:"data"`
			}

			request, err := createGetStatusRequest(ctx, endpoint, serviceID, environmentID, zeaburToken)
			if err != nil {
				slog.Error("failed to create request", slog.Any("error", err))
				os.Exit(1)
			}

			response, err := http.DefaultClient.Do(request)
			if err != nil {
				slog.Error("failed to send request", slog.Any("error", err))
				continue
			}

			err = json.NewDecoder(response.Body).Decode(&output)
			_ = response.Body.Close()

			if err != nil {
				slog.Error("failed to decode response", slog.Any("error", err))
				continue
			}

			slog.Info("write status", slog.String("status", output.Data.Service.Status), slog.Time("at", time.Now()))
			status = output.Data.Service.Status == "RUNNING"
			lastMonitoredAt = time.Now()

			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
			}
		}
	}()

	http.HandleFunc("GET /alive", func(w http.ResponseWriter, r *http.Request) {
		jw := json.NewEncoder(w)

		payload := map[string]any{
			"success":       status,
			"lastCheckedAt": lastMonitoredAt,
		}

		if status {
			w.WriteHeader(http.StatusOK)
			_ = jw.Encode(payload)
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
			_ = jw.Encode(payload)
		}
	})

	slog.Info("server started", slog.String("addr", addr))
	if err := http.ListenAndServe(addr, nil); err != nil {
		slog.Error("failed to start server", slog.Any("error", err))
	}
}

func createGetStatusRequest(ctx context.Context, endpoint, serviceID, environmentID, zeaburToken string) (*http.Request, error) {
	payload := map[string]any{
		"query": `query Service($id: ObjectID, $environmentId: ObjectID!) {
			service(_id: $id) {
			  status(environmentID: $environmentId)
			}
		}`,
		"variables": map[string]any{
			"id":            serviceID,
			"environmentId": environmentID,
		},
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		slog.Error("failed to marshal payload", slog.Any("error", err))
		return nil, err
	}

	request, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(payloadJSON))
	if err != nil {
		slog.Error("failed to create GraphQL request", slog.Any("error", err))
		return nil, err
	}

	request.Header.Set("Authorization", "Bearer "+zeaburToken)
	request.Header.Set("Content-Type", "application/json")

	return request, nil
}
