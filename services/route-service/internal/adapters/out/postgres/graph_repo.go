package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/accessible-path/route-service/internal/domain/entity"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresGraphRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresGraphRepository(pool *pgxpool.Pool) *PostgresGraphRepository {
	return &PostgresGraphRepository{pool: pool}
}

func (r *PostgresGraphRepository) SaveGraph(graph *entity.Graph) error {
	ctx := context.Background()

	graphJSON, err := json.Marshal(graph)
	if err != nil {
		return fmt.Errorf("marshal graph: %w", err)
	}

	_, err = r.pool.Exec(ctx, `
		INSERT INTO osm_graph_cache (id, nodes, edges, updated_at)
		VALUES (1, $1::jsonb->'nodes', $1::jsonb->'edges', NOW())
		ON CONFLICT (id) DO UPDATE SET nodes = $1::jsonb->'nodes', edges = $1::jsonb->'edges', updated_at = NOW()
	`, graphJSON)

	return err
}

func (r *PostgresGraphRepository) LoadGraph() (*entity.Graph, error) {
	ctx := context.Background()

	var nodesJSON, edgesJSON []byte
	err := r.pool.QueryRow(ctx, `SELECT nodes, edges FROM osm_graph_cache WHERE id = 1`).Scan(&nodesJSON, &edgesJSON)
	if err != nil {
		return nil, err
	}

	graph := entity.NewGraph()

	// Fast path: current blob format with numeric keys ("12345"). Decoding
	// straight into int64 maps avoids the string-heavy legacy structures,
	// which matters for multi-million-node city-wide graphs.
	var fastNodes map[int64]entity.Node
	if err := json.Unmarshal(nodesJSON, &fastNodes); err == nil {
		var fastEdges map[int64][]entity.Edge
		if err := json.Unmarshal(edgesJSON, &fastEdges); err == nil {
			for id, n := range fastNodes {
				graph.AddNode(entity.Node{ID: id, Latitude: n.Latitude, Longitude: n.Longitude})
			}
			for from, edges := range fastEdges {
				for _, e := range edges {
					graph.AddEdge(from, e)
				}
			}
			return graph, nil
		}
	}

	// Legacy path: string keys ("osm:12345"); both endpoints were strings
	// in the old edge shape. Converted to the numeric ids used in memory.
	legacyNodes := make(map[string]struct {
		Latitude  float64
		Longitude float64
	})
	if err := json.Unmarshal(nodesJSON, &legacyNodes); err != nil {
		return nil, fmt.Errorf("unmarshal nodes: %w", err)
	}
	var legacyEdges map[string][]struct {
		To       json.RawMessage
		Distance float64
		Severity int
	}
	if err := json.Unmarshal(edgesJSON, &legacyEdges); err != nil {
		return nil, fmt.Errorf("unmarshal edges: %w", err)
	}

	for key, n := range legacyNodes {
		id, err := parseNodeID(key)
		if err != nil {
			return nil, err
		}
		graph.AddNode(entity.Node{ID: id, Latitude: n.Latitude, Longitude: n.Longitude})
	}
	for key, edges := range legacyEdges {
		from, err := parseNodeID(key)
		if err != nil {
			return nil, err
		}
		for _, e := range edges {
			to, err := parseEdgeRef(e.To)
			if err != nil {
				return nil, err
			}
			graph.AddEdge(from, entity.Edge{To: to, Distance: e.Distance, Severity: e.Severity})
		}
	}

	return graph, nil
}

// parseEdgeRef converts an edge destination that is either a JSON number
// (current format) or a prefixed string ("osm:12345", legacy format).
func parseEdgeRef(raw json.RawMessage) (int64, error) {
	s := strings.TrimSpace(string(raw))
	if strings.HasPrefix(s, `"`) {
		var str string
		if err := json.Unmarshal(raw, &str); err != nil {
			return 0, err
		}
		return parseNodeID(str)
	}
	return strconv.ParseInt(s, 10, 64)
}

// parseNodeID converts a graph blob key ("osm:12345" or "12345") to the
// numeric node id used in memory.
func parseNodeID(key string) (int64, error) {
	key = strings.TrimPrefix(key, "osm:")
	id, err := strconv.ParseInt(key, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid node id %q: %w", key, err)
	}
	return id, nil
}
