package main

import (
	"context"
	"log"

	"github.com/mahdyarief/geomap-indonesia/internal/importer"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("import kodepos: %v", err)
	}
}

func run() error {
	ctx := context.Background()
	pool, err := importer.ConnectPool(ctx)
	if err != nil {
		return err
	}
	defer pool.Close()

	path := importer.DataDir() + "/wilayah_kodepos/db/wilayah_kodepos.sql"
	rows, err := importer.ParseSQLRows(path)
	if err != nil {
		return err
	}

	// The kodepos PK is the 5-digit kode; many kelurahan share the same
	// kodepos, so keep the first mapping only (first-wins).
	seen := make(map[string]string, len(rows))
	for _, r := range rows {
		if len(r) < 2 {
			continue
		}
		kodepos := r[1]
		if _, ok := seen[kodepos]; !ok {
			seen[kodepos] = importer.CleanKode(r[0])
		}
	}

	data := make([][]any, 0, len(seen))
	for kodepos, kelurahan := range seen {
		data = append(data, []any{kodepos, kelurahan})
	}
	log.Printf("kodepos: %d unique of %d rows", len(data), len(rows))
	return importer.CopyTable(ctx, pool, "kodepos",
		[]string{"kode", "kelurahan_kode"}, data)
}
