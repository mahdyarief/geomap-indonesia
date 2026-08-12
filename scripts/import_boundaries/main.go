package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	"github.com/jackc/pgx/v5"

	"github.com/mahdyarief/geomap-indonesia/internal/importer"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("import boundaries: %v", err)
	}
}

func run() error {
	ctx := context.Background()
	pool, err := importer.ConnectPool(ctx)
	if err != nil {
		return err
	}
	defer pool.Close()

	dir := importer.DataDir() + "/wilayah_boundaries"
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		log.Printf("cloning cahyadsn/wilayah_boundaries into %s ...", dir)
		cmd := exec.Command("git", "clone", "--depth", "1",
			"https://github.com/cahyadsn/wilayah_boundaries.git", dir)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("clone boundaries: %w", err)
		}
	}

	files, err := filepath.Glob(filepath.Join(dir, "db", "*", "*.sql"))
	if err != nil {
		return err
	}
	log.Printf("boundary files: %d", len(files))

	total := 0
	for _, f := range files {
		rows, err := importer.ParseSQLRows(f)
		if err != nil {
			return err
		}
		batch := &pgx.Batch{}
		n := 0
		for _, r := range rows {
			if len(r) < 5 {
				continue
			}
			kode := importer.CleanKode(r[0])
			table, ok := tableForKode(kode)
			if !ok {
				continue
			}
			lat, err1 := strconv.ParseFloat(r[2], 64)
			lng, err2 := strconv.ParseFloat(r[3], 64)
			if err1 != nil || err2 != nil {
				continue
			}
			gj, err := pathToGeoJSON(r[4])
			if err != nil {
				return fmt.Errorf("%s kode %s: %w", f, kode, err)
			}
			batch.Queue(fmt.Sprintf(
				"UPDATE %s SET geometry = ST_SetSRID(ST_GeomFromGeoJSON($2), 4326), lat = $3, lng = $4 WHERE kode = $1",
				table), kode, gj, lat, lng)
			n++
		}
		if n > 0 {
			if err := importer.ExecBatch(ctx, pool, batch); err != nil {
				return fmt.Errorf("%s: %w", f, err)
			}
			total += n
			log.Printf("%s: %d rows", filepath.Base(f), n)
		}
	}
	log.Printf("boundaries: %d rows updated", total)
	return nil
}

func tableForKode(kode string) (string, bool) {
	switch len(kode) {
	case 2:
		return "provinsi", true
	case 4:
		return "kabupaten", true
	case 6:
		return "kecamatan", true
	case 10:
		return "kelurahan", true
	default:
		return "", false
	}
}

// pathToGeoJSON converts the boundary path (a JSON array of polygons in
// [lat,lng] order) into a GeoJSON MultiPolygon string with [lng,lat] order
// suitable for ST_GeomFromGeoJSON. Single polygons are wrapped.
func pathToGeoJSON(path string) (string, error) {
	var multi [][][][]float64
	if err := json.Unmarshal([]byte(path), &multi); err != nil {
		var poly [][][]float64
		if err2 := json.Unmarshal([]byte(path), &poly); err2 != nil {
			return "", err
		}
		multi = [][][][]float64{poly}
	}
	for _, polygons := range multi {
		for _, ring := range polygons {
			for _, pt := range ring {
				pt[0], pt[1] = pt[1], pt[0]
			}
		}
	}
	b, err := json.Marshal(map[string]any{"type": "MultiPolygon", "coordinates": multi})
	if err != nil {
		return "", err
	}
	return string(b), nil
}
