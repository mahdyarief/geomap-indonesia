import { useQuery } from '@tanstack/react-query'
import { Link } from 'react-router'
import { api } from '@/lib/api'
import type { Wilayah } from '@/lib/types'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'

export default function DashboardPage() {
  const { data, isLoading, error } = useQuery({
    queryKey: ['wilayah'],
    queryFn: () => api<Wilayah[]>('/api/v1/wilayah'),
  })

  return (
    <div className="flex flex-col gap-6">
      <div>
        <h1 className="text-xl font-semibold">Dashboard</h1>
        <p className="mt-1 text-sm text-muted-foreground">
          Daftar provinsi di Indonesia. Pilih satu untuk menelusuri kabupaten, kecamatan, hingga kelurahan.
        </p>
      </div>

      {isLoading && <p className="text-sm text-muted-foreground">Memuat data…</p>}
      {error && (
        <p className="text-sm text-destructive">
          {(error as Error).message}
        </p>
      )}

      {data && (
        <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
          {data.map((prov) => (
            <Link key={prov.kode} to={`/app/wilayah/${prov.kode}`}>
              <Card className="transition-colors hover:border-ring">
                <CardHeader className="flex flex-row items-center justify-between gap-2">
                  <CardTitle className="truncate">{prov.nama}</CardTitle>
                  <Badge variant="muted">{prov.type}</Badge>
                </CardHeader>
                <CardContent className="text-sm text-muted-foreground">
                  Kode {prov.kode}
                </CardContent>
              </Card>
            </Link>
          ))}
        </div>
      )}

      {data && data.length === 0 && !isLoading && (
        <p className="text-sm text-muted-foreground">Belum ada data wilayah.</p>
      )}
    </div>
  )
}