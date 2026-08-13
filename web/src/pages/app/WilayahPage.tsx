import { useQuery } from '@tanstack/react-query'
import { Link, useParams } from 'react-router'
import { api } from '@/lib/api'
import type { Wilayah, WilayahDetail } from '@/lib/types'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { ArrowLeft } from 'lucide-react'

export default function WilayahPage() {
  const { kode } = useParams<{ kode: string }>()

  const detail = useQuery({
    queryKey: ['wilayah', kode],
    queryFn: () => api<WilayahDetail>(`/api/v1/wilayah/${kode}`),
    enabled: !!kode,
  })

  const children = useQuery({
    queryKey: ['wilayah', kode, 'children'],
    queryFn: () => api<Wilayah[]>(`/api/v1/wilayah/${kode}/children`),
    enabled: !!kode,
  })

  const provinces = useQuery({
    queryKey: ['wilayah', 'provinsi'],
    queryFn: () => api<Wilayah[]>('/api/v1/wilayah?limit=100'),
    enabled: !kode,
  })

  if (!kode) {
    return (
      <div className="flex flex-col gap-6">
        <div>
          <h1 className="text-xl font-semibold">Wilayah</h1>
          <p className="mt-1 text-sm text-muted-foreground">
            Pilih provinsi untuk melihat detail dan wilayah turunannya.
          </p>
        </div>

        {provinces.isLoading && <p className="text-sm text-muted-foreground">Memuat…</p>}
        {provinces.error && (
          <p className="text-sm text-destructive">{(provinces.error as Error).message}</p>
        )}
        {provinces.data && provinces.data.length === 0 && (
          <p className="text-sm text-muted-foreground">Tidak ada data wilayah.</p>
        )}
        {provinces.data && provinces.data.length > 0 && (
          <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
            {provinces.data.map((p) => (
              <Link key={p.kode} to={`/app/wilayah/${p.kode}`}>
                <Card className="transition-colors hover:border-ring">
                  <CardHeader className="flex flex-row items-center justify-between gap-2">
                    <CardTitle className="truncate">{p.nama}</CardTitle>
                    <Badge variant="muted">{p.type}</Badge>
                  </CardHeader>
                  <CardContent className="text-sm text-muted-foreground">
                    Kode {p.kode}
                  </CardContent>
                </Card>
              </Link>
            ))}
          </div>
        )}
      </div>
    )
  }

  if (detail.isLoading) {
    return <p className="text-sm text-muted-foreground">Memuat detail…</p>
  }

  if (detail.error || !detail.data) {
    return (
      <div className="flex flex-col gap-3">
        <p className="text-sm text-destructive">
          {(detail.error as Error)?.message || 'Wilayah tidak ditemukan.'}
        </p>
        <Link to="/app/wilayah" className="text-sm text-primary underline underline-offset-4">
          Kembali ke daftar wilayah
        </Link>
      </div>
    )
  }

  const w = detail.data

  return (
    <div className="flex flex-col gap-6">
      <div>
        <Link
          to="/app/dashboard"
          className="inline-flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground"
        >
          <ArrowLeft className="h-4 w-4" />
          Kembali
        </Link>
        <h1 className="mt-2 text-xl font-semibold">
          {w.nama} <Badge className="ml-2 align-middle" variant="muted">{w.type}</Badge>
        </h1>
        <p className="mt-1 text-sm text-muted-foreground">
          Kode {w.kode}
          {w.parent ? ` • Bagian dari ${w.parent.nama}` : ''}
        </p>
      </div>

      <Card>
        <CardHeader>
          <CardTitle className="text-base">Informasi Wilayah</CardTitle>
        </CardHeader>
        <CardContent className="grid gap-3 text-sm sm:grid-cols-2 lg:grid-cols-3">
          <div>
            <p className="text-muted-foreground">Luas</p>
            <p className="font-medium">{w.luas != null ? `${w.luas.toLocaleString()} km²` : '—'}</p>
          </div>
          <div>
            <p className="text-muted-foreground">Zona Waktu</p>
            <p className="font-medium">{w.zona_waktu || '—'}</p>
          </div>
          <div>
            <p className="text-muted-foreground">Elevasi</p>
            <p className="font-medium">{w.elevasi != null ? `${w.elevasi} m` : '—'}</p>
          </div>
          <div>
            <p className="text-muted-foreground">Penduduk</p>
            <p className="font-medium">
              {w.penduduk ? w.penduduk.total.toLocaleString('id-ID') : '—'}
            </p>
          </div>
          <div>
            <p className="text-muted-foreground">Centroid</p>
            <p className="font-medium">
              {w.centroid
                ? `${w.centroid.lat.toFixed(4)}, ${w.centroid.lng.toFixed(4)}`
                : '—'}
            </p>
          </div>
        </CardContent>
      </Card>

      <div>
        <h2 className="mb-3 text-sm font-semibold uppercase tracking-wide text-muted-foreground">
          Wilayah Turunan
        </h2>
        {children.isLoading && (
          <p className="text-sm text-muted-foreground">Memuat…</p>
        )}
        {children.error && (
          <p className="text-sm text-destructive">{(children.error as Error).message}</p>
        )}
        {children.data && children.data.length === 0 && (
          <p className="text-sm text-muted-foreground">Tidak ada wilayah turunan.</p>
        )}
        {children.data && children.data.length > 0 && (
          <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
            {children.data.map((child) => (
              <Link key={child.kode} to={`/app/wilayah/${child.kode}`}>
                <Card className="transition-colors hover:border-ring">
                  <CardHeader className="flex flex-row items-center justify-between gap-2">
                    <CardTitle className="truncate">{child.nama}</CardTitle>
                    <Badge variant="muted">{child.type}</Badge>
                  </CardHeader>
                  <CardContent className="text-sm text-muted-foreground">
                    Kode {child.kode}
                  </CardContent>
                </Card>
              </Link>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}