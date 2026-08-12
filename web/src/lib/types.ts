export interface Wilayah {
  kode: string
  nama: string
  type: string
}

export interface Parent {
  kode: string
  nama: string
}

export interface Centroid {
  lat: number
  lng: number
}

export interface Penduduk {
  total: number
  pria: number
  wanita: number
}

export interface WilayahDetail extends Wilayah {
  parent?: Parent
  centroid?: Centroid
  luas?: number
  penduduk?: Penduduk
  zona_waktu?: string
  elevasi?: number
}

export type SearchType = 'provinsi' | 'kabupaten' | 'kecamatan' | 'kelurahan'

export interface SearchResult extends Wilayah {
  parent?: Parent
  province?: string
}

export interface ReverseLevel {
  kode: string
  nama: string
}

export interface ReverseResult {
  input: Centroid
  provinsi: ReverseLevel
  kabupaten: ReverseLevel
  kecamatan: ReverseLevel
  kelurahan: ReverseLevel
  kodepos: string | null
  centroid: Centroid
}

export interface KodeposLookup {
  kodepos: string
  wilayah: {
    kode: string
    nama: string
    kecamatan: string
    kabupaten: string
    provinsi: string
  }
}

export interface KodeposByWilayah {
  kodepos: string[]
}

export interface Boundaries {
  type: string
  properties: Wilayah
  geometry: {
    type: string
    coordinates: unknown
  }
}

export interface AuthResponse {
  token: string
  token_type: string
  expires_in: number
}