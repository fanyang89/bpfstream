package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/marcboeker/go-duckdb/v2"
	"github.com/rs/zerolog/log"
)

// elapsed_ns=1104481628/4 rows=321039 ns_per_row=860 -> 1240694 ops
func BenchmarkImportFromBpf(b *testing.B) {
	ctx := context.Background()
	dsn := "bench.ddb"
	tableName := "append_bench"

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
	log.Info().Int64("elapsed_ns", elapsed.Nanoseconds()).Int("n", b.N).Msg("Done in time")
}
