package service

import (
	"container/heap"
	"errors"
	"log/slog"

	"github.com/accessible-path/route-service/internal/domain/entity"
	"github.com/accessible-path/route-service/internal/domain/valueobject"
	"github.com/accessible-path/route-service/internal/ports/out"
)

var (
	ErrNoPathFound         = errors.New("no path found")
	ErrInvalidProfile      = errors.New("invalid mobility profile")
	ErrStartOutOfCoverage  = errors.New("start point is too far from mapped roads (outside map coverage)")
	ErrFinishOutOfCoverage = errors.New("finish point is too far from mapped roads (outside map coverage)")
	ErrSearchLimitExceeded = errors.New("route search exceeded its limits; try a shorter distance or a different profile")
)

// maxSnapDistanceMeters is how far a requested point may be from the nearest
// graph node before the route is rejected with a coverage error.
const maxSnapDistanceMeters = 3000.0

// maxSearchIterations caps A* work so a pathological request cannot hang the
// service. With ~500k nodes and dense edges this is far beyond any real path.
const maxSearchIterations = 2000000

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

	startNodeID, startDist, startFound := r.findNearestNode(start)
	finishNodeID, finishDist, finishFound := r.findNearestNode(finish)

	if !startFound || startDist > maxSnapDistanceMeters {
		return nil, ErrStartOutOfCoverage
	}
	if !finishFound || finishDist > maxSnapDistanceMeters {
		return nil, ErrFinishOutOfCoverage
	}

	barriers, err := r.barrierRepo.GetBarriersInBounds(
		start.Latitude, start.Longitude,
		finish.Latitude, finish.Longitude,
	)
	if err != nil {
		// A barrier-service outage must not take down routing entirely:
		// log and build the route without barrier blocks.
		slog.Warn("barrier fetch failed, building route without barriers", "error", err)
		barriers = nil
	}

	blockedNodes := make(map[int64]bool)
	for _, b := range barriers {
		if b.Severity <= maxSeverity {
			nodeID, dist, found := r.findNearestNode(valueobject.Coordinates{Latitude: b.Latitude, Longitude: b.Longitude})
			if found && dist <= maxSnapDistanceMeters {
				blockedNodes[nodeID] = true
			}
		}
	}

	return r.aStar(startNodeID, finishNodeID, blockedNodes, maxSeverity)
}

func (r *Router) findNearestNode(coord valueobject.Coordinates) (int64, float64, bool) {
	return r.graph.FindNearest(coord.Latitude, coord.Longitude)
}

type aStarNode struct {
	nodeID int64
	gScore float64
	fScore float64
	index  int
}

type priorityQueue []*aStarNode

func (pq priorityQueue) Len() int           { return len(pq) }
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

func (r *Router) aStar(startID, finishID int64, blockedNodes map[int64]bool, maxSeverity int) (*entity.Route, error) {
	startNode, startOK := r.graph.GetNode(startID)
	finishNode, finishOK := r.graph.GetNode(finishID)

	if !startOK || !finishOK {
		return nil, ErrNoPathFound
	}

	if blockedNodes[startID] || blockedNodes[finishID] {
		return nil, ErrNoPathFound
	}

	openSet := make(priorityQueue, 0)
	gScores := make(map[int64]float64)
	cameFrom := make(map[int64]*aStarNode)

	start := &aStarNode{
		nodeID: startID,
		gScore: 0,
		fScore: r.heuristic(startNode, finishNode),
	}
	heap.Push(&openSet, start)
	gScores[startID] = 0

	for iterations := 0; openSet.Len() > 0; iterations++ {
		current := heap.Pop(&openSet).(*aStarNode)

		// The node may have been pushed several times with improving gScores;
		// entries with a stale (outdated) gScore must be skipped, not re-expanded.
		if gScores[current.nodeID] < current.gScore {
			continue
		}

		if iterations > maxSearchIterations {
			return nil, ErrSearchLimitExceeded
		}

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

			neighborNode, ok := r.graph.GetNode(edge.To)
			if !ok {
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

func (r *Router) heuristic(a, b entity.Node) float64 {
	coordA := valueobject.Coordinates{Latitude: a.Latitude, Longitude: a.Longitude}
	coordB := valueobject.Coordinates{Latitude: b.Latitude, Longitude: b.Longitude}
	return coordA.HaversineDistance(coordB)
}

func (r *Router) reconstructPath(cameFrom map[int64]*aStarNode, end *aStarNode, finishNode entity.Node) *entity.Route {
	route := entity.NewRoute()
	var path []int64
	for id := end.nodeID; ; {
		path = append(path, id)
		prev := cameFrom[id]
		if prev == nil {
			break
		}
		id = prev.nodeID
	}

	for i := len(path) - 1; i >= 0; i-- {
		node, ok := r.graph.GetNode(path[i])
		if !ok {
			continue
		}
		var dist float64
		var sev int
		if i < len(path)-1 {
			nextNode, ok := r.graph.GetNode(path[i+1])
			if !ok {
				continue
			}
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
		route.AddNode(&entity.Node{ID: node.ID, Latitude: node.Latitude, Longitude: node.Longitude}, dist, sev)
	}

	return route
}
