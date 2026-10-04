package graph

import (
	"transjakarta/internal/domain"
)

type RoutingWeights struct {
	WalkingWeight   float64
	TransitWeight   float64
	TransferPenalty float64 // cost equivalent penalty in minutes
	StopPenalty     float64 // penalty per stop
}

func DefaultWeights(pref domain.Preference) RoutingWeights {
	switch pref {
	case domain.PreferenceFastest:
		return RoutingWeights{
			WalkingWeight:   1.2,
			TransitWeight:   1.0,
			TransferPenalty: 6.0,
			StopPenalty:     0.5,
		}
	case domain.PreferenceFewestTransfers:
		return RoutingWeights{
			WalkingWeight:   1.0,
			TransitWeight:   1.0,
			TransferPenalty: 25.0, // strong penalty to avoid transfers
			StopPenalty:     0.2,
		}
	case domain.PreferenceLeastWalking:
		return RoutingWeights{
			WalkingWeight:   3.5, // strong walking penalty
			TransitWeight:   0.8,
			TransferPenalty: 4.0,
			StopPenalty:     0.5,
		}
	default: // BALANCED
		return RoutingWeights{
			WalkingWeight:   1.5,
			TransitWeight:   1.0,
			TransferPenalty: 8.0,
			StopPenalty:     0.5,
		}
	}
}

type EdgeType string

const (
	EdgeTransit  EdgeType = "TRANSIT"
	EdgeTransfer EdgeType = "TRANSFER"
	EdgeWalk     EdgeType = "WALK"
)

// GraphNode represents a transit state: (StopID, RouteID, DirectionID)
type NodeID string

type Edge struct {
	FromNode       NodeID
	ToNode         NodeID
	Type           EdgeType
	RouteID        string
	TripID         string
	DirectionID    int
	FromStop       domain.Stop
	ToStop         domain.Stop
	DistanceMeters float64
	DurationMin    float64
	Weight         float64
}

type TransitGraph struct {
	Nodes    map[NodeID]domain.Stop
	Adjacency map[NodeID][]Edge
	// Spatial index for stops: StopID -> Stop
	StopsByID map[string]domain.Stop
	// Stop to Nodes mapping: StopID -> []NodeID
	StopToNodes map[string][]NodeID
	// Routes metadata
	RoutesByID map[string]domain.Route
	// Shapes mapping: ShapeID -> []Coordinate
	ShapesByID map[string][]domain.Coordinate
}

func NewTransitGraph() *TransitGraph {
	return &TransitGraph{
		Nodes:       make(map[NodeID]domain.Stop),
		Adjacency:   make(map[NodeID][]Edge),
		StopsByID:   make(map[string]domain.Stop),
		StopToNodes: make(map[string][]NodeID),
		RoutesByID:  make(map[string]domain.Route),
		ShapesByID:  make(map[string][]domain.Coordinate),
	}
}

func MakeNodeID(stopID, routeID string, directionID int) NodeID {
	return NodeID(stopID + "@" + routeID + "#" + string(rune('0'+directionID)))
}
