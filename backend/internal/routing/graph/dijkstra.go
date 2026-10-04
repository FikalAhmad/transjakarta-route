package graph

import (
	"container/heap"
	"fmt"
	"math"
	"sort"
	"strings"

	"transjakarta/internal/domain"
	"transjakarta/internal/geospatial"
)

type PQItem struct {
	Node     NodeID
	Priority float64
	Index    int
}

type PriorityQueue []*PQItem

func (pq PriorityQueue) Len() int           { return len(pq) }
func (pq PriorityQueue) Less(i, j int) bool { return pq[i].Priority < pq[j].Priority }
func (pq PriorityQueue) Swap(i, j int)      { pq[i], pq[j] = pq[j], pq[i]; pq[i].Index = i; pq[j].Index = j }
func (pq *PriorityQueue) Push(x interface{}) {
	n := len(*pq)
	item := x.(*PQItem)
	item.Index = n
	*pq = append(*pq, item)
}
func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.Index = -1
	*pq = old[0 : n-1]
	return item
}

type RoutingEngine struct {
	Graph *TransitGraph
}

func NewRoutingEngine(g *TransitGraph) *RoutingEngine {
	return &RoutingEngine{Graph: g}
}

// FindCandidateStops returns the N nearest physical stops within maxRadiusMeters
func (e *RoutingEngine) FindCandidateStops(lat, lon float64, limit int, maxRadiusMeters float64) []domain.Stop {
	type candidate struct {
		stop domain.Stop
		dist float64
	}
	var list []candidate
	for _, s := range e.Graph.StopsByID {
		d := geospatial.HaversineDistance(lat, lon, s.Latitude, s.Longitude)
		if d <= maxRadiusMeters {
			sCopy := s
			sCopy.DistanceMeters = d
			list = append(list, candidate{stop: sCopy, dist: d})
		}
	}

	sort.Slice(list, func(i, j int) bool {
		return list[i].dist < list[j].dist
	})

	var result []domain.Stop
	for i := 0; i < len(list) && i < limit; i++ {
		result = append(result, list[i].stop)
	}
	return result
}

// SearchRoutes finds candidate transit routes from origin to destination
func (e *RoutingEngine) SearchRoutes(req domain.RouteSearchRequest) ([]domain.RoutePlanResult, error) {
	weights := DefaultWeights(req.Preference)

	// 1. Find candidate origin & destination stops (within 1.5 km)
	originStops := e.FindCandidateStops(req.Origin.Latitude, req.Origin.Longitude, 5, 1500.0)
	if len(originStops) == 0 {
		return nil, fmt.Errorf("no transit stops found near origin within 1.5km")
	}

	destStops := e.FindCandidateStops(req.Destination.Latitude, req.Destination.Longitude, 5, 1500.0)
	if len(destStops) == 0 {
		return nil, fmt.Errorf("no transit stops found near destination within 1.5km")
	}

	destStopSet := make(map[string]domain.Stop)
	for _, ds := range destStops {
		destStopSet[ds.StopID] = ds
	}

	// 2. Setup Dijkstra search with virtual source
	const virtualSource NodeID = "__VIRTUAL_SOURCE__"
	const virtualTarget NodeID = "__VIRTUAL_TARGET__"

	dist := make(map[NodeID]float64)
	prevEdge := make(map[NodeID]*Edge)
	prevNode := make(map[NodeID]NodeID)

	pq := make(PriorityQueue, 0)
	heap.Init(&pq)

	dist[virtualSource] = 0
	heap.Push(&pq, &PQItem{Node: virtualSource, Priority: 0})

	// Add virtual walk edges from source to origin stops transit nodes
	for _, os := range originStops {
		walkDist := os.DistanceMeters
		walkDur := geospatial.EstimateWalkMinutes(walkDist)
		walkCost := walkDur * weights.WalkingWeight

		for _, nID := range e.Graph.StopToNodes[os.StopID] {
			edge := &Edge{
				FromNode:       virtualSource,
				ToNode:         nID,
				Type:           EdgeWalk,
				FromStop:       domain.Stop{StopName: "Origin Location", Latitude: req.Origin.Latitude, Longitude: req.Origin.Longitude},
				ToStop:         os,
				DistanceMeters: walkDist,
				DurationMin:    walkDur,
				Weight:         walkCost,
			}
			cost := edge.Weight
			if d, ok := dist[nID]; !ok || cost < d {
				dist[nID] = cost
				prevEdge[nID] = edge
				prevNode[nID] = virtualSource
				heap.Push(&pq, &PQItem{Node: nID, Priority: cost})
			}
		}
	}

	// Run Dijkstra
	targetReachedNodes := make(map[NodeID]float64)

	for pq.Len() > 0 {
		curr := heap.Pop(&pq).(*PQItem)
		u := curr.Node
		currentDist := curr.Priority

		if currentDist > dist[u] {
			continue
		}

		// Check if current node is at one of destination stops
		stopID := extractStopID(u)
		if dStop, isDest := destStopSet[stopID]; isDest {
			walkDestDist := dStop.DistanceMeters
			walkDestDur := geospatial.EstimateWalkMinutes(walkDestDist)
			walkDestCost := walkDestDur * weights.WalkingWeight
			totalCost := currentDist + walkDestCost

			targetReachedNodes[u] = totalCost
		}

		// Traverse neighbors
		for _, edge := range e.Graph.Adjacency[u] {
			v := edge.ToNode
			var edgeCost float64

			switch edge.Type {
			case EdgeTransit:
				edgeCost = (edge.DurationMin * weights.TransitWeight) + weights.StopPenalty
			case EdgeTransfer:
				edgeCost = (edge.DurationMin * weights.WalkingWeight) + weights.TransferPenalty
			default:
				edgeCost = edge.DurationMin * weights.WalkingWeight
			}

			newDist := currentDist + edgeCost
			if oldDist, exists := dist[v]; !exists || newDist < oldDist {
				dist[v] = newDist
				eCopy := edge
				prevEdge[v] = &eCopy
				prevNode[v] = u
				heap.Push(&pq, &PQItem{Node: v, Priority: newDist})
			}
		}
	}

	if len(targetReachedNodes) == 0 {
		return nil, fmt.Errorf("no transit route available between origin and destination")
	}

	// Rank destination candidate terminal nodes
	type scoredNode struct {
		node NodeID
		cost float64
	}
	var scored []scoredNode
	for n, c := range targetReachedNodes {
		scored = append(scored, scoredNode{node: n, cost: c})
	}
	sort.Slice(scored, func(i, j int) bool {
		return scored[i].cost < scored[j].cost
	})

	var candidateRoutes []domain.RoutePlanResult
	seenSummaries := make(map[string]bool)

	for _, sn := range scored {
		plan := e.reconstructPath(sn.node, prevEdge, prevNode, req)
		if plan != nil && !seenSummaries[plan.Summary] {
			seenSummaries[plan.Summary] = true
			candidateRoutes = append(candidateRoutes, *plan)
			if len(candidateRoutes) >= 3 {
				break
			}
		}
	}

	return candidateRoutes, nil
}

