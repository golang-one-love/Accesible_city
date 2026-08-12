package external

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/accessible-path/route-service/internal/domain/entity"
	"github.com/accessible-path/route-service/internal/domain/valueobject"
)

const osmHighwayFilter = "^(footway|pedestrian|path|cycleway|residential|living_street|service|track|unclassified|tertiary|secondary|primary|steps)$"

type OSMClient struct {
	overpassURL string
	httpClient  *http.Client
}

func NewOSMClient(overpassURL string) *OSMClient {
	return &OSMClient{
		overpassURL: overpassURL,
		httpClient:  &http.Client{Timeout: 120 * time.Second},
	}
}

func (c *OSMClient) FetchRoads(ctx context.Context, minLat, minLon, maxLat, maxLon float64) (*entity.Graph, error) {
	query := fmt.Sprintf(
		`[out:json][timeout:90];(way["highway"~"%s"](%f,%f,%f,%f););out body;>;out skel qt;`,
		osmHighwayFilter, minLat, minLon, maxLat, maxLon,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.overpassURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.URL.RawQuery = url.Values{"data": []string{query}}.Encode()
	req.Header.Set("User-Agent", "accessible-path-route-service/1.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("overpass request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("overpass returned %d", resp.StatusCode)
	}

	var data struct {
		Elements []struct {
			Type  string            `json:"type"`
			ID    int64             `json:"id"`
			Lat   float64           `json:"lat"`
			Lon   float64           `json:"lon"`
			Nodes []int64           `json:"nodes"`
			Tags  map[string]string `json:"tags"`
		} `json:"elements"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("decode overpass response: %w", err)
	}

	nodes := make(map[int64]*entity.Node)
	for _, el := range data.Elements {
		if el.Type == "node" {
			nodes[el.ID] = &entity.Node{
				ID:        osmNodeID(el.ID),
				Latitude:  el.Lat,
				Longitude: el.Lon,
			}
		}
	}

	graph := entity.NewGraph()
	for _, n := range nodes {
		graph.AddNode(n)
	}

	for _, el := range data.Elements {
		if el.Type != "way" || len(el.Nodes) < 2 {
			continue
		}
		severity := highwaySeverity(el.Tags["highway"])
		if severity == 0 {
			continue
		}
		for i := 0; i < len(el.Nodes)-1; i++ {
			from, okFrom := nodes[el.Nodes[i]]
			to, okTo := nodes[el.Nodes[i+1]]
			if !okFrom || !okTo {
				continue
			}
			dist := valueobject.Coordinates{Latitude: from.Latitude, Longitude: from.Longitude}.
				HaversineDistance(valueobject.Coordinates{Latitude: to.Latitude, Longitude: to.Longitude})
			graph.AddEdge(&entity.Edge{From: from.ID, To: to.ID, Distance: dist, Severity: severity})
			graph.AddEdge(&entity.Edge{From: to.ID, To: from.ID, Distance: dist, Severity: severity})
		}
	}

	return graph, nil
}

func osmNodeID(id int64) string {
	return "osm:" + strconv.FormatInt(id, 10)
}

// highwaySeverity returns 0 for highways that are not pedestrian-routable.
func highwaySeverity(h string) int {
	switch h {
	case "footway", "pedestrian", "path", "cycleway":
		return 1
	case "living_street", "residential", "service", "track":
		return 2
	case "unclassified":
		return 3
	case "tertiary":
		return 4
	case "secondary", "primary", "steps":
		return 5
	default:
		return 0
	}
}