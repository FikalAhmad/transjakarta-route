package graph_test

import (
	"testing"

	"transjakarta/internal/domain"
	"transjakarta/internal/geospatial"
	"transjakarta/internal/normalization"
	"transjakarta/internal/routing/graph"
	"transjakarta/internal/validation"
)

func TestNormalizationAndClassification(t *testing.T) {
	// Test stop ID determinism
	id1 := normalization.GenerateStopID("Harmoni Central Busway", -6.162764, 106.819961)
	id2 := normalization.GenerateStopID("Harmoni Central Busway", -6.162764, 106.819961)
	if id1 != id2 {
		t.Fatalf("Expected deterministic stop IDs, got %s and %s", id1, id2)
	}

	// Test service classification
	if normalization.ClassifyService("1") != domain.ServiceBRT {
		t.Errorf("Expected route 1 to be BRT")
	}
	if normalization.ClassifyService("JAK.26") != domain.ServiceMikrotrans {
		t.Errorf("Expected JAK.26 to be Mikrotrans")
	}
	if normalization.ClassifyService("6H") != domain.ServiceNonBRT {
		t.Errorf("Expected 6H to be NON_BRT")
	}
}

func TestHaversineDistance(t *testing.T) {
	// Blok M to Monas is ~6-7 km
	d := geospatial.HaversineDistance(-6.2433, 106.8017, -6.1762, 106.8229)
	if d < 6000 || d > 9000 {
		t.Fatalf("Haversine distance calculation out of realistic range: %.2f m", d)
	}
}

func TestRoutingSearchFixture(t *testing.T) {
	g := graph.NewTransitGraph()

	// Add fixture routes
	g.RoutesByID["1"] = domain.Route{RouteID: "1", RouteLongName: "Blok M - Kota", ServiceType: domain.ServiceBRT}
	g.RoutesByID["2"] = domain.Route{RouteID: "2", RouteLongName: "Pulo Gadung - Harmoni", ServiceType: domain.ServiceBRT}

	sA := domain.Stop{StopID: "STP-A", StopName: "Stop A", Latitude: -6.24, Longitude: 106.80}
	sB := domain.Stop{StopID: "STP-B", StopName: "Stop B (Transfer)", Latitude: -6.20, Longitude: 106.82}
	sC := domain.Stop{StopID: "STP-C", StopName: "Stop C", Latitude: -6.17, Longitude: 106.83}

	g.StopsByID["STP-A"] = sA
	g.StopsByID["STP-B"] = sB
	g.StopsByID["STP-C"] = sC

	nA1 := graph.MakeNodeID("STP-A", "1", 0)
	nB1 := graph.MakeNodeID("STP-B", "1", 0)
	nB2 := graph.MakeNodeID("STP-B", "2", 0)
	nC2 := graph.MakeNodeID("STP-C", "2", 0)

	g.Nodes[nA1] = sA
	g.Nodes[nB1] = sB
	g.Nodes[nB2] = sB
	g.Nodes[nC2] = sC

	g.StopToNodes["STP-A"] = []graph.NodeID{nA1}
	g.StopToNodes["STP-B"] = []graph.NodeID{nB1, nB2}
	g.StopToNodes["STP-C"] = []graph.NodeID{nC2}

	// Transit edges
	g.Adjacency[nA1] = append(g.Adjacency[nA1], graph.Edge{
		FromNode: nA1, ToNode: nB1, Type: graph.EdgeTransit, RouteID: "1", FromStop: sA, ToStop: sB, DistanceMeters: 4000, DurationMin: 12,
	})
	// Transfer edge at Stop B
	g.Adjacency[nB1] = append(g.Adjacency[nB1], graph.Edge{
		FromNode: nB1, ToNode: nB2, Type: graph.EdgeTransfer, FromStop: sB, ToStop: sB, DistanceMeters: 20, DurationMin: 2,
	})
	// Transit edge on route 2
	g.Adjacency[nB2] = append(g.Adjacency[nB2], graph.Edge{
		FromNode: nB2, ToNode: nC2, Type: graph.EdgeTransit, RouteID: "2", FromStop: sB, ToStop: sC, DistanceMeters: 3500, DurationMin: 10,
	})

	engine := graph.NewRoutingEngine(g)

	// Search from near Stop A to near Stop C
	req := domain.RouteSearchRequest{
		Origin:      domain.Coordinate{Latitude: -6.241, Longitude: 106.801},
		Destination: domain.Coordinate{Latitude: -6.169, Longitude: 106.831},
		Preference:  domain.PreferenceBalanced,
	}

	results, err := engine.SearchRoutes(req)
	if err != nil {
		t.Fatalf("SearchRoutes error: %v", err)
	}

	if len(results) == 0 {
		t.Fatalf("Expected at least 1 route result")
	}

	plan := results[0]
	if plan.TransferCount != 1 {
		t.Errorf("Expected 1 transfer, got %d", plan.TransferCount)
	}

	if err := validation.ValidateRoutePlan(&plan); err != nil {
		t.Errorf("Plan validation failed: %v", err)
	}
}
