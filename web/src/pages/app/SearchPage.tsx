import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Link } from 'react-router'
import { api } from '@/lib/api'
import type { SearchResult } from '@/lib/types'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Search } from 'lucide-react'

export default function SearchPage() {
  const [query, setQuery] = useState('')
  const [submitted, setSubmitted] = useState('')

  const { data, isLoading, error } = useQuery({
    queryKey: ['search', submitted],
    queryFn: () => api<SearchResult[]>(`/api/v1/search?q=${encodeURIComponent(submitted)}`),
    enabled: submitted.length > 0,
  })

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    setSubmitted(query.trim())
  }

  return (
    <div className="flex flex-col gap-6">
      <div>
        <h1 className="text-xl font-semibold">Pencarian Wilayah</h1>
        <p className="mt-1 text-sm text-muted-foreground">
          Cari provinsi, kabupaten, kecamatan, atau kelurahan berdasarkan nama.
        </p>
      </div>

      <form onSubmit={handleSubmit} className="flex max-w-xl flex-col gap-2">
        <Label htmlFor="q">Kata kunci</Label>
        <div className="flex gap-2">
          <Input
            id="q"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="contoh: Bandung, Jawa Barat, Kuta…"
            autoComplete="off"
          />
          <Button type="submit" disabled={!query.trim() || isLoading}>
            <Search className="h-4 w-4" />
            Cari
          </Button>
        </div>
      </form>

      {isLoading && <p className="text-sm text-muted-foreground">Mencari…</p>}
      {error && <p className="text-sm text-destructive">{(error as Error).message}</p>}

      {submitted && !isLoading && !error && data && data.length === 0 && (
        <p className="text-sm text-muted-foreground">Tidak ada hasil untuk “{submitted}”.</p>
      )}

      {data && data.length > 0 && (
        <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
          {data.map((r) => (
            <Link key={r.kode} to={`/app/wilayah/${r.kode}`}>
              <Card className="transition-colors hover:border-ring">
                <CardHeader className="flex flex-row items-center justify-between gap-2">
                  <CardTitle className="truncate">{r.nama}</CardTitle>
                  <Badge variant="muted">{r.type}</Badge>
                </CardHeader>
                <CardContent className="text-sm text-muted-foreground">
                  {r.province || (r.parent ? r.parent.nama : '')} • Kode {r.kode}
                </CardContent>
              </Card>
            </Link>
          ))}
        </div>
      )}
    </div>
  )
}