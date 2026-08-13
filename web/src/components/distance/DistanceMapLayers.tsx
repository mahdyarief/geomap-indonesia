import { useEffect } from 'react'
import { Marker, Popup, Polyline, useMap, useMapEvents } from 'react-leaflet'
import L from 'leaflet'
import type { Centroid } from '@/lib/types'
import { coordLabel } from '@/lib/useDistanceRoute'

export function ClickHandler({ onPick }: { onPick: (lat: number, lng: number) => void }) {
  useMapEvents({ click: (e) => onPick(e.latlng.lat, e.latlng.lng) })
  return null
}

export function FitBounds({ targets }: { targets: [number, number][] }) {
  const map = useMap()
  useEffect(() => {
    if (targets.length < 2) return
    map.fitBounds(L.latLngBounds(targets), { padding: [40, 40] })
  }, [targets, map])
  return null
}

export function NumberIcon({ n }: { n: number }) {
  return L.divIcon({
    className: '',
    html: `<div style="background:#0a7a5b;color:#fff;border-radius:9999px;width:22px;height:22px;display:flex;align-items:center;justify-content:center;font-size:12px;font-weight:600;border:2px solid #fff;box-shadow:0 1px 3px rgba(0,0,0,.35)">${n}</div>`,
    iconSize: [22, 22],
    iconAnchor: [11, 11],
  })
}

export function DistanceLayers({
  points,
  routeCoords,
}: {
  points: Centroid[]
  routeCoords: [number, number][]
}) {
  return (
    <>
      {points.map((c, i) => (
        <Marker key={i} position={[c.lat, c.lng]} icon={NumberIcon({ n: i + 1 })}>
          <Popup>
            Titik {i + 1}: {coordLabel(c)}
          </Popup>
        </Marker>
      ))}
      {routeCoords.length > 0 && (
        <Polyline positions={routeCoords} pathOptions={{ color: '#0a7a5b', weight: 4 }} />
      )}
    </>
  )
}