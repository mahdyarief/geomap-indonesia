import type { UseMutationResult } from '@tanstack/react-query'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { Label } from '@/components/ui/label'
import type { Centroid } from '@/lib/types'
import type { RouteSummary } from '@/lib/useDistanceRoute'
import { coordLabel } from '@/lib/useDistanceRoute'

type Props = {
  points: Centroid[]
  hint: string
  ready: boolean
  mutation: UseMutationResult<RouteSummary, Error, Centroid[], unknown>
  onHitung: () => void
  onRemove: (idx: number) => void
  onReset: () => void
}

export function DistanceSidebar({ points, hint, ready, mutation, onHitung, onRemove, onReset }: Props) {
  return (
    <Card className="h-fit">
      <CardContent className="flex flex-col gap-4">
        <div className="flex items-center justify-between">
          <Label>Titik Terpilih</Label>
          <span className="text-xs text-muted-foreground">{points.length} titik</span>
        </div>

        {points.length === 0 ? (
          <p className="text-sm text-muted-foreground">Klik peta untuk menambahkan titik pertama.</p>
        ) : (
          <ol className="flex max-h-48 flex-col gap-1.5 overflow-y-auto pr-1">
            {points.map((c, i) => (
              <li key={i} className="flex items-center justify-between gap-2 text-sm">
                <div className="flex min-w-0 items-center gap-2">
                  <span className="flex h-5 w-5 shrink-0 items-center justify-center rounded-full bg-emerald-700 text-[11px] font-semibold text-white">
                    {i + 1}
                  </span>
                  <span className="truncate text-muted-foreground">{coordLabel(c)}</span>
                </div>
                <button
                  type="button"
                  onClick={() => onRemove(i)}
                  className="shrink-0 text-xs text-destructive hover:underline"
                >
                  Hapus
                </button>
              </li>
            ))}
          </ol>
        )}

        {hint && <p className="text-sm text-destructive">{hint}</p>}
        {mutation.isError && (
          <p className="text-sm text-destructive">{(mutation.error as Error).message}</p>
        )}

        <div className="flex flex-wrap items-center gap-2">
          <Button onClick={onHitung} disabled={!ready || mutation.isPending}>
            {mutation.isPending ? 'Menghitung…' : ready ? 'Hitung Jarak' : 'Tambahkan minimal 2 titik'}
          </Button>
          <Button type="button" variant="outline" onClick={onReset} disabled={points.length === 0 && !mutation.data}>
            Reset
          </Button>
        </div>

        {mutation.data && (
          <div className="flex flex-col gap-1 text-sm">
            <p className="font-medium">Hasil rute:</p>
            <p>
              Total jarak: <span className="font-semibold">{mutation.data.totalKm.toFixed(2)} km</span>
            </p>
            <p>
              Estimasi waktu: <span className="font-semibold">{mutation.data.totalMin.toFixed(0)} menit</span>
            </p>
          </div>
        )}

        <p className="text-xs text-muted-foreground">Titik di luar batas Indonesia akan diabaikan.</p>
      </CardContent>
    </Card>
  )
}