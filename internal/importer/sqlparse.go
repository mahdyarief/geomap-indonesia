package importer

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// maxLine is the largest single line we expect to encounter. Boundary
// geometries for a single wilayah can be hundreds of KB on one line.
const maxLine = 64 << 20 // 64 MiB

// ParseSQLRows reads a MySQL dump file and returns every value row from all
// INSERT statements it contains. Each row is a slice of raw (unquoted) fields.
// The parser is line-based: each data row must live on a single line, which
// holds for every file in the cahyadsn data set except wilayah_level_1_2.sql
// (handled separately by regexp).
func ParseSQLRows(path string) ([][]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 1<<20), maxLine)

	var rows [][]string
	inValues := false
	first := true

	for scanner.Scan() {
		line := strings.TrimSuffix(scanner.Text(), "\r")
		if first {
			line = strings.TrimPrefix(line, "\ufeff") // strip UTF-8 BOM
			first = false
		}
		trimmed := strings.TrimSpace(line)

		switch {
		case strings.HasPrefix(trimmed, "INSERT INTO"):
			inValues = strings.Contains(trimmed, "VALUES")
		case strings.EqualFold(trimmed, "VALUES"):
			inValues = true
		case !inValues:
			// comment, DDL, SET NAMES, LOCK/UNLOCK, etc.
		case trimmed == "" || strings.HasPrefix(trimmed, "--"):
			inValues = false
		case strings.HasPrefix(trimmed, "("):
			fields, err := splitRow(trimmed)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", path, err)
			}
			rows = append(rows, fields)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan %s: %w", path, err)
	}
	return rows, nil
}

// splitRow parses one SQL value tuple such as
//
//	('11.01.40001','Pulau X','-3.3','97.1','TBP',0.0006,''),
//
// stripping the trailing comma/semicolon and the outer parentheses, then
// splitting on commas that are not inside single quotes. Doubled quotes ('')
// inside a string are unescaped.
func splitRow(s string) ([]string, error) {
	s = strings.TrimSpace(s)
	s = strings.TrimSuffix(s, ";")
	s = strings.TrimSuffix(s, ",")
	if !strings.HasPrefix(s, "(") || !strings.HasSuffix(s, ")") {
		return nil, fmt.Errorf("invalid row: %.80s", s)
	}
	s = s[1 : len(s)-1]

	var fields []string
	var cur strings.Builder
	inQuote := false
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case c == '\'':
			if inQuote && i+1 < len(s) && s[i+1] == '\'' {
				cur.WriteByte('\'')
				i++
			} else {
				inQuote = !inQuote
			}
		case c == ',' && !inQuote:
			fields = append(fields, unquote(cur.String()))
			cur.Reset()
		default:
			cur.WriteByte(c)
		}
	}
	if inQuote {
		return nil, fmt.Errorf("unterminated quote in row: %.80s", s)
	}
	fields = append(fields, unquote(cur.String()))
	return fields, nil
}

// unquote strips the surrounding single quotes from a SQL literal and
// converts doubled quotes back to a single quote. Numeric literals and NULL
// are returned as-is (trimmed).
func unquote(s string) string {
	s = strings.TrimSpace(s)
	if len(s) >= 2 && s[0] == '\'' && s[len(s)-1] == '\'' {
		return strings.ReplaceAll(s[1:len(s)-1], "''", "'")
	}
	return s
}

// CleanKode removes the dots from a wilayah kode ("11.01.01.2001" ->
// "1101012001") and trims whitespace.
func CleanKode(kode string) string {
	return strings.ReplaceAll(strings.TrimSpace(kode), ".", "")
}

// DataDir returns the data directory from the DATA_DIR environment variable,
// defaulting to "./data" (matching .env.example).
func DataDir() string {
	if d := strings.TrimSpace(os.Getenv("DATA_DIR")); d != "" {
		return d
	}
	return "./data"
}