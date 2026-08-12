package models

// WilayahType identifies the administrative level of a wilayah entry.
type WilayahType string

const (
	TypeProvinsi  WilayahType = "provinsi"
	TypeKabupaten WilayahType = "kabupaten"
	TypeKecamatan WilayahType = "kecamatan"
	TypeKelurahan WilayahType = "kelurahan"
)

// ParentInfo is a compact reference to a parent wilayah.
type ParentInfo struct {
	Kode string `json:"kode"`
	Nama string `json:"nama"`
}

// Centroid holds a point coordinate (lat/lng).
type Centroid struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

// Penduduk holds population statistics.
type Penduduk struct {
	Total  int `json:"total"`
	Pria   int `json:"pria"`
	Wanita int `json:"wanita"`
}

// WilayahListItem is a single row in a wilayah list response.
type WilayahListItem struct {
	Kode string      `json:"kode"`
	Nama string      `json:"nama"`
	Type WilayahType `json:"type"`
}

// WilayahDetail is the full detail response for GET /wilayah/:kode.
type WilayahDetail struct {
	Kode      string      `json:"kode"`
	Nama      string      `json:"nama"`
	Type      WilayahType `json:"type"`
	Parent    *ParentInfo `json:"parent,omitempty"`
	Centroid  *Centroid   `json:"centroid,omitempty"`
	Luas      *float64    `json:"luas,omitempty"`
	Penduduk  *Penduduk   `json:"penduduk,omitempty"`
	LogoURL   *string     `json:"logo_url,omitempty"`
	ZonaWaktu *string     `json:"zona_waktu,omitempty"`
	Elevasi   *int        `json:"elevasi,omitempty"`

	// Scan helpers (not serialized).
	CentroidLat    *float64 `json:"-"`
	CentroidLng    *float64 `json:"-"`
	PendudukTotal  *int     `json:"-"`
	PendudukPria   *int     `json:"-"`
	PendudukWanita *int     `json:"-"`
	ParentKode     string   `json:"-"`
	ParentNama     string   `json:"-"`
}

// ReverseResult is the response for reverse geocoding.
type ReverseResult struct {
	Input     Centroid    `json:"input"`
	Provinsi  ParentInfo  `json:"provinsi"`
	Kabupaten ParentInfo  `json:"kabupaten"`
	Kecamatan ParentInfo  `json:"kecamatan"`
	Kelurahan ParentInfo  `json:"kelurahan"`
	Kodepos   []string    `json:"kodepos"`
	Centroid  *Centroid   `json:"centroid,omitempty"`

	// Scan helpers (not serialized).
	CentroidLat *float64 `json:"-"`
	CentroidLng *float64 `json:"-"`
}

// SearchResult is a single match from GET /search.
type SearchResult struct {
	Kode     string      `json:"kode"`
	Nama     string      `json:"nama"`
	Type     WilayahType `json:"type"`
	Parent   *ParentInfo `json:"parent,omitempty"`
	Province string      `json:"province,omitempty"`

	// Scan helpers (not serialized).
	ParentKode string `json:"-"`
	ParentNama string `json:"-"`
}

// HierarchyNode is a node in the hierarchy tree.
type HierarchyNode struct {
	Kode     string          `json:"kode"`
	Nama     string          `json:"nama"`
	Type     WilayahType     `json:"type"`
	Parent   *ParentInfo     `json:"parent,omitempty"`
	Children []WilayahDetail `json:"children,omitempty"`
}

// KodeposWilayah is the wilayah information attached to a kodepos.
type KodeposWilayah struct {
	Kode      string `json:"kode"`
	Nama      string `json:"nama"`
	Kecamatan string `json:"kecamatan"`
	Kabupaten string `json:"kabupaten"`
	Provinsi  string `json:"provinsi"`
}

// KodeposLookup is the response for GET /kodepos/:kode.
type KodeposLookup struct {
	Kodepos string         `json:"kodepos"`
	Wilayah KodeposWilayah `json:"wilayah"`
}

// KodeposByWilayah is the response for GET /kodepos?wilayah=...
type KodeposByWilayah struct {
	Kodepos []string `json:"kodepos"`
}

// BatchReverseRequest is the body of POST /batch/reverse.
type BatchReverseRequest struct {
	Points []Centroid `json:"points"`
}

// BatchReverseResult pairs an input point with its reverse-geocoding result.
type BatchReverseResult struct {
	Input  Centroid      `json:"input"`
	Result *ReverseResult `json:"result,omitempty"`
	Error  *string       `json:"error,omitempty"`
}

// Pagination describes pagination metadata in list responses.
type Pagination struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}
