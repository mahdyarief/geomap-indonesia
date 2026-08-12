import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { api } from '@/lib/api'
import type { ReverseResult } from '@/lib/types'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { MapPin } from 'lucide-react'

export default function ReversePage() {
  const [lat, setLat] = useState('')
  const [lng, setLng] = useState('')
  const [submitted, setSubmitted] = useState<{ lat: string; lng: string } | null>(null)

  const { data, isLoading, error } = useQuery({
    queryKey: ['reverse', submitted?.lat ?? null, submitted?.lng ?? null],
    queryFn: () =>
      api<ReverseResult>(
        `/api/v1/reverse?lat=${encodeURIComponent(submitted!.lat)}&lng=${encodeURIComponent(submitted!.lng)}`,
      ),
    enabled: !!submitted,
  })

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (!lat.trim() || !lng.trim()) return
    setSubmitted({ lat: lat.trim(), lng: lng.trim() })
  }

  return (
    <div className="flex flex-col gap-6">
      <div>
        <h1 className="text-xl font-semibold">Reverse Geocode</h1>
        <p className="mt-1 text-sm text-muted-foreground">
          Ubah koordinat (lat, lng) menjadi nama wilayah administratif dan kode pos.
        </p>
      </div>

      <form onSubmit={handleSubmit} className="flex max-w-xl flex-col gap-3">
        <div className="grid gap-3 sm:grid-cols-2">
          <div className="flex flex-col gap-1.5">
            <Label htmlFor="lat">Latitude</Label>
            <Input
              id="lat"
              type="number"
              step="any"
              value={lat}
              onChange={(e) => setLat(e.target.value)}
              placeholder="-6.9175"
              required
            />
          </div>
          <div className="flex flex-col gap-1.5">
            <Label htmlFor="lng">Longitude</Label>
            <Input
              id="lng"
              type="number"
              step="any"
              value={lng}
              onChange={(e) => setLng(e.target.value)}
              placeholder="107.6191"
              required
            />
          </div>
        </div>
        <div>
          <Button type="submit" disabled={!lat.trim() || !lng.trim() || isLoading}>
            <MapPin className="h-4 w-4" />
            Cari Wilayah
          </Button>
        </div>
      </form>

      {isLoading && <p className="text-sm text-muted-foreground">Mencari…</p>}
      {error && <p className="text-sm text-destructive">{(error as Error).message}</p>}

      {data && (
        <Card>
          <CardHeader>
            <CardTitle className="text-base">Hasil Reverse Geocode</CardTitle>
          </CardHeader>
          <CardContent className="flex flex-col gap-3 text-sm">
            <div>
              <p className="text-muted-foreground">Koordinat input</p>
              <p className="font-medium">
                {data.input.lat.toFixed(5)}, {data.input.lng.toFixed(5)}
              </p>
            </div>
            <div className="grid gap-3 sm:grid-cols-2">
              <div>
                <p className="text-muted-foreground">Provinsi</p>
                <p className="font-medium">{data.provinsi.nama}</p>
              </div>
              <div>
                <p className="text-muted-foreground">Kabupaten/Kota</p>
                <p className="font-medium">{data.kabupaten.nama}</p>
              </div>
              <div>
                <p className="text-muted-foreground">Kecamatan</p>
                <p className="font-medium">{data.kecamatan.nama}</p>
              </div>
              <div>
                <p className="text-muted-foreground">Kelurahan</p>
                <p className="font-medium">{data.kelurahan.nama}</p>
              </div>
            </div>
            <div>
              <p className="text-muted-foreground">Kode Pos</p>
              <p className="font-medium">{data.kodepos || '—'}</p>
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  )
}