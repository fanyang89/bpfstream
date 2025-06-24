package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"

	"github.com/marcboeker/go-duckdb/v2"
)

func BenchmarkImportFromBpf(b *testing.B) {
	r, err := os.Open("testdata/vfs-raw.ndjson")
	if err != nil {
		b.Fatal(err)
	}
	defer func() { _ = r.Close() }()

	ctx := context.Background()
	dsn := "bench.ddb"
	tableName := "append_bench"

	connector, err := duckdb.NewConnector(dsn, nil)
	if err != nil {
		b.Fatal(err)
	}

	conn, err := connector.Connect(ctx)
	if err != nil {
		b.Fatal(err)
	}

	db := sql.OpenDB(connector)
	defer func() { _ = db.Close() }()

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

	for n := 0; n < b.N; n++ {
		err = writeToDuckDB(r, appender.AppendRow)
		if err != nil {
			b.Fatal(err)
		}
	}
}
