# TransJakarta Intelligent Route Planner

Fullstack Intelligent Transit Route Planner for TransJakarta Network built with Go, PostGIS, React, TypeScript, MapLibre GL JS, and Dijkstra Routing Engine.

## Tech Stack
- **Backend**: Go 1.24, REST API (`/api/v1/routes/search`, `/api/v1/stops/nearby`, `/health`, `/ready`)
- **Database**: PostgreSQL 16 + PostGIS Spatial Indexing
- **Algorithm**: Multi-Criteria Dijkstra on Transit State Graph (Transfer penalty, Walking weights)
- **Frontend**: React, Vite, TypeScript, Tailwind CSS, MapLibre GL JS, Lucide icons
- **Data Source**: `data/raw/combine_tije_opendata_cleaned.csv`

---

## Quick Start Guide

### 1. Data Validation & Graph Building
```bash
# Validate dataset integrity
cd backend
go run ./cmd/validate-data/main.go -csv ../data/raw/combine_tije_opendata_cleaned.csv

# Build and verify transit graph
go run ./cmd/build-graph/main.go -csv ../data/raw/combine_tije_opendata_cleaned.csv
```

### 2. Run Backend API Server
```bash
cd backend
go run ./cmd/api/main.go
# Server starts on http://localhost:8080
```

### 3. Run Frontend UI
```bash
cd frontend
npm install
npm run dev
# Vite runs on http://localhost:5173
```

---

## Docker Deployment
```bash
docker compose up -d
```
