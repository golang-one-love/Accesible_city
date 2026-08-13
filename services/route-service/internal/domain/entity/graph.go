package entity

import (
	"encoding/json"
	"math"
)

type Node struct {
	ID        int64
	Latitude  float64
	Longitude float64
}

type Edge struct {
	To       int64
	Distance float64
	Severity int // 1-5, higher = more difficult
}

// Graph keeps two representations:
//   - a build form (map-based) used while nodes/edges are added from OSM or
//     the legacy database blob, and
//   - a compact form produced by BuildIndex: flat arrays for nodes/edges and
//     a spatial grid. The compact form is several times smaller in memory,
//     which matters for multi-million-node city-wide graphs.
type Graph struct {
	buildNodes map[int64]Node
	buildEdges map[int64][]Edge

	compact bool

	nodeByID  map[int64]int32 // node id -> index into nodeData
	nodeData  []Node
	edgeStart []int64 // edgeStart[i]..edgeStart[i+1] are edges of nodeData[i]
	edgeData  []Edge

	grid map[[2]int32][]int32 // spatial index: cellKey -> node indices
}

func NewGraph() *Graph {
	return &Graph{
		buildNodes: make(map[int64]Node),
		buildEdges: make(map[int64][]Edge),
	}
}

func (g *Graph) AddNode(node Node) {
	g.buildNodes[node.ID] = node
}

func (g *Graph) AddEdge(from int64, edge Edge) {
	g.buildEdges[from] = append(g.buildEdges[from], edge)
}

// NumNodes reports the number of nodes in the graph.
func (g *Graph) NumNodes() int {
	if g.compact {
		return len(g.nodeData)
	}
	return len(g.buildNodes)
}

func (g *Graph) GetNode(id int64) (Node, bool) {
	if g.compact {
		idx, ok := g.nodeByID[id]
		if !ok {
			return Node{}, false
		}
		return g.nodeData[idx], true
	}
	node, ok := g.buildNodes[id]
	return node, ok
}

func (g *Graph) GetEdges(from int64) []Edge {
	if g.compact {
		idx, ok := g.nodeByID[from]
		if !ok {
			return nil
		}
		return g.edgeData[g.edgeStart[idx]:g.edgeStart[idx+1]]
	}
	return g.buildEdges[from]
}

// graphJSON mirrors the persisted blob format (map keys are node ids).
type graphJSON struct {
	Nodes map[int64]Node   `json:"nodes"`
	Edges map[int64][]Edge `json:"edges"`
}

// MarshalJSON persists the graph in either representation.
func (g *Graph) MarshalJSON() ([]byte, error) {
	if g.compact {
		nodes := make(map[int64]Node, len(g.nodeData))
		for _, n := range g.nodeData {
			nodes[n.ID] = n
		}
		edges := make(map[int64][]Edge, len(g.nodeData))
		for i := range g.nodeData {
			edges[g.nodeData[i].ID] = g.edgeData[g.edgeStart[i]:g.edgeStart[i+1]]
		}
		return json.Marshal(graphJSON{Nodes: nodes, Edges: edges})
	}
	return json.Marshal(graphJSON{Nodes: g.buildNodes, Edges: g.buildEdges})
}

// gridCellSize is ~111 meters; cells bucket nodes so nearest-node lookups
// do not scan the whole graph on every request.
const gridCellSize = 0.001

// BuildIndex converts the graph to its compact form and buckets all nodes
// into a lat/lon grid for fast FindNearest. Must be called once after the
// graph is populated (before serving requests). Adding nodes or edges after
// this call is not allowed.
func (g *Graph) BuildIndex() {
	if g.compact {
		return
	}

	g.nodeByID = make(map[int64]int32, len(g.buildNodes))
	g.nodeData = make([]Node, 0, len(g.buildNodes))
	for id, n := range g.buildNodes {
		g.nodeByID[id] = int32(len(g.nodeData))
		g.nodeData = append(g.nodeData, n)
	}
	g.buildNodes = nil

	g.edgeStart = make([]int64, len(g.nodeData)+1)
	g.edgeData = make([]Edge, 0, len(g.buildEdges))
	for i := 0; i < len(g.nodeData); i++ {
		g.edgeStart[i] = int64(len(g.edgeData))
		g.edgeData = append(g.edgeData, g.buildEdges[g.nodeData[i].ID]...)
	}
	g.edgeStart[len(g.nodeData)] = int64(len(g.edgeData))
	g.buildEdges = nil

	g.grid = make(map[[2]int32][]int32, len(g.nodeData)/4)
	for i, n := range g.nodeData {
		key := cellKey(n.Latitude, n.Longitude)
		g.grid[key] = append(g.grid[key], int32(i))
	}

	g.compact = true
}

func cellIndex(lat, lon float64) (int32, int32) {
	return int32(math.Floor(lat / gridCellSize)), int32(math.Floor(lon / gridCellSize))
}

func cellKey(lat, lon float64) [2]int32 {
	ci, cj := cellIndex(lat, lon)
	return [2]int32{ci, cj}
}

// FindNearest returns the node closest to (lat, lon) and its distance in
// meters. Falls back to a full scan when BuildIndex was not called or no
// node was found in the surrounding grid cells.
func (g *Graph) FindNearest(lat, lon float64) (int64, float64, bool) {
	if g.compact {
		if len(g.nodeData) == 0 {
			return 0, 0, false
		}
		if g.grid == nil {
			return g.findNearestFullScan(lat, lon)
		}

		ci, cj := cellIndex(lat, lon)
		cellMeters := gridCellSize * 111000.0

		var bestID int64
		bestDist := math.MaxFloat64
		found := false
		for ring := 0; ring <= 25; ring++ {
			// A node in ring r+1 is at least (r-1.5)*cellMeters away (cell
			// spacing minus the point's worst-case offset inside its cell), so
			// once bestDist is below that, expanding further cannot find a
			// closer node.
			if found && bestDist <= (float64(ring)-1.5)*cellMeters {
				break
			}
			for di := -ring; di <= ring; di++ {
				for dj := -ring; dj <= ring; dj++ {
					if di > -ring && di < ring && dj > -ring && dj < ring {
						continue // interior cells were scanned in earlier rings
					}
					for _, idx := range g.grid[[2]int32{ci + int32(di), cj + int32(dj)}] {
						n := g.nodeData[idx]
						d := haversineMeters(lat, lon, n.Latitude, n.Longitude)
						if d < bestDist {
							bestDist = d
							bestID = n.ID
							found = true
						}
					}
				}
			}
		}
		if found {
			return bestID, bestDist, true
		}
		return g.findNearestFullScan(lat, lon)
	}

	if len(g.buildNodes) == 0 {
		return 0, 0, false
	}
	var bestID int64
	bestDist := math.MaxFloat64
	found := false
	for id, n := range g.buildNodes {
		d := haversineMeters(lat, lon, n.Latitude, n.Longitude)
		if d < bestDist {
			bestDist = d
			bestID = id
			found = true
		}
	}
	if !found {
		return 0, 0, false
	}
	return bestID, bestDist, true
}

func (g *Graph) findNearestFullScan(lat, lon float64) (int64, float64, bool) {
	var bestID int64
	bestDist := math.MaxFloat64
	found := false
	for i, n := range g.nodeData {
		d := haversineMeters(lat, lon, n.Latitude, n.Longitude)
		if d < bestDist {
			bestDist = d
			bestID = n.ID
			found = true
		}
		_ = i
	}
	if !found {
		return 0, 0, false
	}
	return bestID, bestDist, true
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
