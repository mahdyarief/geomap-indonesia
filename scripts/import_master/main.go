package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/mahdyarief/geomap-indonesia/internal/importer"
)

// reLevel12 matches the non-geometry fields of wilayah_level_1_2.sql rows:
// ('11','Aceh','Banda Aceh', 5.57..., 95.34..., 11, 7, 56835.019, 5623479, '[[[...]]]', 1)
var reLevel12 = regexp.MustCompile(`\('(\d+)','[^']*','[^']*',\s*([-0-9.]+),\s*([-0-9.]+),\s*(\d+),\s*(\d+),`)

func main() {
	if err := run(); err != nil {
		log.Fatalf("import master: %v", err)
	}
}

func run() error {
	ctx := context.Background()
	pool, err := importer.ConnectPool(ctx)
	if err != nil {
		return err
	}
	defer pool.Close()

	dbDir := importer.DataDir() + "/wilayah/db"

	if _, err := pool.Exec(ctx, "TRUNCATE provinsi, kabupaten, kecamatan, kelurahan, pulau, kodepos RESTART IDENTITY CASCADE"); err != nil {
		return fmt.Errorf("truncate: %w", err)
	}

	if err := importWilayah(ctx, pool, dbDir+"/wilayah.sql"); err != nil {
		return err
	}
	if err := updatePenduduk(ctx, pool, dbDir+"/wilayah_penduduk.sql"); err != nil {
		return err
	}
	if err := updateLuas(ctx, pool, dbDir+"/wilayah_luas.sql"); err != nil {
		return err
	}
	if err := importPulau(ctx, pool, dbDir+"/wilayah_pulau.sql"); err != nil {
		return err
	}
	return updateLevel12(ctx, pool, dbDir+"/wilayah_level_1_2.sql")
}

func importWilayah(ctx context.Context, pool *pgxpool.Pool, path string) error {
	rows, err := importer.ParseSQLRows(path)
	if err != nil {
		return err
	}
	var prov, kab, kec, kel [][]any
	for _, r := range rows {
		if len(r) < 2 {
			continue
		}
		kode := importer.CleanKode(r[0])
		nama := r[1]
		switch len(kode) {
		case 2:
			prov = append(prov, []any{kode, nama})
		case 4:
			kab = append(kab, []any{kode, nama, kode[:2]})
		case 6:
			kec = append(kec, []any{kode, nama, kode[:4]})
		case 10:
			kel = append(kel, []any{kode, nama, kode[:6]})
		}
	}
	log.Printf("master: provinsi=%d kabupaten=%d kecamatan=%d kelurahan=%d",
		len(prov), len(kab), len(kec), len(kel))

	for _, step := range []struct {
		table string
		cols  []string
		rows  [][]any
	}{
		{"provinsi", []string{"kode", "nama"}, prov},
		{"kabupaten", []string{"kode", "nama", "provinsi_kode"}, kab},
		{"kecamatan", []string{"kode", "nama", "kabupaten_kode"}, kec},
		{"kelurahan", []string{"kode", "nama", "kecamatan_kode"}, kel},
	} {
		if err := importer.CopyTable(ctx, pool, step.table, step.cols, step.rows); err != nil {
			return fmt.Errorf("copy %s: %w", step.table, err)
		}
	}
	return nil
}

func updatePenduduk(ctx context.Context, pool *pgxpool.Pool, path string) error {
	rows, err := importer.ParseSQLRows(path)
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
		pria, err1 := strconv.Atoi(r[2])
		wanita, err2 := strconv.Atoi(r[3])
		total, err3 := strconv.Atoi(r[4])
		if err1 != nil || err2 != nil || err3 != nil {
			continue
		}
		switch len(kode) {
		case 2:
			batch.Queue("UPDATE provinsi SET penduduk_total=$2, penduduk_pria=$3, penduduk_wanita=$4 WHERE kode=$1", kode, total, pria, wanita)
			n++
		case 4:
			batch.Queue("UPDATE kabupaten SET penduduk_total=$2, penduduk_pria=$3, penduduk_wanita=$4 WHERE kode=$1", kode, total, pria, wanita)
			n++
		}
	}
	if err := importer.ExecBatch(ctx, pool, batch); err != nil {
		return err
	}
	log.Printf("penduduk: %d rows updated", n)
	return nil
}

