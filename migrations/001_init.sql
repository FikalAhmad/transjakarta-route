-- 001_init.sql

CREATE EXTENSION IF NOT EXISTS postgis;

-- Routes table
CREATE TABLE IF NOT EXISTS routes (
    route_id VARCHAR(64) PRIMARY KEY,
    route_short_name VARCHAR(64) NOT NULL,
    route_long_name VARCHAR(255) NOT NULL,
    route_color VARCHAR(16),
    service_type VARCHAR(32) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Trips table
CREATE TABLE IF NOT EXISTS trips (
    trip_id VARCHAR(64) PRIMARY KEY,
    route_id VARCHAR(64) NOT NULL REFERENCES routes(route_id) ON DELETE CASCADE,
    service_id VARCHAR(64) NOT NULL,
    direction_id INT NOT NULL,
    shape_id VARCHAR(64),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Stops table
CREATE TABLE IF NOT EXISTS stops (
    stop_id VARCHAR(64) PRIMARY KEY,
    stop_name VARCHAR(255) NOT NULL,
    latitude DOUBLE PRECISION NOT NULL,
    longitude DOUBLE PRECISION NOT NULL,
    geom GEOMETRY(Point, 4326),
    location_type INT DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Stop Times / Trip Stops table
CREATE TABLE IF NOT EXISTS stop_times (
    id SERIAL PRIMARY KEY,
    trip_id VARCHAR(64) NOT NULL REFERENCES trips(trip_id) ON DELETE CASCADE,
    stop_id VARCHAR(64) NOT NULL REFERENCES stops(stop_id) ON DELETE CASCADE,
    stop_sequence INT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT unique_trip_stop_sequence UNIQUE (trip_id, stop_sequence)
);

-- Shapes table (for polyline visualization)
CREATE TABLE IF NOT EXISTS shapes (
    id SERIAL PRIMARY KEY,
    shape_id VARCHAR(64) NOT NULL,
    shape_pt_lat DOUBLE PRECISION NOT NULL,
    shape_pt_lon DOUBLE PRECISION NOT NULL,
    shape_pt_sequence INT NOT NULL,
    geom GEOMETRY(Point, 4326),
    CONSTRAINT unique_shape_sequence UNIQUE (shape_id, shape_pt_sequence)
);

-- Dataset Metadata
CREATE TABLE IF NOT EXISTS dataset_metadata (
    id SERIAL PRIMARY KEY,
    dataset_version VARCHAR(64) NOT NULL,
    source_filename VARCHAR(255) NOT NULL,
    row_count INT NOT NULL,
    imported_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indices
CREATE INDEX IF NOT EXISTS idx_trips_route_id ON trips(route_id);
CREATE INDEX IF NOT EXISTS idx_trips_direction_id ON trips(direction_id);
CREATE INDEX IF NOT EXISTS idx_stop_times_trip_id ON stop_times(trip_id);
CREATE INDEX IF NOT EXISTS idx_stop_times_stop_id ON stop_times(stop_id);
CREATE INDEX IF NOT EXISTS idx_stop_times_sequence ON stop_times(trip_id, stop_sequence);
CREATE INDEX IF NOT EXISTS idx_stops_geom ON stops USING GIST(geom);
CREATE INDEX IF NOT EXISTS idx_shapes_shape_id ON shapes(shape_id);
