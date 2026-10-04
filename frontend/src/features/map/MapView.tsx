import React, { useEffect } from 'react'
import { MapContainer, TileLayer, Marker, Popup, Polyline, useMap, useMapEvents } from 'react-leaflet'
import L from 'leaflet'
import 'leaflet/dist/leaflet.css'
import type { RoutePlanResult, Coordinate, Stop } from '../../types/api'

// Custom Leaflet DivIcons
const iconA = L.divIcon({
  className: '',
  html: '<div class="custom-pin custom-pin-a">A</div>',
  iconSize: [32, 32],
  iconAnchor: [16, 16],
})

const iconB = L.divIcon({
  className: '',
  html: '<div class="custom-pin custom-pin-b">B</div>',
  iconSize: [32, 32],
  iconAnchor: [16, 16],
})

const iconTJ = L.divIcon({
  className: '',
  html: '<div class="custom-pin custom-pin-tj">TJ</div>',
  iconSize: [26, 26],
  iconAnchor: [13, 13],
})

// Map click listener component
const MapClickHandler: React.FC<{ onMapClick?: (coords: Coordinate) => void }> = ({ onMapClick }) => {
  useMapEvents({
    click(e) {
      if (onMapClick) {
        onMapClick({ latitude: e.latlng.lat, longitude: e.latlng.lng })
      }
    },
  })
  return null
}

// Auto-fit bounds controller
const MapAutoFitter: React.FC<{
  origin: Coordinate | null
  destination: Coordinate | null
  routeCoords: [number, number][]
}> = ({ origin, destination, routeCoords }) => {
  const map = useMap()

  useEffect(() => {
    if (routeCoords.length > 0) {
      const bounds = L.latLngBounds(routeCoords.map((c) => [c[0], c[1]]))
      map.fitBounds(bounds, { padding: [60, 60], maxZoom: 15 })
    } else if (origin && destination) {
      const bounds = L.latLngBounds([
        [origin.latitude, origin.longitude],
        [destination.latitude, destination.longitude],
      ])
      map.fitBounds(bounds, { padding: [80, 80], maxZoom: 14 })
    }
  }, [origin, destination, routeCoords, map])

  return null
}

interface MapViewProps {
  origin: Coordinate | null
  destination: Coordinate | null
  selectedRoute: RoutePlanResult | null
  onMapClick?: (coords: Coordinate) => void
}

export const MapView: React.FC<MapViewProps> = ({
  origin,
  destination,
  selectedRoute,
  onMapClick,
}) => {
  // Convert [lon, lat] from GeoJSON to Leaflet [lat, lon]
  const routeLatLngs: [number, number][] =
    selectedRoute?.full_geometry?.coordinates?.map((coord) => [coord[1], coord[0]]) || []

  const previewLatLngs: [number, number][] =
    origin && destination && routeLatLngs.length === 0
      ? [
          [origin.latitude, origin.longitude],
          [destination.latitude, destination.longitude],
        ]
      : []

  return (
    <div className="relative w-full h-full min-h-[400px]">
      <MapContainer
        center={[-6.1754, 106.8272]}
        zoom={12}
        scrollWheelZoom={true}
        className="w-full h-full"
      >
        {/* OpenStreetMap TileLayer */}
        <TileLayer
          attribution='&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors'
          url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png"
        />

        <MapClickHandler onMapClick={onMapClick} />
        <MapAutoFitter origin={origin} destination={destination} routeCoords={routeLatLngs} />

        {/* Marker A (Origin) */}
        {origin && (
          <Marker position={[origin.latitude, origin.longitude]} icon={iconA}>
            <Popup>
              <b>Titik Keberangkatan (A)</b>
            </Popup>
          </Marker>
        )}

        {/* Marker B (Destination) */}
        {destination && (
          <Marker position={[destination.latitude, destination.longitude]} icon={iconB}>
            <Popup>
              <b>Titik Tujuan (B)</b>
            </Popup>
          </Marker>
        )}

        {/* Intermediate Bus Halte Stops */}
        {selectedRoute?.segments?.map((seg, sIdx) => {
          if (seg.type === 'TRANSIT' && seg.intermediate_stops) {
            return seg.intermediate_stops.map((stop: Stop, idx: number) => {
              const isBoarding = idx === 0
              const isAlighting = idx === (seg.intermediate_stops?.length || 0) - 1
              if (isBoarding || isAlighting) {
                return (
                  <Marker
                    key={`${sIdx}-${idx}-${stop.stop_id}`}
                    position={[stop.latitude, stop.longitude]}
                    icon={iconTJ}
                  >
                    <Popup>
                      <div className="font-sans text-xs">
                        <b className="text-blue-700 text-sm">{stop.stop_name}</b>
                        <br />
                        <span>{isBoarding ? '🟢 Naik Bus' : '🔴 Turun Bus'} ({seg.route_id} - {seg.route_name})</span>
                      </div>
                    </Popup>
                  </Marker>
                )
              }
              return null
            })
          }
          return null
        })}

        {/* Preview Direct Line A -> B */}
        {previewLatLngs.length > 0 && (
          <Polyline
            positions={previewLatLngs}
            pathOptions={{
              color: '#f43f5e',
              weight: 4,
              dashArray: '8, 8',
              opacity: 0.85,
            }}
          />
        )}

        {/* Active TransJakarta Route Polyline */}
        {routeLatLngs.length > 0 && (
          <>
            {/* Outer white casing */}
            <Polyline
              positions={routeLatLngs}
              pathOptions={{
                color: '#ffffff',
                weight: 9,
                opacity: 0.9,
              }}
            />
            {/* Main vibrant blue route line */}
            <Polyline
              positions={routeLatLngs}
              pathOptions={{
                color: '#2563eb',
                weight: 5,
                opacity: 1.0,
              }}
            />
          </>
        )}
      </MapContainer>
    </div>
  )
}
