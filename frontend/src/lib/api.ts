import type { RoutePlanResult, RouteSearchRequest, RouteSearchResponse, Stop } from '../types/api'

export async function searchRoutes(req: RouteSearchRequest): Promise<RoutePlanResult[]> {
  const res = await fetch('/api/v1/routes/search', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(req),
  })

  if (!res.ok) {
    const errBody = await res.json().catch(() => ({}))
    throw new Error(errBody.error?.message || `Search failed with HTTP ${res.status}`)
  }

  const json: RouteSearchResponse = await res.json()
  return json.data.routes
}

export async function getNearbyStops(lat: number, lon: number, radius = 1000, limit = 10): Promise<Stop[]> {
  const params = new URLSearchParams({
    latitude: lat.toString(),
    longitude: lon.toString(),
    radius: radius.toString(),
    limit: limit.toString(),
  })
  const res = await fetch(`/api/v1/stops/nearby?${params}`)
  if (!res.ok) {
    throw new Error('Failed to fetch nearby stops')
  }
  const json = await res.json()
  return json.stops || []
}
