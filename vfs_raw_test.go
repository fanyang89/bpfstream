package main

import (
	"bytes"
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/marcboeker/go-duckdb/v2"
	"github.com/rs/zerolog/log"
)

const testDataFile = "testdata/vfs-raw.ndjson"
const rows = 321039

// linux-amd64, cpu: AMD Ryzen 7 9700X 8-Core Processor

// Read from memory and parse only
// ops=1109877
func BenchmarkParseOnly(b *testing.B) {
	buffer, err := os.ReadFile(testDataFile)
	if err != nil {
		b.Fatal(err)
	}

	now := time.Now()

	for n := 0; n < b.N; n++ {
		r := bytes.NewReader(buffer)
		err = writeToDuckDB(r, func(...driver.Value) error { return nil })
		if err != nil {
			b.Fatal(err)
		}
	}

	elapsed := time.Now().Sub(now)
	log.Info().
		Int("n", b.N).
		Int64("elapsed_ns", elapsed.Nanoseconds()).
		Int64("ops", 1000000000/(elapsed.Nanoseconds()/4/rows)).
		Msg("Done in time")
}

// Read from memory and write to DuckDB
// ops=1172332
func BenchmarkReadMemoryWriteDuckDB(b *testing.B) {
	ctx := context.Background()
	dsn := "bench.ddb"
	tableName := "append_bench"

	_ = os.Remove(dsn)
	_ = os.Remove(dsn + ".wal")

	connector, err := duckdb.NewConnector(dsn, nil)
	if err != nil {
		b.Fatal(err)
	}

	log.Info().Str("dsn", dsn).Msg("Connecting to db")
	conn, err := connector.Connect(ctx)
	if err != nil {
		b.Fatal(err)
	}

	db := sql.OpenDB(connector)
	log.Info().Msg("DB created")

	_, err = db.Exec(fmt.Sprintf(dropTableSql, tableName))
	if err != nil {
		b.Fatal(err)
	}

	_, err = db.Exec(fmt.Sprintf(createTableSql, tableName))
	if err != nil {
		b.Fatal(err)
	}

	appender, err := duckdb.NewAppenderFromConn(conn, "", tableName)
	if err != nil {
		b.Fatal(err)
	}
	defer func() { _ = appender.Close() }()

	buffer, err := os.ReadFile(testDataFile)
	if err != nil {
		b.Fatal(err)
	}

	now := time.Now()

	for n := 0; n < b.N; n++ {
		r := bytes.NewReader(buffer)
		err = writeToDuckDB(r, appender.AppendRow)
		if err != nil {
			b.Fatal(err)
		}
		_ = appender.Flush()
	}

	elapsed := time.Now().Sub(now)
	log.Info().
		Int("n", b.N).
		Int64("elapsed_ns", elapsed.Nanoseconds()).
		Int64("ops", 1000000000/(elapsed.Nanoseconds()/4/rows)).
		Msg("Done in time")
}

// elapsed_ns=1104481628/4 rows=321039 ns_per_row=860 -> 1240694 ops
func BenchmarkImportFromBpf(b *testing.B) {
	ctx := context.Background()
	dsn := "bench.ddb"
	tableName := "append_bench"

	_ = os.Remove(dsn)
	_ = os.Remove(dsn + ".wal")

	connector, err := duckdb.NewConnector(dsn, nil)
	if err != nil {
		b.Fatal(err)
	}

	log.Info().Str("dsn", dsn).Msg("Connecting to db")
	conn, err := connector.Connect(ctx)
	if err != nil {
		b.Fatal(err)
	}

	db := sql.OpenDB(connector)
	log.Info().Msg("DB created")

	_, err = db.Exec(fmt.Sprintf(dropTableSql, tableName))
	if err != nil {
		b.Fatal(err)
	}

	_, err = db.Exec(fmt.Sprintf(createTableSql, tableName))
	if err != nil {
		b.Fatal(err)
	}

	appender, err := duckdb.NewAppenderFromConn(conn, "", tableName)
	if err != nil {
		b.Fatal(err)
	}
	defer func() { _ = appender.Close() }()

	now := time.Now()

	for n := 0; n < b.N; n++ {
		var r *os.File
		r, err = os.Open("testdata/vfs-raw.ndjson")
		if err != nil {
			b.Fatal(err)
		}

		err = writeToDuckDB(r, appender.AppendRow)
		if err != nil {
			b.Fatal(err)
		}

		_ = appender.Flush()
		_ = r.Close()
	}

	elapsed := time.Now().Sub(now)
	log.Info().
		Int("n", b.N).
		Int64("elapsed_ns", elapsed.Nanoseconds()).
		Int64("ops", 1000000000/(elapsed.Nanoseconds()/4/rows)).
		Msg("Done in time")
}