func extractStopID(nodeID NodeID) string {
	parts := strings.Split(string(nodeID), "@")
	if len(parts) > 0 {
		return parts[0]
	}
	return string(nodeID)
}

func (e *RoutingEngine) reconstructPath(destNode NodeID, prevEdge map[NodeID]*Edge, prevNode map[NodeID]NodeID, req domain.RouteSearchRequest) *domain.RoutePlanResult {
	var edges []*Edge
	curr := destNode

	for curr != "__VIRTUAL_SOURCE__" {
		edge, ok := prevEdge[curr]
		if !ok || edge == nil {
			break
		}
		edges = append([]*Edge{edge}, edges...)
		curr = prevNode[curr]
	}

	if len(edges) == 0 {
		return nil
	}

	// Add final walk to destination
	lastTransitStop := edges[len(edges)-1].ToStop
	finalWalkDist := geospatial.HaversineDistance(lastTransitStop.Latitude, lastTransitStop.Longitude, req.Destination.Latitude, req.Destination.Longitude)
	finalWalkDur := geospatial.EstimateWalkMinutes(finalWalkDist)

	destStopObj := domain.Stop{StopName: "Destination Location", Latitude: req.Destination.Latitude, Longitude: req.Destination.Longitude}

	finalWalkEdge := &Edge{
		Type:           EdgeWalk,
		FromStop:       lastTransitStop,
		ToStop:         destStopObj,
		DistanceMeters: finalWalkDist,
		DurationMin:    finalWalkDur,
	}
	edges = append(edges, finalWalkEdge)

	// Combine contiguous transit edges on the same trip/route into a single Segment
	var segments []domain.Segment
	var fullCoords [][]float64

	var currentTransitSegment *domain.Segment
	var totalWalkDist float64
	var totalTransitDist float64
	var totalDuration float64
	var transferCount int

	for _, edge := range edges {
		switch edge.Type {
		case EdgeWalk:
			if currentTransitSegment != nil {
				segments = append(segments, *currentTransitSegment)
				currentTransitSegment = nil
			}
			fromCopy := edge.FromStop
			toCopy := edge.ToStop
			walkSeg := domain.Segment{
				Type:            domain.SegmentTypeWalk,
				Description:     fmt.Sprintf("Walk %.0f m to %s", edge.DistanceMeters, edge.ToStop.StopName),
				DistanceMeters:  edge.DistanceMeters,
				DurationMinutes: edge.DurationMin,
				From:            &fromCopy,
				To:              &toCopy,
				Geometry: &domain.GeoJSONGeometry{
					Type: "LineString",
					Coordinates: [][]float64{
						{edge.FromStop.Longitude, edge.FromStop.Latitude},
						{edge.ToStop.Longitude, edge.ToStop.Latitude},
					},
				},
			}
			segments = append(segments, walkSeg)
			totalWalkDist += edge.DistanceMeters
			totalDuration += edge.DurationMin
			fullCoords = append(fullCoords, []float64{edge.FromStop.Longitude, edge.FromStop.Latitude})
			fullCoords = append(fullCoords, []float64{edge.ToStop.Longitude, edge.ToStop.Latitude})

		case EdgeTransfer:
			if currentTransitSegment != nil {
				segments = append(segments, *currentTransitSegment)
				currentTransitSegment = nil
			}
			transferCount++
			atCopy := edge.FromStop
			transferSeg := domain.Segment{
				Type:            domain.SegmentTypeTransfer,
				Description:     fmt.Sprintf("Transfer at %s (walk ~%.0f m)", edge.FromStop.StopName, edge.DistanceMeters),
				DistanceMeters:  edge.DistanceMeters,
				DurationMinutes: edge.DurationMin,
				AtStop:          &atCopy,
			}
			segments = append(segments, transferSeg)
			totalWalkDist += edge.DistanceMeters
			totalDuration += edge.DurationMin

		case EdgeTransit:
			routeInfo := e.Graph.RoutesByID[edge.RouteID]
			if currentTransitSegment == nil || currentTransitSegment.RouteID != edge.RouteID {
				if currentTransitSegment != nil {
					segments = append(segments, *currentTransitSegment)
				}
				bStop := edge.FromStop
				aStop := edge.ToStop
				currentTransitSegment = &domain.Segment{
					Type:              domain.SegmentTypeTransit,
					RouteID:           edge.RouteID,
					RouteName:         routeInfo.RouteLongName,
					RouteColor:        routeInfo.RouteColor,
					ServiceType:       routeInfo.ServiceType,
					DirectionID:       edge.DirectionID,
					BoardingStop:      &bStop,
					AlightingStop:     &aStop,
					IntermediateStops: []domain.Stop{edge.FromStop, edge.ToStop},
					DistanceMeters:    edge.DistanceMeters,
					DurationMinutes:   edge.DurationMin,
					Description:       fmt.Sprintf("Take %s (%s) from %s to %s", edge.RouteID, routeInfo.RouteLongName, edge.FromStop.StopName, edge.ToStop.StopName),
				}
			} else {
				aStop := edge.ToStop
				currentTransitSegment.AlightingStop = &aStop
				currentTransitSegment.IntermediateStops = append(currentTransitSegment.IntermediateStops, edge.ToStop)
				currentTransitSegment.DistanceMeters += edge.DistanceMeters
				currentTransitSegment.DurationMinutes += edge.DurationMin
				currentTransitSegment.Description = fmt.Sprintf("Take %s (%s) from %s to %s", currentTransitSegment.RouteID, routeInfo.RouteLongName, currentTransitSegment.BoardingStop.StopName, edge.ToStop.StopName)
			}
			totalTransitDist += edge.DistanceMeters
			totalDuration += edge.DurationMin
			fullCoords = append(fullCoords, []float64{edge.FromStop.Longitude, edge.FromStop.Latitude})
			fullCoords = append(fullCoords, []float64{edge.ToStop.Longitude, edge.ToStop.Latitude})
		}
	}

	if currentTransitSegment != nil {
		segments = append(segments, *currentTransitSegment)
	}

	// Create summary
	var transitParts []string
	for _, seg := range segments {
		if seg.Type == domain.SegmentTypeTransit {
			transitParts = append(transitParts, seg.RouteID)
		}
	}
	summary := strings.Join(transitParts, " ➔ ")
	if summary == "" {
		summary = "Direct Walking"
	}

	// Standard TransJakarta Flat Fare = Rp 3.500 (Free transfer inside BRT / Rp 3.500 flat)
	fare := 3500
	if transferCount > 0 {
		fare = 3500 // Integration flat rate
	}

	plan := &domain.RoutePlanResult{
		RouteID:               fmt.Sprintf("ROUTE-TJ-%s", strings.ReplaceAll(summary, " ➔ ", "-")),
		Summary:               summary,
		TotalDurationMinutes:  math.Round(totalDuration),
		TotalDistanceMeters:   math.Round(totalWalkDist + totalTransitDist),
		WalkingDistanceMeters: math.Round(totalWalkDist),
		TransitDistanceMeters: math.Round(totalTransitDist),
		TransferCount:         transferCount,
		TotalFare:             fare,
		Segments:              segments,
		FullGeometry: &domain.GeoJSONGeometry{
			Type:        "LineString",
			Coordinates: fullCoords,
		},
	}

	return plan
}
