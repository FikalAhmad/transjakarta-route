package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
)

type FullRow struct {
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
	f, _ := os.Open("data/raw/combine_tije_opendata_cleaned.csv")
	defer f.Close()

	r := csv.NewReader(f)
	r.Comma = ';'
	r.Read()

	totalRows := 0
	duplicateRows := 0
	seenExact := make(map[string]bool)

	routes := make(map[string]string)
	routeColors := make(map[string]string)
	trips := make(map[string]string) // trip_id -> route_id
	tripDir := make(map[string]string)
	tripShapes := make(map[string]string)
	services := make(map[string]int)
	stops := make(map[string]struct{ Lat, Lon float64 })
	shapes := make(map[string]int)

	type StopTimeKey struct {
		TripID   string
		StopSeq  int
		StopName string
	}
	tripStopTimes := make(map[StopTimeKey]bool)

	var allRows []FullRow

	for {
		rec, err := r.Read()
		if err == io.EOF {
			break
		}
		totalRows++

		lineKey := fmt.Sprintf("%s|%s|%s|%s|%s|%s|%s|%s|%s|%s|%s|%s|%s",
			rec[0], rec[1], rec[2], rec[3], rec[4], rec[5], rec[6], rec[7], rec[8], rec[9], rec[10], rec[11], rec[12])
		if seenExact[lineKey] {
			duplicateRows++
		}
		seenExact[lineKey] = true

		stopSeq, _ := strconv.Atoi(rec[6])
		lat, _ := strconv.ParseFloat(rec[8], 64)
		lon, _ := strconv.ParseFloat(rec[9], 64)
		shapeSeq, _ := strconv.Atoi(rec[11])

		row := FullRow{
			RouteID:         rec[0],
			RouteLongName:   rec[1],
			RouteColor:      rec[2],
			TripID:          rec[3],
			ServiceID:       rec[4],
			DirectionID:     rec[5],
			StopSequence:    stopSeq,
			StopName:        rec[7],
			StopLat:         lat,
			StopLon:         lon,
			LocationType:    rec[10],
			ShapePtSequence: shapeSeq,
			ShapeID:         rec[12],
		}
		allRows = append(allRows, row)

		routes[row.RouteID] = row.RouteLongName
		routeColors[row.RouteID] = row.RouteColor
		trips[row.TripID] = row.RouteID
		tripDir[row.TripID] = row.DirectionID
		tripShapes[row.TripID] = row.ShapeID
		services[row.ServiceID]++
		stops[row.StopName] = struct{ Lat, Lon float64 }{row.StopLat, row.StopLon}
		shapes[row.ShapeID]++
		tripStopTimes[StopTimeKey{row.TripID, row.StopSequence, row.StopName}] = true
	}

	fmt.Printf("Total CSV Rows: %d\n", totalRows)
	fmt.Printf("Exact Duplicate Rows: %d\n", duplicateRows)
	fmt.Printf("Unique Exact Rows: %d\n", len(seenExact))
	fmt.Printf("Unique Routes: %d\n", len(routes))
	fmt.Printf("Unique Trips: %d\n", len(trips))
	fmt.Printf("Unique Services: %v\n", services)
	fmt.Printf("Unique Stop Names: %d\n", len(stops))
	fmt.Printf("Unique Shapes: %d\n", len(shapes))
	fmt.Printf("Unique Trip-Stop_Sequence Pairs: %d\n", len(tripStopTimes))

	// Categorization sample
	brtCount := 0
	nonBrtCount := 0
	for rID := range routes {
		// Corridors 1-14 generally BRT
		if _, err := strconv.Atoi(rID); err == nil {
			brtCount++
		} else {
			nonBrtCount++
		}
	}
	fmt.Printf("Routes: Pure Numeric (likely BRT): %d, Non-Numeric: %d\n", brtCount, nonBrtCount)
}
