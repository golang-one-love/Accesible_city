package cache

import (
	"sync"

	"github.com/accessible-path/route-service/internal/domain/entity"
	"github.com/accessible-path/route-service/internal/ports/out"
)

type MemoryGraphCache struct {
	mu    sync.RWMutex
	graph *entity.Graph
}

func NewMemoryGraphCache() *MemoryGraphCache {
	return &MemoryGraphCache{}
}

func (c *MemoryGraphCache) Get() *entity.Graph {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.graph
}

func (c *MemoryGraphCache) Set(graph *entity.Graph) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.graph = graph
}

var _ out.GraphRepository = (*MemoryGraphCache)(nil)

func (c *MemoryGraphCache) SaveGraph(graph *entity.Graph) error {
	c.Set(graph)
	return nil
}

func (c *MemoryGraphCache) LoadGraph() (*entity.Graph, error) {
	return c.Get(), nil
}