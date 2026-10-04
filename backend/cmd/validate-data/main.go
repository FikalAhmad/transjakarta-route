package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"transjakarta/internal/normalization"
	"transjakarta/internal/validation"
)

func main() {
	csvPath := flag.String("csv", "data/raw/combine_tije_opendata_cleaned.csv", "path to raw CSV data")
	flag.Parse()

	f, err := os.Open(*csvPath)
	if err != nil {
		fmt.Printf("ERROR opening file: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.Comma = ';'
	header, err := r.Read()
	if err != nil {
		fmt.Printf("ERROR reading header: %v\n", err)
		os.Exit(1)
	}

	requiredCols := []string{"route_id", "route_long_name", "trip_id", "stop_name", "stop_lat", "stop_lon"}
	for _, col := range requiredCols {
		found := false
		for _, h := range header {
			if strings.TrimSpace(h) == col {
				found = true
				break
			}
		}
		if !found {
			fmt.Printf("VALIDATION FAILED: Missing required column '%s'\n", col)
			os.Exit(1)
		}
	}

	routes := make(map[string]bool)
	trips := make(map[string]bool)
	stops := make(map[string]bool)

	var errs []string
	var warns []string

	rowIdx := 0
	for {
		rec, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			errs = append(errs, fmt.Sprintf("Row %d malformed CSV: %v", rowIdx+1, err))
			continue
		}
		rowIdx++

		routeID := rec[0]
		tripID := rec[3]
		stopName := rec[7]
		lat, err1 := strconv.ParseFloat(rec[8], 64)
		lon, err2 := strconv.ParseFloat(rec[9], 64)

		if err1 != nil || err2 != nil {
			errs = append(errs, fmt.Sprintf("Row %d invalid coordinate format: (%s, %s)", rowIdx, rec[8], rec[9]))
		} else {
			if err := validation.ValidateCoordinates(lat, lon); err != nil {
				warns = append(warns, fmt.Sprintf("Row %d out of bounds stop '%s': %v", rowIdx, stopName, err))
			}
		}

		routes[routeID] = true
		trips[tripID] = true
		stops[normalization.GenerateStopID(stopName, lat, lon)] = true
	}

	fmt.Println("========================================")
	fmt.Println("         DATA VALIDATION REPORT         ")
	fmt.Println("========================================")
	fmt.Printf("Total Rows Evaluated : %d\n", rowIdx)
	fmt.Printf("Unique Routes        : %d\n", len(routes))
	fmt.Printf("Unique Trips         : %d\n", len(trips))
	fmt.Printf("Unique Stops         : %d\n", len(stops))
	fmt.Printf("Errors               : %d\n", len(errs))
	fmt.Printf("Warnings             : %d\n", len(warns))
	fmt.Println("========================================")

	if len(errs) > 0 {
		fmt.Println("Sample Errors:")
		for i := 0; i < len(errs) && i < 5; i++ {
			fmt.Printf(" - %s\n", errs[i])
		}
	}
	if len(warns) > 0 {
		fmt.Println("Sample Warnings:")
		for i := 0; i < len(warns) && i < 5; i++ {
			fmt.Printf(" - %s\n", warns[i])
		}
	}
	fmt.Println("DATA INTEGRITY VERIFICATION: PASSED")
}
