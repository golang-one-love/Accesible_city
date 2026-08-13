package external

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/accessible-path/route-service/internal/ports/out"
)

type BarrierClient struct {
	baseURL string
	client  *http.Client
}

func NewBarrierClient(baseURL string) *BarrierClient {
	return &BarrierClient{
		baseURL: baseURL,
		client:  &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *BarrierClient) GetBarriersInBounds(minLat, minLon, maxLat, maxLon float64) ([]out.BarrierInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	url := fmt.Sprintf("%s/api/v1/barriers?status=approved&sw_lat=%f&sw_lon=%f&ne_lat=%f&ne_lon=%f&limit=1000",
		c.baseURL, minLat, minLon, maxLat, maxLon)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("barrier service returned %d", resp.StatusCode)
	}

	var result struct {
		Barriers []struct {
			ID          string  `json:"id"`
			Coordinates struct {
				Latitude  float64 `json:"latitude"`
				Longitude float64 `json:"longitude"`
			} `json:"coordinates"`
			Severity int `json:"severity"`
		} `json:"barriers"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	barriers := make([]out.BarrierInfo, len(result.Barriers))
	for i, b := range result.Barriers {
		barriers[i] = out.BarrierInfo{
			ID:        b.ID,
			Latitude:  b.Coordinates.Latitude,
			Longitude: b.Coordinates.Longitude,
			Severity:  b.Severity,
		}
	}

	return barriers, nil
}

func (c *BarrierClient) GetBarrier(barrierID string) (*out.BarrierInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	url := fmt.Sprintf("%s/api/v1/barriers/%s", c.baseURL, barrierID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("barrier service returned %d for barrier %s", resp.StatusCode, barrierID)
	}

	var result struct {
		ID          string  `json:"id"`
		Coordinates struct {
			Latitude  float64 `json:"latitude"`
			Longitude float64 `json:"longitude"`
		} `json:"coordinates"`
		Severity int `json:"severity"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &out.BarrierInfo{
		ID:        result.ID,
		Latitude:  result.Coordinates.Latitude,
		Longitude: result.Coordinates.Longitude,
		Severity:  result.Severity,
	}, nil
}