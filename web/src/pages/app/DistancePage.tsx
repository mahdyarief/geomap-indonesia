import { MapContainer, TileLayer } from 'react-leaflet'
import { useDistanceRoute } from '@/lib/useDistanceRoute'
import { ClickHandler, DistanceLayers, FitBounds } from '@/components/distance/DistanceMapLayers'
import { DistanceSidebar } from '@/components/distance/DistanceSidebar'

const JAKARTA_CENTER: [number, number] = [-6.2, 106.8]

export default function DistancePage() {
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

  return (
    <div className="flex flex-col gap-4">
      <div>
        <h1 className="text-xl font-semibold">Hitung Jarak via Jalan</h1>
        <p className="mt-1 text-sm text-muted-foreground">
          Klik beberapa titik pada peta (urutan = rute). Total jarak dan waktu dihitung
          dari rangkaian rute jalan terdekat antar titik berturut-turut via pgRouting.
        </p>
      </div>

      <div className="grid gap-4 lg:grid-cols-[minmax(0,1fr)_300px]">
        <div className="h-[62vh] overflow-hidden rounded-xl border border-border">
          <MapContainer center={JAKARTA_CENTER} zoom={5} className="h-full w-full">
            <TileLayer
              attribution='&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a>'
              url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png"
            />
            <ClickHandler onPick={addPoint} />
            <DistanceLayers points={points} routeCoords={routeCoords} />
            <FitBounds targets={fitTargets} />
          </MapContainer>
        </div>

        <DistanceSidebar
          points={points}
          hint={hint}
          ready={ready}
          mutation={mutation}
          onHitung={() => ready && mutation.mutate(points)}
          onRemove={removePoint}
          onReset={resetRoute}
        />
      </div>
    </div>
  )
}