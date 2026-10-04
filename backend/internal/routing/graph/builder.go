package graph

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"

	"transjakarta/internal/domain"
	"transjakarta/internal/geospatial"
	"transjakarta/internal/normalization"
)

type CSVRecord struct {
	RouteID         string
	RouteLongName   string
	RouteColor      string
	TripID          string
	ServiceID       string
	DirectionID     int
	StopSequence    int
	StopName        string
	StopLat         float64
	StopLon         float64
	LocationType    int
	ShapePtSequence int
	ShapeID         string
}

// BuildGraphFromCSV parses the raw CSV, normalizes stops/trips, and constructs the in-memory TransitGraph
func BuildGraphFromCSV(csvPath string) (*TransitGraph, error) {
	f, err := os.Open(csvPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open CSV file %s: %w", csvPath, err)
	}
	defer f.Close()

	reader := csv.NewReader(f)
	reader.Comma = ';'

	// Read header
	_, err = reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV header: %w", err)
	}

	g := NewTransitGraph()

	// Map trip_id -> list of stop times
	type TripStopItem struct {
		Stop     domain.Stop
		Sequence int
	}
	tripStopsMap := make(map[string][]TripStopItem)
	tripMetaMap := make(map[string]struct {
		RouteID     string
		DirectionID int
		ServiceID   string
	})

	seenTripStopSeq := make(map[string]bool)

	rowCount := 0
	for {
		rec, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}
		rowCount++

		routeID := rec[0]
		routeName := rec[1]
		routeColor := rec[2]
		tripID := rec[3]
		serviceID := rec[4]
		directionID, _ := strconv.Atoi(rec[5])
		stopSeq, _ := strconv.Atoi(rec[6])
		stopName := rec[7]
		stopLat, _ := strconv.ParseFloat(rec[8], 64)
		stopLon, _ := strconv.ParseFloat(rec[9], 64)
		shapeSeq, _ := strconv.Atoi(rec[11])
		shapeID := rec[12]

		// Save route
		if _, exists := g.RoutesByID[routeID]; !exists {
			g.RoutesByID[routeID] = domain.Route{
				RouteID:        routeID,
				RouteShortName: routeID,
				RouteLongName:  routeName,
				RouteColor:     "#" + routeColor,
				ServiceType:    normalization.ClassifyService(routeID),
			}
		}

		// Stop ID generation
		stopID := normalization.GenerateStopID(stopName, stopLat, stopLon)
		stopObj := domain.Stop{
			StopID:       stopID,
			StopName:     stopName,
			Latitude:     stopLat,
			Longitude:    stopLon,
			LocationType: 0,
		}

		g.StopsByID[stopID] = stopObj

		// Shape points
		if shapeID != "" {
			g.ShapesByID[shapeID] = append(g.ShapesByID[shapeID], domain.Coordinate{
				Latitude:  stopLat,
				Longitude: stopLon,
			})
		}

		dedupKey := fmt.Sprintf("%s|%d|%s", tripID, stopSeq, stopID)
		if !seenTripStopSeq[dedupKey] {
			seenTripStopSeq[dedupKey] = true
			tripStopsMap[tripID] = append(tripStopsMap[tripID], TripStopItem{
				Stop:     stopObj,
				Sequence: stopSeq,
			})
			tripMetaMap[tripID] = struct {
				RouteID     string
				DirectionID int
				ServiceID   string
			}{
				RouteID:     routeID,
				DirectionID: directionID,
				ServiceID:   serviceID,
			}
		}
		_ = shapeSeq
	}

	// Build Transit Edges along each trip
	transitEdgeCount := 0
	for tripID, stops := range tripStopsMap {
		meta := tripMetaMap[tripID]

		// Sort by stop_sequence
		sort.Slice(stops, func(i, j int) bool {
			return stops[i].Sequence < stops[j].Sequence
		})

		for i := 0; i < len(stops)-1; i++ {
			fromStop := stops[i].Stop
			toStop := stops[i+1].Stop

			fromNodeID := MakeNodeID(fromStop.StopID, meta.RouteID, meta.DirectionID)
			toNodeID := MakeNodeID(toStop.StopID, meta.RouteID, meta.DirectionID)

			g.Nodes[fromNodeID] = fromStop
			g.Nodes[toNodeID] = toStop

			// Register Stop -> Nodes
			registerStopNode(g, fromStop.StopID, fromNodeID)
			registerStopNode(g, toStop.StopID, toNodeID)

			distMeters := geospatial.HaversineDistance(fromStop.Latitude, fromStop.Longitude, toStop.Latitude, toStop.Longitude)
			durMin := geospatial.EstimateTransitMinutes(distMeters, 1)

			edge := Edge{
				FromNode:       fromNodeID,
				ToNode:         toNodeID,
				Type:           EdgeTransit,
				RouteID:        meta.RouteID,
				TripID:         tripID,
				DirectionID:    meta.DirectionID,
				FromStop:       fromStop,
				ToStop:         toStop,
				DistanceMeters: distMeters,
				DurationMin:    durMin,
			}
			g.Adjacency[fromNodeID] = append(g.Adjacency[fromNodeID], edge)
			transitEdgeCount++
		}
	}

	// Build Transfer Edges between different transit lines at the same physical stop or nearby stops (<= 80 meters)
	transferEdgeCount := 0

	// 1. Same physical stop transfers
	for stopID, nodes := range g.StopToNodes {
		stopObj := g.StopsByID[stopID]
		for i := 0; i < len(nodes); i++ {
			for j := 0; j < len(nodes); j++ {
				if i != j {
					// 2 minutes transfer buffer
					edge := Edge{
						FromNode:       nodes[i],
						ToNode:         nodes[j],
						Type:           EdgeTransfer,
						FromStop:       stopObj,
						ToStop:         stopObj,
						DistanceMeters: 20.0,
						DurationMin:    2.0,
					}
					g.Adjacency[nodes[i]] = append(g.Adjacency[nodes[i]], edge)
					transferEdgeCount++
				}
			}
		}
	}

	// 2. Cross-stop transfers within 80m radius (e.g. across the street)
	allStopIDs := make([]string, 0, len(g.StopsByID))
	for sID := range g.StopsByID {
		allStopIDs = append(allStopIDs, sID)
	}

	for i := 0; i < len(allStopIDs); i++ {
		s1 := g.StopsByID[allStopIDs[i]]
		for j := i + 1; j < len(allStopIDs); j++ {
			s2 := g.StopsByID[allStopIDs[j]]
			d := geospatial.HaversineDistance(s1.Latitude, s1.Longitude, s2.Latitude, s2.Longitude)
			if d <= 80.0 {
				dur := geospatial.EstimateWalkMinutes(d) + 1.5 // walk time + 1.5m buffer
				for _, n1 := range g.StopToNodes[s1.StopID] {
					for _, n2 := range g.StopToNodes[s2.StopID] {
						// Forward edge
						g.Adjacency[n1] = append(g.Adjacency[n1], Edge{
							FromNode:       n1,
							ToNode:         n2,
							Type:           EdgeTransfer,
							FromStop:       s1,
							ToStop:         s2,
							DistanceMeters: d,
							DurationMin:    dur,
						})
						// Reverse edge
						g.Adjacency[n2] = append(g.Adjacency[n2], Edge{
							FromNode:       n2,
							ToNode:         n1,
							Type:           EdgeTransfer,
							FromStop:       s2,
							ToStop:         s1,
							DistanceMeters: d,
							DurationMin:    dur,
						})
						transferEdgeCount += 2
					}
				}
			}
		}
	}

	fmt.Println("--- GRAPH BUILD COMPLETE ---")
	fmt.Printf("Total Stops: %d\n", len(g.StopsByID))
	fmt.Printf("Total Graph Nodes (Stop x Route x Dir): %d\n", len(g.Nodes))
	fmt.Printf("Transit Edges: %d\n", transitEdgeCount)
	fmt.Printf("Transfer Edges: %d\n", transferEdgeCount)
	fmt.Printf("Total Unique Routes: %d\n", len(g.RoutesByID))

	return g, nil
}

func registerStopNode(g *TransitGraph, stopID string, nodeID NodeID) {
	existing := g.StopToNodes[stopID]
	for _, n := range existing {
		if n == nodeID {
			return
		}
	}
	g.StopToNodes[stopID] = append(existing, nodeID)
}
