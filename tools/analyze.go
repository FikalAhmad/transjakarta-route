package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
)

type Row struct {
	RouteID         string
	RouteLongName   string
	RouteColor      string
	TripID          string
	ServiceID       string
	DirectionID     string
	StopSequence    int
	StopName        string
	StopLat         float64
	StopLon         float64
	LocationType    string
	ShapePtSequence int
	ShapeID         string
}

func main() {
	f, err := os.Open("data/raw/combine_tije_opendata_cleaned.csv")
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	reader := csv.NewReader(f)
	reader.Comma = ';'

	header, err := reader.Read()
	if err != nil {
		fmt.Printf("Error reading header: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Header: %v\n", header)

	rowCount := 0
	routes := make(map[string]string)
	trips := make(map[string]bool)
	directions := make(map[string]bool)
	services := make(map[string]bool)
	stops := make(map[string]bool)
	shapes := make(map[string]bool)
	locationTypes := make(map[string]bool)

	minLat, maxLat := 90.0, -90.0
	minLon, maxLon := 180.0, -180.0

	tripStopSeqs := make(map[string][]int)
	tripShapeSeqs := make(map[string][]int)
	tripRouteMap := make(map[string]string)
	tripDirMap := make(map[string]string)
	tripShapeMap := make(map[string]string)

	missingValues := make(map[string]int)

	for {
		rec, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Printf("Error reading row %d: %v\n", rowCount+1, err)
			continue
		}
		rowCount++

		for i, val := range rec {
			if strings.TrimSpace(val) == "" {
				missingValues[header[i]]++
			}
		}

		routeID := rec[0]
		routeName := rec[1]
		tripID := rec[3]
		serviceID := rec[4]
		directionID := rec[5]
		stopSeq, _ := strconv.Atoi(rec[6])
		stopName := rec[7]
		lat, _ := strconv.ParseFloat(rec[8], 64)
		lon, _ := strconv.ParseFloat(rec[9], 64)
		locType := rec[10]
		shapeSeq, _ := strconv.Atoi(rec[11])
		shapeID := rec[12]

		routes[routeID] = routeName
		trips[tripID] = true
		directions[directionID] = true
		services[serviceID] = true
		stops[stopName] = true
		shapes[shapeID] = true
		locationTypes[locType] = true

		if lat != 0 && lon != 0 {
			if lat < minLat {
				minLat = lat
			}
			if lat > maxLat {
				maxLat = lat
			}
			if lon < minLon {
				minLon = lon
			}
			if lon > maxLon {
				maxLon = lon
			}
		}

		tripStopSeqs[tripID] = append(tripStopSeqs[tripID], stopSeq)
		tripShapeSeqs[tripID] = append(tripShapeSeqs[tripID], shapeSeq)
		tripRouteMap[tripID] = routeID
		tripDirMap[tripID] = directionID
		tripShapeMap[tripID] = shapeID
	}

	fmt.Println("--- ANALYSIS SUMMARY ---")
	fmt.Printf("Total rows: %d\n", rowCount)
	fmt.Printf("Unique routes: %d\n", len(routes))
	fmt.Printf("Unique trips: %d\n", len(trips))
	fmt.Printf("Unique directions: %d\n", len(directions))
	fmt.Printf("Unique services: %d\n", len(services))
	fmt.Printf("Unique stop names: %d\n", len(stops))
	fmt.Printf("Unique shapes: %d\n", len(shapes))
	fmt.Printf("Location types: %v\n", locationTypes)
	fmt.Printf("Bounding box: Lat [%.6f, %.6f], Lon [%.6f, %.6f]\n", minLat, maxLat, minLon, maxLon)
	fmt.Printf("Missing values: %v\n", missingValues)

	// Check stop sequences monotonicity per trip
	seqIssues := 0
	for t, seqs := range tripStopSeqs {
		for i := 1; i < len(seqs); i++ {
			if seqs[i] < seqs[i-1] {
				seqIssues++
				fmt.Printf("Trip %s stop sequence anomaly at index %d: %d -> %d\n", t, i, seqs[i-1], seqs[i])
			}
		}
	}
	fmt.Printf("Stop sequence ordering anomalies: %d\n", seqIssues)

	// Print some route samples
	var sampleRoutes []string
	for r, name := range routes {
		sampleRoutes = append(sampleRoutes, fmt.Sprintf("%s: %s", r, name))
	}
	sort.Strings(sampleRoutes)
	fmt.Println("\nSample Routes (first 10):")
	for i := 0; i < int(math.Min(10, float64(len(sampleRoutes)))); i++ {
		fmt.Printf("  %s\n", sampleRoutes[i])
	}
}
