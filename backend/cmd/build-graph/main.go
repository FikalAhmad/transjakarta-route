package main

import (
	"flag"
	"fmt"
	"os"

	"transjakarta/internal/routing/graph"
)

func main() {
	csvPath := flag.String("csv", "data/raw/combine_tije_opendata_cleaned.csv", "path to raw CSV data")
	flag.Parse()

	fmt.Println("=== TRANSJAKARTA ROUTING GRAPH BUILDER ===")
	fmt.Printf("Source data: %s\n", *csvPath)

	g, err := graph.BuildGraphFromCSV(*csvPath)
	if err != nil {
		fmt.Printf("Error building graph: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\nSUCCESS: Graph successfully built.\n")
	fmt.Printf("- Stops: %d\n", len(g.StopsByID))
	fmt.Printf("- State Nodes: %d\n", len(g.Nodes))
	fmt.Printf("- Routes: %d\n", len(g.RoutesByID))
}
