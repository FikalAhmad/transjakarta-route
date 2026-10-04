package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"transjakarta/internal/http/handlers"
	"transjakarta/internal/http/middleware"
	"transjakarta/internal/routing/graph"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	csvPath := flag.String("csv", "../data/raw/combine_tije_opendata_cleaned.csv", "path to cleaned CSV file")
	flag.Parse()

	// Check if file exists at given path or fallback
	if _, err := os.Stat(*csvPath); os.IsNotExist(err) {
		fallback := "data/raw/combine_tije_opendata_cleaned.csv"
		if _, err := os.Stat(fallback); err == nil {
			*csvPath = fallback
		}
	}

	log.Printf("Starting TransJakarta Intelligent Route API on :%s", port)
	log.Printf("Loading transit graph from %s...", *csvPath)

	transitGraph, err := graph.BuildGraphFromCSV(*csvPath)
	if err != nil {
		log.Fatalf("Fatal: failed to load transit graph: %v", err)
	}

	engine := graph.NewRoutingEngine(transitGraph)
	apiHandler := handlers.NewAPIHandler(engine)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", apiHandler.HealthCheck)
	mux.HandleFunc("GET /ready", apiHandler.ReadyCheck)
	mux.HandleFunc("GET /api/v1/stops/nearby", apiHandler.GetNearbyStops)
	mux.HandleFunc("POST /api/v1/routes/search", apiHandler.SearchRoute)
	mux.HandleFunc("GET /api/v1/routes", apiHandler.ListRoutes)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: middleware.LoggerAndCORS(mux),
	}

	log.Printf("Route Engine Server is listening on http://localhost:%s", port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server stopped: %v", err)
	}
}
