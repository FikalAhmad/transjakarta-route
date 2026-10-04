import React, { useState } from 'react'
import type { Coordinate, Preference, RoutePlanResult } from './types/api'
import { searchRoutes } from './lib/api'
import { MapView } from './features/map/MapView'
import { RouteTimeline } from './features/route-result/RouteTimeline'
import { Bus, Search, Loader2, Info } from 'lucide-react'

// Popular default Jakarta landmark coordinates
const PRESET_LOCATIONS = [
  { name: 'Blok M Terminal', lat: -6.2444, lon: 106.7979 },
  { name: 'Monumen Nasional (Monas)', lat: -6.1754, lon: 106.8272 },
  { name: 'Kota Tua Jakarta', lat: -6.1376, lon: 106.8133 },
  { name: 'Bundaran HI', lat: -6.1950, lon: 106.8230 },
  { name: 'Pulo Gadung', lat: -6.1917, lon: 106.9015 },
  { name: 'Kampung Rambutan', lat: -6.3115, lon: 106.8785 },
]

export const App: React.FC = () => {
  const [origin, setOrigin] = useState<Coordinate>({ latitude: -6.2444, longitude: 106.7979 }) // Blok M
  const [destination, setDestination] = useState<Coordinate>({ latitude: -6.1376, longitude: 106.8133 }) // Kota
  const [preference, setPreference] = useState<Preference>('BALANCED')

  const [routes, setRoutes] = useState<RoutePlanResult[]>([])
  const [selectedRouteIndex, setSelectedRouteIndex] = useState<number>(0)
  const [loading, setLoading] = useState<boolean>(false)
  const [error, setError] = useState<string | null>(null)
  const [selectionMode, setSelectionMode] = useState<'origin' | 'destination' | null>(null)

  const handleSearch = async () => {
    setLoading(true)
    setError(null)
    try {
      const results = await searchRoutes({
        origin,
        destination,
        preference,
      })
      setRoutes(results)
      setSelectedRouteIndex(0)
    } catch (err: any) {
      setError(err.message || 'Gagal mencari rute.')
      setRoutes([])
    } finally {
      setLoading(false)
    }
  }

  const handleMapClick = (coords: Coordinate) => {
    if (selectionMode === 'origin') {
      setOrigin(coords)
      setSelectionMode(null)
    } else if (selectionMode === 'destination') {
      setDestination(coords)
      setSelectionMode(null)
    }
  }

  const selectedRoute = routes[selectedRouteIndex] || null

  return (
    <div className="flex h-screen w-screen overflow-hidden bg-slate-50 font-sans text-slate-900">
      {/* Sidebar Panel */}
      <div className="w-[480px] h-full flex flex-col bg-white border-r border-slate-200 shadow-xl z-10">
        {/* Brand Header */}
        <div className="px-6 py-4 bg-blue-700 text-white flex items-center justify-between shadow-md">
          <div className="flex items-center gap-2.5">
            <div className="w-9 h-9 rounded-lg bg-white/15 flex items-center justify-center backdrop-blur-sm">
              <Bus className="w-5 h-5 text-white" />
            </div>
            <div>
              <h1 className="font-bold text-base leading-tight">TransJakarta Intelligent Planner</h1>
              <p className="text-xs text-blue-100 font-normal">Official OpenData Network Route Engine</p>
            </div>
          </div>
          <span className="px-2 py-0.5 text-[10px] font-semibold bg-emerald-500/20 text-emerald-200 border border-emerald-400/30 rounded-full">
            Live v1.0
          </span>
        </div>

        {/* Search & Configuration Container */}
        <div className="p-5 overflow-y-auto flex-1 space-y-4">
          {/* Location Inputs */}
          <div className="space-y-3 bg-slate-50 p-3.5 rounded-xl border border-slate-200/80">
            {/* Origin Input */}
            <div>
              <div className="flex justify-between items-center mb-1">
                <label className="text-xs font-bold text-slate-600 flex items-center gap-1.5">
                  <span className="w-2.5 h-2.5 rounded-full bg-emerald-500 inline-block" />
                  Titik Keberangkatan (Origin)
                </label>
                <button
                  onClick={() => setSelectionMode('origin')}
                  className={`text-[11px] font-medium px-2 py-0.5 rounded transition ${
                    selectionMode === 'origin' ? 'bg-emerald-600 text-white' : 'text-emerald-700 hover:bg-emerald-50'
                  }`}
                >
                  {selectionMode === 'origin' ? 'Klik di Peta...' : 'Pilih di Peta'}
                </button>
              </div>
              <div className="text-xs font-mono bg-white p-2 rounded border border-slate-200 text-slate-700">
                Lat: {origin.latitude.toFixed(5)}, Lon: {origin.longitude.toFixed(5)}
              </div>
            </div>

            {/* Destination Input */}
            <div>
              <div className="flex justify-between items-center mb-1">
                <label className="text-xs font-bold text-slate-600 flex items-center gap-1.5">
                  <span className="w-2.5 h-2.5 rounded-full bg-rose-500 inline-block" />
                  Titik Tujuan (Destination)
                </label>
                <button
                  onClick={() => setSelectionMode('destination')}
                  className={`text-[11px] font-medium px-2 py-0.5 rounded transition ${
                    selectionMode === 'destination' ? 'bg-rose-600 text-white' : 'text-rose-700 hover:bg-rose-50'
                  }`}
                >
                  {selectionMode === 'destination' ? 'Klik di Peta...' : 'Pilih di Peta'}
                </button>
              </div>
              <div className="text-xs font-mono bg-white p-2 rounded border border-slate-200 text-slate-700">
                Lat: {destination.latitude.toFixed(5)}, Lon: {destination.longitude.toFixed(5)}
              </div>
            </div>

            {/* Quick Presets */}
            <div>
              <span className="text-[11px] font-medium text-slate-500 block mb-1">Pilihan Cepat:</span>
              <div className="flex flex-wrap gap-1.5">
                {PRESET_LOCATIONS.map((loc) => (
                  <button
                    key={loc.name}
                    onClick={() => {
                      if (!selectionMode || selectionMode === 'destination') {
                        setDestination({ latitude: loc.lat, longitude: loc.lon })
                      } else {
                        setOrigin({ latitude: loc.lat, longitude: loc.lon })
                      }
                    }}
                    className="text-[11px] px-2 py-1 bg-white hover:bg-slate-100 text-slate-600 rounded border border-slate-200"
                  >
                    {loc.name}
                  </button>
                ))}
              </div>
            </div>
          </div>

          {/* Preferences */}
          <div>
            <label className="text-xs font-bold text-slate-600 block mb-1.5">Preferensi Perjalanan</label>
            <div className="grid grid-cols-2 gap-2">
              {[
                { id: 'BALANCED', label: 'Seimbang', desc: 'Rekomendasi terbaik' },
                { id: 'FASTEST', label: 'Tercepat', desc: 'Durasi minimal' },
                { id: 'FEWEST_TRANSFERS', label: 'Paling Sedikit Transit', desc: 'Minim ganti armada' },
                { id: 'LEAST_WALKING', label: 'Minim Jalan Kaki', desc: 'Halte terdekat' },
              ].map((p) => (
                <button
                  key={p.id}
                  onClick={() => setPreference(p.id as Preference)}
                  className={`text-left p-2.5 rounded-lg border transition ${
                    preference === p.id
                      ? 'border-blue-600 bg-blue-50/70 text-blue-900 font-semibold'
                      : 'border-slate-200 bg-white hover:bg-slate-50 text-slate-700'
                  }`}
                >
                  <div className="text-xs font-bold">{p.label}</div>
                  <div className="text-[10px] text-slate-500">{p.desc}</div>
                </button>
              ))}
            </div>
          </div>

          {/* Search Action Button */}
          <button
            onClick={handleSearch}
            disabled={loading}
            className="w-full py-3 bg-blue-600 hover:bg-blue-700 active:bg-blue-800 disabled:bg-blue-400 text-white font-bold rounded-xl shadow-md transition flex items-center justify-center gap-2 cursor-pointer"
          >
            {loading ? (
              <>
                <Loader2 className="w-5 h-5 animate-spin" />
                <span>Menghitung Rute Optimal...</span>
              </>
            ) : (
              <>
                <Search className="w-5 h-5" />
                <span>Cari Rute TransJakarta</span>
              </>
            )}
          </button>

          {/* Error Message */}
          {error && (
            <div className="p-3 bg-rose-50 border border-rose-200 text-rose-700 text-xs rounded-lg flex items-center gap-2">
              <Info className="w-4 h-4 text-rose-500 shrink-0" />
              <span>{error}</span>
            </div>
          )}

          {/* Route Options Tabs */}
          {routes.length > 0 && (
            <div className="space-y-3 pt-2">
              <div className="text-xs font-bold text-slate-500 uppercase tracking-wider">
                Rute Ditemukan ({routes.length})
              </div>
              <div className="flex gap-2">
                {routes.map((r, i) => (
                  <button
                    key={r.route_id + i}
                    onClick={() => setSelectedRouteIndex(i)}
                    className={`flex-1 p-2.5 rounded-lg border text-left transition ${
                      selectedRouteIndex === i
                        ? 'border-blue-600 bg-blue-600 text-white shadow-sm'
                        : 'border-slate-200 bg-white hover:bg-slate-50 text-slate-700'
                    }`}
                  >
                    <div className="text-xs font-bold truncate">Opsi #{i + 1}: {r.summary}</div>
                    <div className={`text-[11px] mt-0.5 ${selectedRouteIndex === i ? 'text-blue-100' : 'text-slate-500'}`}>
                      {r.total_duration_minutes} mnt • {r.transfer_count} transit
                    </div>
                  </button>
                ))}
              </div>

              {/* Selected Route Step-by-Step Breakdown */}
              {selectedRoute && <RouteTimeline plan={selectedRoute} />}
            </div>
          )}
        </div>
      </div>

      {/* Main Map View Area */}
      <div className="flex-1 h-full relative">
        <MapView
          origin={origin}
          destination={destination}
          selectedRoute={selectedRoute}
          onMapClick={handleMapClick}
        />
      </div>
    </div>
  )
}
export default App
