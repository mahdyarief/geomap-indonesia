package models

// BuildDetail assembles the JSON-ready fields from the raw scan helpers.
func (d *WilayahDetail) BuildDetail() {
	if d.ParentKode != "" {
		d.Parent = &ParentInfo{Kode: d.ParentKode, Nama: d.ParentNama}
	}
	if d.CentroidLat != nil && d.CentroidLng != nil {
		d.Centroid = &Centroid{Lat: *d.CentroidLat, Lng: *d.CentroidLng}
	}
	if d.PendudukTotal != nil {
		pria, wanita := 0, 0
		if d.PendudukPria != nil {
			pria = *d.PendudukPria
		}
		if d.PendudukWanita != nil {
			wanita = *d.PendudukWanita
		}
		d.Penduduk = &Penduduk{Total: *d.PendudukTotal, Pria: pria, Wanita: wanita}
	}
}

// BuildReverse assembles the JSON-ready centroid from the raw scan helpers.
func (r *ReverseResult) BuildReverse() {
	if r.CentroidLat != nil && r.CentroidLng != nil {
		r.Centroid = &Centroid{Lat: *r.CentroidLat, Lng: *r.CentroidLng}
	}
}

// BuildSearch assembles the JSON-ready parent from the raw scan helpers.
func (s *SearchResult) BuildSearch() {
	if s.ParentKode != "" {
		s.Parent = &ParentInfo{Kode: s.ParentKode, Nama: s.ParentNama}
	}
}
