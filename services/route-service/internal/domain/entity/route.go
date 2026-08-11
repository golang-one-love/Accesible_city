package entity

type Route struct {
	Nodes       []*Node
	TotalDistance float64
	MaxSeverity   int
}

func NewRoute() *Route {
	return &Route{
		Nodes:         make([]*Node, 0),
		TotalDistance: 0,
		MaxSeverity:   0,
	}
}

func (r *Route) AddNode(node *Node, distance float64, severity int) {
	r.Nodes = append(r.Nodes, node)
	r.TotalDistance += distance
	if severity > r.MaxSeverity {
		r.MaxSeverity = severity
	}
}