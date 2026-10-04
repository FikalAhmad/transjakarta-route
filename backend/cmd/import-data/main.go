package main

import (
	"database/sql"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"

	_ "github.com/lib/pq"

	"transjakarta/internal/domain"
	"transjakarta/internal/normalization"
)

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgrespassword@localhost:5432/transjakarta?sslmode=disable"
	}

	csvPath := flag.String("csv", "data/raw/combine_tije_opendata_cleaned.csv", "path to cleaned CSV file")
	flag.Parse()

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Database connection error: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Printf("Warning: Postgres not reachable at %s. Skipping DB insert. (Graph build and API remain fully functional in-memory)", dbURL)
		return
	}

	fmt.Println("Connected to PostgreSQL successfully. Beginning idempotent data import...")

	f, err := os.Open(*csvPath)
	if err != nil {
		log.Fatalf("Error opening CSV: %v", err)
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.Comma = ';'
	r.Read() // Header

	routes := make(map[string]domain.Route)
	stops := make(map[string]domain.Stop)

	rowCount := 0
	for {
		rec, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}
		rowCount++

		rID := rec[0]
		rName := rec[1]
		rCol := rec[2]
		sName := rec[7]
		lat, _ := strconv.ParseFloat(rec[8], 64)
		lon, _ := strconv.ParseFloat(rec[9], 64)

		routes[rID] = domain.Route{
			RouteID:        rID,
			RouteShortName: rID,
			RouteLongName:  rName,
			RouteColor:     rCol,
			ServiceType:    normalization.ClassifyService(rID),
		}

		sID := normalization.GenerateStopID(sName, lat, lon)
		stops[sID] = domain.Stop{
			StopID:    sID,
			StopName:  sName,
			Latitude:  lat,
			Longitude: lon,
		}
	}

	// Insert Routes
	tx, err := db.Begin()
	if err != nil {
		log.Fatalf("Transaction start error: %v", err)
	}

	stmtRoute, _ := tx.Prepare(`
		INSERT INTO routes (route_id, route_short_name, route_long_name, route_color, service_type)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (route_id) DO NOTHING;
	`)
	for _, r := range routes {
		stmtRoute.Exec(r.RouteID, r.RouteShortName, r.RouteLongName, r.RouteColor, string(r.ServiceType))
	}
	stmtRoute.Close()

	// Insert Stops with PostGIS Point
	stmtStop, _ := tx.Prepare(`
		INSERT INTO stops (stop_id, stop_name, latitude, longitude, geom)
		VALUES ($1, $2, $3, $4, ST_SetSRID(ST_MakePoint($4, $3), 4326))
		ON CONFLICT (stop_id) DO NOTHING;
	`)
	for _, s := range stops {
		stmtStop.Exec(s.StopID, s.StopName, s.Latitude, s.Longitude)
	}
	stmtStop.Close()

	if err := tx.Commit(); err != nil {
		log.Fatalf("Failed to commit import transaction: %v", err)
	}

	fmt.Printf("Idempotent Import Summary:\n- Rows parsed: %d\n- Routes synced: %d\n- Stops synced: %d\n", rowCount, len(routes), len(stops))
}
