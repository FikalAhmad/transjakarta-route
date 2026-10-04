import React from 'react'
import type { RoutePlanResult } from '../../types/api'
import { Clock, Footprints, ArrowRightLeft, MapPin, CheckCircle2 } from 'lucide-react'

interface RouteTimelineProps {
  plan: RoutePlanResult
}

export const RouteTimeline: React.FC<RouteTimelineProps> = ({ plan }) => {
  return (
    <div className="space-y-4">
      {/* Route Header Metrics */}
      <div className="bg-white p-4 rounded-xl shadow-sm border border-slate-100 flex items-center justify-between">
        <div>
          <div className="text-xs font-semibold uppercase tracking-wider text-slate-400">Total Perjalanan</div>
          <div className="text-2xl font-black text-slate-800 flex items-center gap-1.5 mt-0.5">
            <Clock className="w-5 h-5 text-blue-600" />
            {plan.total_duration_minutes} <span className="text-sm font-normal text-slate-500">menit</span>
          </div>
        </div>

        <div className="flex gap-4 border-l pl-4 border-slate-100">
          <div>
            <div className="text-xs text-slate-400 font-medium">Jalan Kaki</div>
            <div className="text-sm font-bold text-slate-700 flex items-center gap-1 mt-0.5">
              <Footprints className="w-4 h-4 text-emerald-600" />
              {plan.walking_distance_meters} m
            </div>
          </div>
          <div>
            <div className="text-xs text-slate-400 font-medium">Transit</div>
            <div className="text-sm font-bold text-slate-700 flex items-center gap-1 mt-0.5">
              <ArrowRightLeft className="w-4 h-4 text-amber-500" />
              {plan.transfer_count}x
            </div>
          </div>
          <div>
            <div className="text-xs text-slate-400 font-medium">Tarif</div>
            <div className="text-sm font-bold text-blue-600 mt-0.5">
              Rp {plan.total_fare.toLocaleString('id-ID')}
            </div>
          </div>
        </div>
      </div>

      {/* Segments Step-by-Step */}
      <div className="bg-white p-5 rounded-xl shadow-sm border border-slate-100 space-y-4">
        <h3 className="text-sm font-bold text-slate-700 uppercase tracking-wider">Panduan Langkah Rute</h3>
        <div className="relative pl-6 border-l-2 border-slate-200 space-y-6">
          {plan.segments.map((seg, idx) => {
            if (seg.type === 'WALK') {
              return (
                <div key={idx} className="relative group">
                  <div className="absolute -left-[31px] top-0.5 w-4 h-4 rounded-full bg-emerald-500 border-2 border-white flex items-center justify-center text-white" />
                  <div className="text-sm font-semibold text-slate-800 flex items-center gap-2">
                    <Footprints className="w-4 h-4 text-emerald-600" />
                    Jalan Kaki ke {seg.to?.stop_name}
                  </div>
                  <div className="text-xs text-slate-500 mt-0.5">
                    {seg.distance_meters} m • ~{seg.duration_minutes.toFixed(0)} menit
                  </div>
                </div>
              )
            }

            if (seg.type === 'TRANSIT') {
              return (
                <div key={idx} className="relative group">
                  <div className="absolute -left-[33px] top-0.5 w-5 h-5 rounded-full bg-blue-600 border-2 border-white flex items-center justify-center text-white text-[10px] font-bold">
                    TJ
                  </div>
                  <div className="bg-blue-50/60 p-3 rounded-lg border border-blue-100">
                    <div className="flex items-center gap-2">
                      <span className="px-2 py-0.5 bg-blue-600 text-white rounded font-bold text-xs">
                        {seg.route_id}
                      </span>
                      <span className="text-xs font-semibold text-slate-700">
                        {seg.route_name}
                      </span>
                    </div>
                    <div className="mt-2 text-xs text-slate-600 space-y-1">
                      <div className="flex items-center gap-1.5 font-medium text-slate-800">
                        <MapPin className="w-3.5 h-3.5 text-blue-500" />
                        Naik di: {seg.boarding_stop?.stop_name}
                      </div>
                      <div className="flex items-center gap-1.5 font-medium text-slate-800">
                        <CheckCircle2 className="w-3.5 h-3.5 text-emerald-500" />
                        Turun di: {seg.alighting_stop?.stop_name}
                      </div>
                      <div className="text-slate-500 pt-1">
                        {seg.intermediate_stops?.length || 0} halte terlewati • ~{seg.duration_minutes.toFixed(0)} menit ({(seg.distance_meters / 1000).toFixed(1)} km)
                      </div>
                    </div>
                  </div>
                </div>
              )
            }

            if (seg.type === 'TRANSFER') {
              return (
                <div key={idx} className="relative group">
                  <div className="absolute -left-[31px] top-0.5 w-4 h-4 rounded-full bg-amber-500 border-2 border-white flex items-center justify-center text-white" />
                  <div className="text-xs font-bold text-amber-600 uppercase tracking-wide flex items-center gap-1.5">
                    <ArrowRightLeft className="w-3.5 h-3.5" />
                    Transit di {seg.at_stop?.stop_name}
                  </div>
                  <div className="text-xs text-slate-500 mt-0.5">
                    Pindah busway (~{seg.duration_minutes} menit)
                  </div>
                </div>
              )
            }

            return null
          })}
        </div>
      </div>
    </div>
  )
}
