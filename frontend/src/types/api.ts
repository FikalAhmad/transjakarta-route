export type Preference = 'BALANCED' | 'FASTEST' | 'FEWEST_TRANSFERS' | 'LEAST_WALKING'

export type ServiceType = 'BRT' | 'NON_BRT' | 'MIKROTRANS' | 'AMARI'

export interface Coordinate {
  latitude: number
  longitude: number
}

export interface Stop {
  stop_id: string
  stop_name: string
  latitude: number
  longitude: number
  location_type?: number
  distance_meters?: number
}

export interface GeoJSONGeometry {
  type: string
  coordinates: [number, number][]
}

export interface Segment {
  type: 'WALK' | 'TRANSIT' | 'TRANSFER'
  description: string
  distance_meters: number
  duration_minutes: number
  from?: Stop
  to?: Stop
  at_stop?: Stop
  route_id?: string
  route_name?: string
  route_color?: string
  service_type?: ServiceType
  direction_id?: number
  boarding_stop?: Stop
  alighting_stop?: Stop
  intermediate_stops?: Stop[]
  geometry?: GeoJSONGeometry
}

export interface RoutePlanResult {
  route_id: string
  summary: string
  total_duration_minutes: number
  total_distance_meters: number
  walking_distance_meters: number
  transit_distance_meters: number
  transfer_count: number
  total_fare: number
  segments: Segment[]
  full_geometry?: GeoJSONGeometry
}

export interface RouteSearchRequest {
  origin: Coordinate
  destination: Coordinate
  preference: Preference
}

export interface RouteSearchResponse {
  data: {
    routes: RoutePlanResult[]
  }
}
