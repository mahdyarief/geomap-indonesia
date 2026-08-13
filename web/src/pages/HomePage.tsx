import { useEffect, useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { MapContainer, TileLayer, Marker, Popup, useMap, GeoJSON } from 'react-leaflet'
import L from 'leaflet'
import { Link } from 'react-router'
import { api } from '@/lib/api'
import { useAuthStore } from '@/store/auth'
import { Button, buttonVariants } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { cn } from '@/lib/utils'
import { useDistanceRoute } from '@/lib/useDistanceRoute'
import { ClickHandler, DistanceLayers, FitBounds } from '@/components/distance/DistanceMapLayers'
import { DistanceSidebar } from '@/components/distance/DistanceSidebar'
import type { Wilayah, Boundaries, ReverseResult } from '@/lib/types'

const JAKARTA_CENTER: [number, number] = [-6.2, 106.8]

function Pin({ lat, lng }: { lat: number; lng: number }) {
  return (
    <Marker position={[lat, lng]}>
      <Popup>
        <span className="text-sm">
          {lat.toFixed(5)}, {lng.toFixed(5)}
        </span>
      </Popup>
    </Marker>
  )
}

function FitPolygon({ feature }: { feature: Boundaries | undefined }) {
  const map = useMap()
  useEffect(() => {
    if (!feature?.geometry) return
    const layer = L.geoJSON(feature as unknown as GeoJSON.Feature)
    const bounds = layer.getBounds()
    if (bounds.isValid()) {
      map.fitBounds(bounds, { padding: [40, 40] })
    }
  }, [feature, map])
  return null
}

export default function HomePage() {
  const token = useAuthStore((s) => s.token)
  const {
    points,
    hint,
    mutation,
    ready,
    routeCoords,
    fitTargets,
    addPoint,
    removePoint,
    resetRoute,
  } = useDistanceRoute()
  const [selProv, setSelProv] = useState('')
  const [selKab, setSelKab] = useState('')
  const [selKec, setSelKec] = useState('')
  const [selKel, setSelKel] = useState('')
  const [coordLat, setCoordLat] = useState('')
  const [coordLng, setCoordLng] = useState('')
  const [coordSubmitted, setCoordSubmitted] = useState<{ lat: string; lng: string } | null>(null)

  const provinces = useQuery({
    queryKey: ['wilayah', 'provinsi'],
    queryFn: () => api<Wilayah[]>('/api/v1/wilayah?limit=100'),
    enabled: !!token,
  })

  const kabupaten = useQuery({
    queryKey: ['wilayah', selProv, 'children'],
    queryFn: () => api<Wilayah[]>(`/api/v1/wilayah/${selProv}/children`),
    enabled: !!selProv && !!token,
  })

  const kecamatan = useQuery({
    queryKey: ['wilayah', selKab, 'children'],
    queryFn: () => api<Wilayah[]>(`/api/v1/wilayah/${selKab}/children`),
    enabled: !!selKab && !!token,
  })

  const kelurahan = useQuery({
    queryKey: ['wilayah', selKec, 'children'],
    queryFn: () => api<Wilayah[]>(`/api/v1/wilayah/${selKec}/children`),
    enabled: !!selKec && !!token,
  })

  const selectedKode = selKel || selKec || selKab || selProv

  const boundary = useQuery({
    queryKey: ['boundary', selectedKode],
    queryFn: () => api<Boundaries>(`/api/v1/boundaries/${selectedKode}`),
    enabled: !!selectedKode && !!token,
  })

  const reverse = useQuery({
    queryKey: ['reverse', coordSubmitted?.lat ?? null, coordSubmitted?.lng ?? null],
    queryFn: () =>
      api<ReverseResult>(
        `/api/v1/reverse?lat=${encodeURIComponent(coordSubmitted!.lat)}&lng=${encodeURIComponent(coordSubmitted!.lng)}`,
      ),
    enabled: !!coordSubmitted && !!token,
  })

  const reverseBoundary = useQuery({
    queryKey: [
      'reverse-boundary',
      reverse.data?.kelurahan?.kode,
      reverse.data?.kecamatan?.kode,
      reverse.data?.kabupaten?.kode,
      reverse.data?.provinsi?.kode,
    ],
    queryFn: async () => {
      const rev = reverse.data!
      const candidates = [
        { kode: rev.kelurahan.kode, nama: rev.kelurahan.nama },
        { kode: rev.kecamatan.kode, nama: rev.kecamatan.nama },
        { kode: rev.kabupaten.kode, nama: rev.kabupaten.nama },
        { kode: rev.provinsi.kode, nama: rev.provinsi.nama },
      ]
      for (const cand of candidates) {
        if (!cand.kode) continue
        try {
          const b = await api<Boundaries>(`/api/v1/boundaries/${cand.kode}`)
          if (b?.geometry) return { ...b, nama: cand.nama, kode: cand.kode }
        } catch {
          // level ini tidak punya data batas (404) -> lanjut ke level yang lebih tinggi
        }
      }
      return null
    },
    enabled: !!reverse.data && !!token,
  })

  const selectedNama =
    (selKel && kelurahan.data?.find((c) => c.kode === selKel)?.nama) ||
    (selKec && kecamatan.data?.find((c) => c.kode === selKec)?.nama) ||
    (selKab && kabupaten.data?.find((c) => c.kode === selKab)?.nama) ||
    (selProv && provinces.data?.find((c) => c.kode === selProv)?.nama) ||
    boundary.data?.properties?.nama ||
    ''

  const handleCoordSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (!coordLat.trim() || !coordLng.trim()) return
    setCoordSubmitted({ lat: coordLat.trim(), lng: coordLng.trim() })
  }

  const handleResetFilter = () => {
    setSelProv('')
    setSelKab('')
    setSelKec('')
    setSelKel('')
    setCoordSubmitted(null)
  }

  return (
    <div className="min-h-screen flex flex-col">
      <header className="border-b border-border bg-background">
        <div className="mx-auto flex h-14 max-w-6xl items-center justify-between px-4">
          <div className="flex items-center gap-2">
            <div className="flex h-7 w-7 items-center justify-center rounded-md bg-primary text-xs font-semibold text-primary-foreground">
              G
            </div>
            <p className="text-sm font-semibold">Indonesia Geomapping</p>
          </div>
          <div className="flex items-center gap-2">
            {token ? (
              <Link to="/app/dashboard" className={cn(buttonVariants({ size: 'sm' }))}>
                Dashboard
              </Link>
            ) : (
              <Link to="/login" className={cn(buttonVariants({ size: 'sm' }))}>
                Masuk
              </Link>
            )}
          </div>
        </div>
      </header>

      <main className="flex-1">
        <div className="mx-auto max-w-6xl px-4 py-6">
          <h1 className="text-xl font-semibold">Peta Wilayah Indonesia</h1>
          <p className="mt-1 text-sm text-muted-foreground">
            Filter lokasi untuk menampilkan batas area, atau klik peta untuk menambah titik dan menghitung jarak rute jalan.
          </p>
        </div>

        <div className="mx-auto max-w-6xl px-4 pb-8">
          <Card className="mb-4">
            <CardContent className="flex flex-col gap-2 sm:flex-row sm:items-end sm:justify-between">
              <div className="flex flex-col gap-2">
                <div className="flex items-center justify-between gap-2">
                  <Label htmlFor="lokasi">Filter Lokasi</Label>
                  <Button
                    type="button"
                    variant="outline"
                    size="sm"
                    onClick={handleResetFilter}
                    disabled={!selProv && !selKab && !selKec && !selKel &&!coordSubmitted && points.length === 0}
                  >
                    Reset
                  </Button>
                </div>
                <div className="flex flex-col gap-2 sm:flex-row sm:flex-wrap">
                  <select
                    id="lokasi"
                    value={selProv}
                    onChange={(e) => {
                      setSelProv(e.target.value)
                      setSelKab('')
                      setSelKec('')
                      setSelKel('')
                    }}
                    className="h-9 min-w-[180px] rounded-md border border-input bg-background px-3 text-sm outline-none focus-visible:ring-2 focus-visible:ring-ring"
                  >
                    <option value="">Provinsi…</option>
                    {provinces.data?.map((p) => (
                      <option key={p.kode} value={p.kode}>
                        {p.nama}
                      </option>
                    ))}
                  </select>
                  <select
                    id="lokasi-kab"
                    value={selKab}
                    onChange={(e) => {
                      setSelKab(e.target.value)
                      setSelKec('')
                      setSelKel('')
                    }}
                    disabled={!selProv}
                    className="h-9 min-w-[180px] rounded-md border border-input bg-background px-3 text-sm outline-none focus-visible:ring-2 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50"
                  >
                    <option value="">Kabupaten…</option>
                    {kabupaten.data?.map((k) => (
                      <option key={k.kode} value={k.kode}>
                        {k.nama}
                      </option>
                    ))}
                  </select>
                  <select
                    id="lokasi-kec"
                    value={selKec}
                    onChange={(e) => {
                      setSelKec(e.target.value)
                      setSelKel('')
                    }}
                    disabled={!selKab}
                    className="h-9 min-w-[180px] rounded-md border border-input bg-background px-3 text-sm outline-none focus-visible:ring-2 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50"
                  >
                    <option value="">Kecamatan…</option>
                    {kecamatan.data?.map((k) => (
                      <option key={k.kode} value={k.kode}>
                        {k.nama}
                      </option>
                    ))}
                  </select>
                  <select
                    id="lokasi-kel"
                    value={selKel}
                    onChange={(e) => setSelKel(e.target.value)}
                    disabled={!selKec}
                    className="h-9 min-w-[180px] rounded-md border border-input bg-background px-3 text-sm outline-none focus-visible:ring-2 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50"
                  >
                    <option value="">Kelurahan…</option>
                    {kelurahan.data?.map((k) => (
                      <option key={k.kode} value={k.kode}>
                        {k.nama}
                      </option>
                    ))}
                  </select>
                </div>
                {!token && (
                  <p className="text-xs text-muted-foreground">
                    Masuk untuk memuat daftar wilayah dan menampilkan batas area.
                  </p>
                )}
              </div>

              {boundary.isLoading && (
                <p className="text-sm text-muted-foreground">Memuat batas area…</p>
              )}
              {boundary.error && (
                <p className="text-sm text-destructive">
                  {(boundary.error as Error).message}
                </p>
              )}
              {selectedNama && !boundary.isLoading && !boundary.error && (
                <p className="text-sm font-medium">
                  {selectedNama}
                  {boundary.data?.geometry ? ' — batas area ditampilkan' : ' — belum ada data batas'}
                </p>
              )}
            </CardContent>
          </Card>

          <Card className="mb-4">
            <CardContent className="flex flex-col gap-3">
              <form onSubmit={handleCoordSubmit} className="flex flex-col gap-3">
                <div>
                  <Label htmlFor="coord-lat">Koordinat (reverse geocode)</Label>
                  <p className="mt-1 text-xs text-muted-foreground">
                    Masukkan lat &amp; lng untuk mengetahui lokasi dan menampilkan batas area.
                  </p>
                </div>
                <div className="flex flex-col gap-2 sm:flex-row sm:items-end">
                  <div className="flex flex-col gap-1.5">
                    <Label htmlFor="coord-lat">Latitude</Label>
                    <Input
                      id="coord-lat"
                      type="number"
                      step="any"
                      value={coordLat}
                      onChange={(e) => setCoordLat(e.target.value)}
                      placeholder="-6.9175"
                      className="w-full sm:w-40"
                    />
                  </div>
                  <div className="flex flex-col gap-1.5">
                    <Label htmlFor="coord-lng">Longitude</Label>
                    <Input
                      id="coord-lng"
                      type="number"
                      step="any"
                      value={coordLng}
                      onChange={(e) => setCoordLng(e.target.value)}
                      placeholder="107.6191"
                      className="w-full sm:w-40"
                    />
                  </div>
                  <Button type="submit" disabled={!coordLat.trim() || !coordLng.trim() || reverse.isLoading}>
                    Cari Lokasi &amp; Batas
                  </Button>
                </div>
              </form>

              {reverse.isLoading && <p className="text-sm text-muted-foreground">Mencari lokasi…</p>}
              {reverse.error && (
                <p className="text-sm text-destructive">{(reverse.error as Error).message}</p>
              )}
              {reverse.data && (
                <div className="flex flex-col gap-1 text-sm">
                  <p className="font-medium">Lokasi koordinat:</p>
                  <p>
                    {reverse.data.kelurahan.nama}, {reverse.data.kecamatan.nama},{' '}
                    {reverse.data.kabupaten.nama}, {reverse.data.provinsi.nama}
                  </p>
                  <p className="text-xs text-muted-foreground">
                    Kode pos: {reverse.data.kodepos || '—'}
                  </p>
                  {reverseBoundary.isLoading && (
                    <p className="text-xs text-muted-foreground">Memuat batas area…</p>
                  )}
                  {reverseBoundary.data?.geometry ? (
                    <p className="text-xs font-medium">
                      Batas area: {reverseBoundary.data.nama} — ditampilkan
                    </p>
                  ) : (
                    !reverseBoundary.isLoading && (
                      <p className="text-xs text-muted-foreground">Belum ada data batas untuk area ini.</p>
                    )
                  )}
                </div>
              )}
            </CardContent>
          </Card>

          <div className="grid gap-4 lg:grid-cols-[minmax(0,1fr)_300px]">
            <div className="h-[62vh] overflow-hidden rounded-xl border border-border">
              <MapContainer center={JAKARTA_CENTER} zoom={5} className="h-full w-full">
                <TileLayer
                  attribution='&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a>'
                  url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png"
                />
                <ClickHandler onPick={addPoint} />
                {coordSubmitted && (
                  <Pin
                    lat={parseFloat(coordSubmitted.lat)}
                    lng={parseFloat(coordSubmitted.lng)}
                  />
                )}
                {boundary.data?.geometry && (
                  <>
                    <GeoJSON
                      key={selectedKode}
                      data={boundary.data as unknown as GeoJSON.Feature}
                      style={() => ({
                        color: '#0a7a5b',
                        weight: 2,
                        fillColor: '#10b981',
                        fillOpacity: 0.2,
                      })}
                    />
                    <FitPolygon feature={boundary.data} />
                  </>
                )}
                {reverseBoundary.data?.geometry && (
                  <>
                    <GeoJSON
                      key={`rev-${reverseBoundary.data.kode}`}
                      data={reverseBoundary.data as unknown as GeoJSON.Feature}
                      style={() => ({
                        color: '#1d4ed8',
                        weight: 2,
                        fillColor: '#3b82f6',
                        fillOpacity: 0.15,
                      })}
                    />
                    <FitPolygon feature={reverseBoundary.data} />
                  </>
                )}
                {token && <DistanceLayers points={points} routeCoords={routeCoords} />}
                {token && <FitBounds targets={fitTargets} />}
              </MapContainer>
            </div>

            {token ? (
              <DistanceSidebar
                points={points}
                hint={hint}
                ready={ready}
                mutation={mutation}
                onHitung={() => ready && mutation.mutate(points)}
                onRemove={removePoint}
                onReset={resetRoute}
              />
            ) : (
              <Card className="h-fit">
                <CardContent className="flex flex-col gap-2">
                  <p className="text-sm font-medium">Hitung Jarak via Jalan</p>
                  <p className="text-sm text-muted-foreground">
                    Klik beberapa titik pada peta untuk menghitung jarak dan waktu rute jalan.
                  </p>
                  <Link to="/login" className={cn(buttonVariants({ size: 'sm', className: 'mt-1 w-fit' }))}>
                    Masuk untuk menggunakan fitur hitung jarak
                  </Link>
                </CardContent>
              </Card>
            )}
          </div>
        </div>
      </main>
    </div>
  )
}