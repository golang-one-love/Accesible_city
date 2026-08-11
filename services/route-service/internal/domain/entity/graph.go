package entity

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