func updateLuas(ctx context.Context, pool *pgxpool.Pool, path string) error {
	rows, err := importer.ParseSQLRows(path)
	if err != nil {
		return err
	}
	batch := &pgx.Batch{}
	n := 0
	for _, r := range rows {
		if len(r) < 3 {
			continue
		}
		kode := importer.CleanKode(r[0])
		if r[2] == "NULL" || r[2] == "" {
			continue
		}
		luas, err := strconv.ParseFloat(r[2], 64)
		if err != nil {
			continue
		}
		switch len(kode) {
		case 2:
			batch.Queue("UPDATE provinsi SET luas=$2 WHERE kode=$1", kode, luas)
			n++
		case 4:
			batch.Queue("UPDATE kabupaten SET luas=$2 WHERE kode=$1", kode, luas)
			n++
		}
	}
	if err := importer.ExecBatch(ctx, pool, batch); err != nil {
		return err
	}
	log.Printf("luas: %d rows updated", n)
	return nil
}

func importPulau(ctx context.Context, pool *pgxpool.Pool, path string) error {
	rows, err := importer.ParseSQLRows(path)
	if err != nil {
		return err
	}
	data := make([][]any, 0, len(rows))
	seen := make(map[string]bool, len(rows))
	for _, r := range rows {
		if len(r) < 7 {
			continue
		}
		kode := importer.CleanKode(r[0])
		if seen[kode] {
			continue
		}
		seen[kode] = true
		lat, err1 := strconv.ParseFloat(r[2], 64)
		lng, err2 := strconv.ParseFloat(r[3], 64)
		if err1 != nil || err2 != nil {
			continue
		}
		var luas any
		if r[5] != "NULL" && r[5] != "" {
			if f, err := strconv.ParseFloat(r[5], 64); err == nil {
				luas = f
			}
		}
		data = append(data, []any{kode, r[1], lat, lng, luas, r[4], r[6]})
	}
	log.Printf("pulau: %d rows", len(data))
	return importer.CopyTable(ctx, pool, "pulau",
		[]string{"kode", "nama", "lat", "lng", "luas", "status", "notes"}, data)
}

func updateLevel12(ctx context.Context, pool *pgxpool.Pool, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	matches := reLevel12.FindAllSubmatch(data, -1)
	batch := &pgx.Batch{}
	n := 0
	for _, m := range matches {
		kode := string(m[1])
		lat, err1 := strconv.ParseFloat(string(m[2]), 64)
		lng, err2 := strconv.ParseFloat(string(m[3]), 64)
		elv, err3 := strconv.Atoi(string(m[4]))
		tz, err4 := strconv.Atoi(string(m[5]))
		if err1 != nil || err2 != nil || err3 != nil || err4 != nil {
			continue
		}
		switch len(kode) {
		case 2:
			batch.Queue("UPDATE provinsi SET lat=$2, lng=$3, elevasi=$4, zona_waktu=$5 WHERE kode=$1", kode, lat, lng, elv, zonaWaktu(tz))
			n++
		case 4:
			batch.Queue("UPDATE kabupaten SET lat=$2, lng=$3, elevasi=$4, zona_waktu=$5 WHERE kode=$1", kode, lat, lng, elv, zonaWaktu(tz))
			n++
		}
	}
	if err := importer.ExecBatch(ctx, pool, batch); err != nil {
		return err
	}
	log.Printf("level_1_2: %d rows updated", n)
	return nil
}

func zonaWaktu(tz int) string {
	switch tz {
	case 7:
		return "WIB"
	case 8:
		return "WITA"
	case 9:
		return "WIT"
	default:
		return fmt.Sprintf("UTC+%d", tz)
	}
}
