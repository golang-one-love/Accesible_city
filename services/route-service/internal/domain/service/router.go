package service

import (
	"container/heap"
	"errors"
	"math"

	"github.com/accessible-path/route-service/internal/domain/entity"
	"github.com/accessible-path/route-service/internal/domain/valueobject"
	"github.com/accessible-path/route-service/internal/ports/out"
)

var (
	ErrNoPathFound    = errors.New("no path found")
	ErrInvalidProfile = errors.New("invalid mobility profile")
)

type Router struct {
	graph       *entity.Graph
	barrierRepo out.BarrierProvider
}

func NewRouter(graph *entity.Graph, barrierRepo out.BarrierProvider) *Router {
	return &Router{
		graph:       graph,
		barrierRepo: barrierRepo,
	}
}

func (r *Router) SetGraph(graph *entity.Graph) {
	r.graph = graph
}

func (r *Router) BuildRoute(start, finish valueobject.Coordinates, profile entity.MobilityProfile) (*entity.Route, error) {
	if !profile.IsValid() {
		return nil, ErrInvalidProfile
	}

	maxSeverity := profile.MaxSeverity()

	startNode := r.findNearestNode(start)
	finishNode := r.findNearestNode(finish)

	if startNode == nil || finishNode == nil {
		return nil, ErrNoPathFound
	}

	barriers, err := r.barrierRepo.GetBarriersInBounds(
		start.Latitude, start.Longitude,
		finish.Latitude, finish.Longitude,
	)
	if err != nil {
		return nil, err
	}

	blockedNodes := make(map[string]bool)
	for _, b := range barriers {
		if b.Severity <= maxSeverity {
			node := r.findNearestNode(valueobject.Coordinates{Latitude: b.Latitude, Longitude: b.Longitude})
			if node != nil {
				blockedNodes[node.ID] = true
			}
		}
	}

	return r.aStar(startNode.ID, finishNode.ID, blockedNodes, maxSeverity)
}

func (r *Router) findNearestNode(coord valueobject.Coordinates) *entity.Node {
	var nearest *entity.Node
	minDist := math.MaxFloat64

	for _, node := range r.graph.Nodes {
		d := coord.HaversineDistance(valueobject.Coordinates{Latitude: node.Latitude, Longitude: node.Longitude})
		if d < minDist {
			minDist = d
			nearest = node
		}
	}
	return nearest
}

type aStarNode struct {
	nodeID string
	gScore float64
	fScore float64
	index  int
}

type priorityQueue []*aStarNode

func (pq priorityQueue) Len() int { return len(pq) }
func (pq priorityQueue) Less(i, j int) bool { return pq[i].fScore < pq[j].fScore }
func (pq priorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}
func (pq *priorityQueue) Push(x interface{}) {
	n := x.(*aStarNode)
	n.index = len(*pq)
	*pq = append(*pq, n)
}
func (pq *priorityQueue) Pop() interface{} {
	old := *pq
	n := old[len(old)-1]
	old[len(old)-1] = nil
	*pq = old[0 : len(old)-1]
	return n
}

func (r *Router) aStar(startID, finishID string, blockedNodes map[string]bool, maxSeverity int) (*entity.Route, error) {
	startNode := r.graph.GetNode(startID)
	finishNode := r.graph.GetNode(finishID)

	if startNode == nil || finishNode == nil {
		return nil, ErrNoPathFound
	}

	if blockedNodes[startID] || blockedNodes[finishID] {
		return nil, ErrNoPathFound
	}

	openSet := make(priorityQueue, 0)
	gScores := make(map[string]float64)
	cameFrom := make(map[string]*aStarNode)

	start := &aStarNode{
		nodeID: startID,
		gScore: 0,
		fScore: r.heuristic(startNode, finishNode),
	}
	heap.Push(&openSet, start)
	gScores[startID] = 0

	for openSet.Len() > 0 {
		current := heap.Pop(&openSet).(*aStarNode)

		if current.nodeID == finishID {
			return r.reconstructPath(cameFrom, current, finishNode), nil
		}

		for _, edge := range r.graph.GetEdges(current.nodeID) {
			if blockedNodes[edge.To] {
				continue
			}
			if edge.Severity > maxSeverity {
				continue
			}

			neighborNode := r.graph.GetNode(edge.To)
			if neighborNode == nil {
				continue
			}

			tentativeG := current.gScore + edge.Distance
			if g, ok := gScores[edge.To]; !ok || tentativeG < g {
				cameFrom[edge.To] = current
				gScores[edge.To] = tentativeG

				neighbor := &aStarNode{
					nodeID: edge.To,
					gScore: tentativeG,
					fScore: tentativeG + r.heuristic(neighborNode, finishNode),
				}
				heap.Push(&openSet, neighbor)
			}
		}
	}

	return nil, ErrNoPathFound
}

func (r *Router) heuristic(a, b *entity.Node) float64 {
	coordA := valueobject.Coordinates{Latitude: a.Latitude, Longitude: a.Longitude}
	coordB := valueobject.Coordinates{Latitude: b.Latitude, Longitude: b.Longitude}
	return coordA.HaversineDistance(coordB)
}

func (r *Router) reconstructPath(cameFrom map[string]*aStarNode, end *aStarNode, finishNode *entity.Node) *entity.Route {
	route := entity.NewRoute()
	var path []string
	for id := end.nodeID; ; {
		path = append(path, id)
		prev := cameFrom[id]
		if prev == nil {
			break
		}
		id = prev.nodeID
	}

	for i := len(path) - 1; i >= 0; i-- {
		node := r.graph.GetNode(path[i])
		if node != nil {
			var dist float64
			var sev int
			if i < len(path)-1 {
				nextNode := r.graph.GetNode(path[i+1])
				if nextNode != nil {
					coord1 := valueobject.Coordinates{Latitude: node.Latitude, Longitude: node.Longitude}
					coord2 := valueobject.Coordinates{Latitude: nextNode.Latitude, Longitude: nextNode.Longitude}
					dist = coord1.HaversineDistance(coord2)
					for _, edge := range r.graph.GetEdges(path[i]) {
						if edge.To == path[i+1] {
							sev = edge.Severity
							break
						}
					}
				}
			}
			route.AddNode(node, dist, sev)
		}
	}

	return route
}