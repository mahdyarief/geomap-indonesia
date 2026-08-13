import { useState } from 'react'
import { useMutation } from '@tanstack/react-query'
import { api } from '@/lib/api'
import type { Centroid, DistanceResult } from '@/lib/types'

// Mirrors the backend validation in internal/handler/distance_handler.go.
export const INDONESIA_MIN_LAT = -11
export const INDONESIA_MAX_LAT = 6
export const INDONESIA_MIN_LNG = 95
export const INDONESIA_MAX_LNG = 141

export const withinBounds = (lat: number, lng: number) =>
  lat >= INDONESIA_MIN_LAT && lat <= INDONESIA_MAX_LAT &&
  lng >= INDONESIA_MIN_LNG && lng <= INDONESIA_MAX_LNG

export const coordLabel = (c: Centroid) => `lat ${c.lat.toFixed(5)}, lng ${c.lng.toFixed(5)}`

export const OUT_OF_BOUNDS_HINT = 'Titik di luar batas Indonesia (lat -11..6, lng 95..141) diabaikan.'

// Backend returns the route geometry as a GeoJSON MultiLineString (ST_LineMerge
// splits into several lines at OSM junctions), so coordinates are nested.
// Flatten any nesting into [lat, lng] pairs (GeoJSON order is [lng, lat]).
function flattenGeoCoords(coords: unknown): [number, number][] {
  const out: [number, number][] = []
  const walk = (node: unknown): void => {
    if (!Array.isArray(node) || node.length === 0) return
    if (typeof node[0] === 'number') {
      if (node.length >= 2) out.push([node[1] as number, node[0] as number])
      return
    }
    for (const child of node) walk(child)
  }
  walk(coords)
  return out
}

export type RouteSummary = {
  totalKm: number
  totalMin: number
  routeCoords: [number, number][]
}

export function useDistanceRoute() {
  const [points, setPoints] = useState<Centroid[]>([])
  const [hint, setHint] = useState('')

  const mutation = useMutation({
    mutationFn: async (pts: Centroid[]): Promise<RouteSummary> => {
      let totalKm = 0
      let totalMin = 0
      const routeCoords: [number, number][] = []
      for (let i = 0; i < pts.length - 1; i++) {
        const r = await api<DistanceResult>('/api/v1/distance', {
          method: 'POST',
          body: JSON.stringify({ origin: pts[i], destination: pts[i + 1] }),
        })
        totalKm += r.distance_km
        totalMin += r.duration_minutes
        if (r.geometry?.coordinates?.length) {
          routeCoords.push(...flattenGeoCoords(r.geometry.coordinates))
        }
      }
      return { totalKm, totalMin, routeCoords }
    },
  })

  const ready = points.length >= 2
  const routeCoords = mutation.data?.routeCoords ?? []
  const fitTargets: [number, number][] = points.map((c) => [c.lat, c.lng] as [number, number])

  const addPoint = (lat: number, lng: number) => {
    const c = { lat, lng }
    if (!withinBounds(c.lat, c.lng)) {
      setHint(OUT_OF_BOUNDS_HINT)
      return
    }
    setHint('')
    setPoints((p) => [...p, c])
    mutation.reset()
  }

  const removePoint = (idx: number) => {
    setPoints((p) => p.filter((_, i) => i !== idx))
    mutation.reset()
  }

  const resetRoute = () => {
    setPoints([])
    setHint('')
    mutation.reset()
  }

  return {
    points,
    hint,
    mutation,
    ready,
    routeCoords,
    fitTargets,
    addPoint,
    removePoint,
    resetRoute,
  }
}