package entity

import (
	"fmt"
	"math"
)

type Node struct {
	ID        string
	Latitude  float64
	Longitude float64
}

type Edge struct {
	From     string
	To       string
	Distance float64
	Severity int // 1-5, higher = more difficult
}

type Graph struct {
	Nodes map[string]*Node
	Edges map[string][]*Edge // fromNodeID -> edges

	grid map[string][]*Node // spatial index: cellKey -> nodes
}

func NewGraph() *Graph {
	return &Graph{
		Nodes: make(map[string]*Node),
		Edges: make(map[string][]*Edge),
	}
}

func (g *Graph) AddNode(node *Node) {
	g.Nodes[node.ID] = node
}

func (g *Graph) AddEdge(edge *Edge) {
	g.Edges[edge.From] = append(g.Edges[edge.From], edge)
}

func (g *Graph) GetNode(id string) *Node {
	return g.Nodes[id]
}

func (g *Graph) GetEdges(from string) []*Edge {
	return g.Edges[from]
}

// gridCellSize is ~111 meters; cells bucket nodes so nearest-node lookups
// do not scan the whole graph on every request.
const gridCellSize = 0.001

// BuildIndex buckets all nodes into a lat/lon grid for fast FindNearest.
// Must be called once after the graph is populated (before serving requests).
func (g *Graph) BuildIndex() {
	g.grid = make(map[string][]*Node, len(g.Nodes)/4)
	for _, n := range g.Nodes {
		key := gridKey(n.Latitude, n.Longitude)
		g.grid[key] = append(g.grid[key], n)
	}
}

func cellIndex(lat, lon float64) (int, int) {
	return int(math.Floor(lat / gridCellSize)), int(math.Floor(lon / gridCellSize))
}

func gridKey(lat, lon float64) string {
	ci, cj := cellIndex(lat, lon)
	return fmt.Sprintf("%d:%d", ci, cj)
}

// FindNearest returns the node closest to (lat, lon) and its distance in
// meters. Falls back to a full scan when BuildIndex was not called or no
// node was found in the surrounding grid cells.
func (g *Graph) FindNearest(lat, lon float64) (*Node, float64) {
	if len(g.Nodes) == 0 {
		return nil, 0
	}
	if g.grid == nil {
		return g.findNearestFullScan(lat, lon)
	}

	ci, cj := cellIndex(lat, lon)
	cellMeters := gridCellSize * 111000.0

	var best *Node
	bestDist := math.MaxFloat64
	for ring := 0; ring <= 25; ring++ {
		// A node in ring r+1 is at least (r-1.5)*cellMeters away (cell
		// spacing minus the point's worst-case offset inside its cell), so
		// once bestDist is below that, expanding further cannot find a
		// closer node.
		if best != nil && bestDist <= (float64(ring)-1.5)*cellMeters {
			break
		}
		for di := -ring; di <= ring; di++ {
			for dj := -ring; dj <= ring; dj++ {
				if di > -ring && di < ring && dj > -ring && dj < ring {
					continue // interior cells were scanned in earlier rings
				}
				for _, n := range g.grid[fmt.Sprintf("%d:%d", ci+di, cj+dj)] {
					d := haversineMeters(lat, lon, n.Latitude, n.Longitude)
					if d < bestDist {
						bestDist = d
						best = n
					}
				}
			}
		}
	}
	if best != nil {
		return best, bestDist
	}
	return g.findNearestFullScan(lat, lon)
}

func (g *Graph) findNearestFullScan(lat, lon float64) (*Node, float64) {
	var best *Node
	bestDist := math.MaxFloat64
	for _, n := range g.Nodes {
		d := haversineMeters(lat, lon, n.Latitude, n.Longitude)
		if d < bestDist {
			bestDist = d
			best = n
		}
	}
	return best, bestDist
}

// haversineMeters is a copy of valueobject.Coordinates.HaversineDistance
// kept local to avoid an entity -> valueobject dependency.
func haversineMeters(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371000 // Earth radius in meters
	rlat1 := lat1 * math.Pi / 180
	rlat2 := lat2 * math.Pi / 180
	dLat := (lat2 - lat1) * math.Pi / 180
	dLon := (lon2 - lon1) * math.Pi / 180

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(rlat1)*math.Cos(rlat2)*math.Sin(dLon/2)*math.Sin(dLon/2)
	return R * 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
}