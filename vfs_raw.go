package main

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/kr/logfmt"
	"github.com/marcboeker/go-duckdb/v2"
	"github.com/minio/simdjson-go"
	"github.com/negrel/assert"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v3"
)

const createTableSql = `CREATE TABLE IF NOT EXISTS %s (
	Ts UBIGINT,
	Probe STRING,
	Tid UBIGINT,
	RC  BIGINT,
	Path STRING,
	Inode UBIGINT,
	"Offset" UBIGINT,
	Length UBIGINT)`

const dropTableSql = `DROP TABLE IF EXISTS %s`

type vfsEvent struct {
	Timestamp   uint64
	Probe       string
	Tid         uint64
	ReturnValue int64
	Path        string
	Inode       uint64
	Offset      uint64
	Length      uint64
}

var ErrUnknownField = errors.New("unknown field")

func (e *vfsEvent) HandleLogfmt(key []byte, val []byte) (err error) {
	k := string(key)
	v := string(val)
	if strings.HasSuffix(v, ",") {
		v = strings.TrimSuffix(v, ",")
	}
	switch k {
	case "ts":
		e.Timestamp, err = strconv.ParseUint(v, 10, 64)
	case "fn":
		e.Probe = v
	case "tid":
		e.Tid, err = strconv.ParseUint(v, 10, 64)
	case "rc":
		e.ReturnValue, err = strconv.ParseInt(v, 10, 64)
	case "path":
		e.Path = v[1 : len(v)-1]
	case "inode":
		e.Inode, err = strconv.ParseUint(v, 10, 64)
	case "offset":
		e.Offset, err = strconv.ParseUint(v, 10, 64)
	case "len":
		e.Length, err = strconv.ParseUint(v, 10, 64)
	default:
		err = ErrUnknownField
	}
	return
}

func (e *vfsEvent) Append(appender *duckdb.Appender) error {
	return appender.AppendRow(e.Timestamp, e.Probe, e.Tid, e.ReturnValue, e.Path, e.Inode, e.Offset, e.Length)
}

var vfsRawCmd = &cli.Command{
	Name: "raw",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "input",
			Aliases: []string{"i"},
			Value:   "-",
		},
		&cli.StringFlag{
			Name:     "dsn",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "table",
			Required: true,
		},
	},
	Action: func(ctx context.Context, command *cli.Command) error {
		dsn := command.String("dsn")
		tableName := command.String("table")

		connector, err := duckdb.NewConnector(dsn, nil)
		if err != nil {
			return err
		}

		conn, err := connector.Connect(ctx)
		if err != nil {
			return err
		}

		db := sql.OpenDB(connector)
		_, err = db.Exec(fmt.Sprintf(dropTableSql, tableName))
		if err != nil {
			return err
		}

		_, err = db.Exec(fmt.Sprintf(createTableSql, tableName))
		if err != nil {
			return err
		}

		appender, err := duckdb.NewAppenderFromConn(conn, "", tableName)
		if err != nil {
			return err
		}
		defer func() { _ = appender.Close() }()

		var r io.Reader
		input := command.String("input")
		if input == "-" {
			r = os.Stdin
		} else {
			f, err := os.Open(input)
			if err != nil {
				return errors.Wrap(err, "open input")
			}
			defer func() { _ = f.Close() }()
			r = f
		}

		var startTime time.Time

		reuse := make(chan *simdjson.ParsedJson, 10)
		res := make(chan simdjson.Stream, 10)
		simdjson.ParseNDStream(r, res, reuse)

		for got := range res {
			err := got.Error
			if err != nil {
				if err == io.EOF {
					break
				}
				return err
			}

			err = got.Value.ForEach(func(iter simdjson.Iter) error {
				var typeEl, dataEl *simdjson.Element
				typeEl, err = iter.FindElement(typeEl, "type")
				assert.NoError(err)
				typeStr, err := typeEl.Iter.String()
				assert.NoError(err)
				dataEl, err = iter.FindElement(dataEl, "data")
				assert.NoError(err)

				switch typeStr {
				case "attached_probes":
					var probesEl *simdjson.Element
					probesEl, err = dataEl.Iter.FindElement(probesEl, "probes")
					assert.NoError(err)
					probes, err := probesEl.Iter.Int()
					assert.NoError(err)
					if probes <= 0 {
						return errors.New("probes not attached")
					}

				case "time":
					assert.True(startTime.IsZero())
					timeStr, err := dataEl.Iter.String()
					assert.NoError(err)
					timeStr = strings.TrimSpace(timeStr)
					startTime, err = time.Parse(time.TimeOnly, timeStr)
					assert.NoError(err)
					log.Info().Str("start_time", startTime.Format(time.TimeOnly)).Msg("Record start from")

				case "lost_events":
					var eventCountEl *simdjson.Element
					eventCountEl, err = dataEl.Iter.FindElement(eventCountEl, "events")
					assert.NoError(err)
					var lostEvents int64
					lostEvents, err = eventCountEl.Iter.Int()
					assert.NoError(err)
					log.Info().Int64("lost_events", lostEvents).Msg("Lost events")

				case "printf":
					buf, err := dataEl.Iter.StringBytes()
					assert.NoError(err)
					var e vfsEvent
					err = logfmt.Unmarshal(buf, &e)
					assert.NoError(err)
					err = e.Append(appender)
					assert.NoError(err)
				}
				return nil
			})
		}
		return nil
	},
}
