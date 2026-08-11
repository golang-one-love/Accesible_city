package postgres

import (
	"context"
	"encoding/json"
	"fmt"

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
	
	nodesJSON, err := json.Marshal(graph.Nodes)
	if err != nil {
		return fmt.Errorf("marshal nodes: %w", err)
	}
	
	edgesJSON, err := json.Marshal(graph.Edges)
	if err != nil {
		return fmt.Errorf("marshal edges: %w", err)
	}

	_, err = r.pool.Exec(ctx, `
		INSERT INTO osm_graph_cache (id, nodes, edges, updated_at)
		VALUES (1, $1, $2, NOW())
		ON CONFLICT (id) DO UPDATE SET nodes = $1, edges = $2, updated_at = NOW()
	`, nodesJSON, edgesJSON)
	
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
	
	if err := json.Unmarshal(nodesJSON, &graph.Nodes); err != nil {
		return nil, fmt.Errorf("unmarshal nodes: %w", err)
	}
	if err := json.Unmarshal(edgesJSON, &graph.Edges); err != nil {
		return nil, fmt.Errorf("unmarshal edges: %w", err)
	}

	return graph, nil
}