import { MapContainer, TileLayer, Marker, Popup, useMapEvents } from 'react-leaflet'
import { Link } from 'react-router'
import { useAuthStore } from '@/store/auth'
import { Button, buttonVariants } from '@/components/ui/button'
import { useState } from 'react'
import { Card, CardContent } from '@/components/ui/card'
import { cn } from '@/lib/utils'

const INDONESIA_CENTER: [number, number] = [-2.5, 118]

function ClickHandler({ onPick }: { onPick: (lat: number, lng: number) => void }) {
  useMapEvents({
    click: (e) => onPick(e.latlng.lat, e.latlng.lng),
  })
  return null
}

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

export default function HomePage() {
  const token = useAuthStore((s) => s.token)
  const [pick, setPick] = useState<{ lat: number; lng: number } | null>(null)
  const [count, setCount] = useState(0)

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
            Klik pada peta untuk memilih koordinat. Gunakan API key untuk eksplorasi wilayah dan reverse geocode.
          </p>
        </div>

        <div className="mx-auto mb-6 max-w-6xl px-4">
          <div className="h-[60vh] overflow-hidden rounded-xl border border-border">
            <MapContainer center={INDONESIA_CENTER} zoom={4} className="h-full w-full">
              <TileLayer
                attribution='&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a>'
                url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png"
              />
              <ClickHandler onPick={(lat, lng) => setPick({ lat, lng })} />
              {pick && <Pin lat={pick.lat} lng={pick.lng} />}
            </MapContainer>
          </div>
        </div>

        {pick && (
          <div className="mx-auto max-w-6xl px-4 pb-6">
            <Card>
              <CardContent className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
                <div>
                  <p className="text-sm font-medium">Koordinat terpilih</p>
                  <p className="text-sm text-muted-foreground">
                    lat {pick.lat.toFixed(5)}, lng {pick.lng.toFixed(5)}
                  </p>
                </div>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => setCount((c) => c + 1)}
                >
                  Salin ({count})
                </Button>
              </CardContent>
            </Card>
          </div>
        )}
      </main>
    </div>
  )
}