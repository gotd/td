package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/go-faster/errors"

	"github.com/gotd/td/wiretrace"
	"github.com/gotd/td/wiretrace/decode"
)

var errUsage = errors.New("usage")

func parseFlags(fs *flag.FlagSet, args []string, nargs int, usageLine string, stderr io.Writer) error {
	fs.SetOutput(stderr)
	fs.Usage = func() {
		_, _ = fmt.Fprintf(stderr, "Usage: wiretrace %s\n", usageLine)
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return errUsage
	}
	if fs.NArg() != nargs {
		fs.Usage()
		return errUsage
	}
	return nil
}

// logicalMessages runs both decode passes over a run and calls fn for every
// logical message.
func logicalMessages(ctx context.Context, run *wiretrace.Run, d *decode.Decoder, bad *badLines,
	fn func(ev *wiretrace.Event, m *decode.Message) error,
) error {
	if err := run.Events(func(ev *wiretrace.Event) error {
		d.Index(ev)
		return nil
	}, bad.add); err != nil {
		return err
	}
	return run.Events(func(ev *wiretrace.Event) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		if ev.Kind != wiretrace.KindMsg {
			return fn(ev, nil)
		}
		body, err := run.Blob(ev)
		if err != nil {
			m, mErr := ev.Msg()
			if mErr != nil {
				return mErr
			}
			msg := decode.Message{
				Seq:         ev.Seq,
				T:           ev.T,
				Dir:         m.Dir,
				SDC:         m.SDC,
				Session:     strconv.FormatInt(int64(m.Session), 10),
				MsgID:       strconv.FormatUint(uint64(m.MsgID), 10),
				MsgIDValue:  uint64(m.MsgID),
				SeqNo:       m.SeqNo,
				Type:        d.Schema().Name(uint32(m.Type)),
				TypeID:      uint32(m.Type),
				Partial:     true,
				DecodeError: "blob: " + err.Error(),
			}
			return fn(ev, &msg)
		}
		msgs, err := d.Messages(ev, body)
		if err != nil {
			return err
		}
		for i := range msgs {
			if err := fn(ev, &msgs[i]); err != nil {
				return err
			}
		}
		return nil
	}, nil)
}

func runDecode(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("decode", flag.ContinueOnError)
	schemaDir := fs.String("schema", "", "directory with api.tl and mtproto.tl, or comma-separated .tl files of the traced build")
	maxBytes := fs.Int("max-bytes", 1024, "summarize TL bytes longer than this (0: no limit)")
	dir := fs.String("dir", "", "only messages of this direction (in or out)")
	method := fs.String("method", "", "only messages whose type or method contains this")
	noBody := fs.Bool("no-body", false, "omit decoded bodies")
	if err := parseFlags(fs, args, 1, "decode [flags] RUN", stderr); err != nil {
		return err
	}
	run, err := openRun(fs.Arg(0))
	if err != nil {
		return err
	}
	defer func() { _ = run.Close() }()
	schema, err := loadSchema(stderr, *schemaDir, run, true)
	if err != nil {
		return err
	}

	w := bufio.NewWriter(stdout)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	var bad badLines
	dropped := uint64(0)
	d := decode.New(schema, decode.Options{MaxBytes: *maxBytes})
	if err := logicalMessages(ctx, run, d, &bad, func(ev *wiretrace.Event, m *decode.Message) error {
		if m == nil {
			if ev.Kind == wiretrace.KindDropped {
				dropped += ev.ID("count")
			}
			return nil
		}
		if *dir != "" && m.Dir != *dir {
			return nil
		}
		if *method != "" && !strings.Contains(m.Type, *method) && !strings.Contains(m.Method, *method) {
			return nil
		}
		if *noBody {
			m.Body = nil
		}
		return enc.Encode(m)
	}); err != nil {
		return err
	}
	bad.report(stderr)
	if dropped > 0 {
		_, _ = fmt.Fprintf(stderr, "wiretrace: the writer dropped %d events (queue overflow)\n", dropped)
	}
	return w.Flush()
}
