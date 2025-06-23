package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/minio/simdjson-go"
	"github.com/negrel/assert"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v3"
)

var vfsCmd = &cli.Command{
	Name: "vfs",
	Commands: []*cli.Command{
		vfsCountCmd,
		vfsRawCmd,
	},
}

var vfsCountCmd = &cli.Command{
	Name: "count",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "input",
			Aliases: []string{"i"},
			Value:   "-",
		},
	},
	Action: func(ctx context.Context, command *cli.Command) error {
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

				case "map":
					var event Event
					event.Fill(dataEl)
					fmt.Printf("%+v\n", event)
				}

				return nil
			})
			if err != nil {
				return err
			}

			reuse <- got.Value
		}

		return nil
	},
}

type Event struct {
	Create   int64
	Open     int64
	Read     int64
	ReadLink int64
	ReadV    int64
	Write    int64
	WriteV   int64
	FSync    int64
}

func (e *Event) Fill(el *simdjson.Element) {
	var err error
	var rootEl *simdjson.Element
	rootEl, err = el.Iter.FindElement(rootEl, "@")
	assert.NoError(err)
	var obj *simdjson.Object
	obj, err = rootEl.Iter.Object(obj)
	assert.NoError(err)
	elements, err := obj.Parse(nil)
	assert.NoError(err)
	for _, m := range elements.Elements {
		var value int64
		value, err = m.Iter.Int()
		assert.NoError(err)
		switch m.Name {
		case "vfs_create":
			e.Create = value
		case "vfs_open":
			e.Open = value
		case "vfs_read":
			e.Read = value
		case "vfs_readlink":
			e.ReadLink = value
		case "vfs_readv":
			e.ReadV = value
		case "vfs_write":
			e.Write = value
		case "vfs_writev":
			e.WriteV = value
		case "vfs_fsync":
			e.FSync = value
		}
	}
}
