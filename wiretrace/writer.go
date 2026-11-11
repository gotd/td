package wiretrace

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/go-faster/errors"
)

// Field is a key and a JSON-encodable value of an event.
//
// Use ID, SignedID and TypeID for 64-bit identifiers and constructor ids, so
// they are encoded the same way tdesktop writes them.
type Field struct {
	Key   string
	Value any
}

// F is a shorthand for Field{Key: key, Value: value}.
func F(key string, value any) Field {
	return Field{Key: key, Value: value}
}

// WriterOptions configures Writer.
type WriterOptions struct {
	// Header are the fields of the "hdr" record (after "seq", "ev", "v" and
	// "seg"), e.g. run, app, layer.
	Header []Field
	// SegmentEventsBytes rotates segments when the events file exceeds it.
	// Default is 64 MiB.
	SegmentEventsBytes int64
	// SegmentBlobsBytes rotates segments when the blobs file exceeds it.
	// Default is 256 MiB.
	SegmentBlobsBytes int64
	// TotalBytes caps the total size; the oldest segments except segment 0
	// are deleted. Default is 2 GiB.
	TotalBytes int64
	// InlineMax also inlines bodies up to this size as "b64". Zero disables.
	InlineMax int
	// Now returns the event time in seconds. Default is the monotonic time
	// since NewWriter.
	Now func() float64
}

func (o *WriterOptions) setDefaults() {
	if o.SegmentEventsBytes == 0 {
		o.SegmentEventsBytes = 64 << 20
	}
	if o.SegmentBlobsBytes == 0 {
		o.SegmentBlobsBytes = 256 << 20
	}
	if o.TotalBytes == 0 {
		o.TotalBytes = 2 << 30
	}
	if o.Now == nil {
		start := time.Now()
		o.Now = func() float64 { return time.Since(start).Seconds() }
	}
}

type closedSegment struct {
	index int
	bytes int64
}

// Writer writes a trace run directory in the same format as the tdesktop
// tracer. It is safe for concurrent use.
type Writer struct {
	dir string
	opt WriterOptions

	mux        sync.Mutex
	seq        uint64
	segment    int
	events     *os.File
	eventsBuf  *bufio.Writer
	blobs      *os.File
	blobsBuf   *bufio.Writer
	eventsSize int64
	blobsSize  int64
	closed     []closedSegment
	line       bytes.Buffer
}

// NewWriter creates dir (mode 0700) and opens segment 0.
func NewWriter(dir string, opt WriterOptions) (*Writer, error) {
	opt.setDefaults()
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, errors.Wrap(err, "create run dir")
	}
	w := &Writer{dir: dir, opt: opt, segment: -1}
	if err := w.openSegment(0); err != nil {
		return nil, err
	}
	return w, nil
}

func (w *Writer) openSegment(index int) error {
	events, err := os.OpenFile(filepath.Join(w.dir, EventsName(index)), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return errors.Wrap(err, "create events")
	}
	blobs, err := os.OpenFile(filepath.Join(w.dir, BlobsName(index)), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		_ = events.Close()
		return errors.Wrap(err, "create blobs")
	}
	w.segment = index
	w.events, w.blobs = events, blobs
	w.eventsBuf, w.blobsBuf = bufio.NewWriter(events), bufio.NewWriter(blobs)
	w.eventsSize, w.blobsSize = 0, 0

	var hdr [BlobFileHeaderSize]byte
	copy(hdr[:], BlobMagic)
	binary.LittleEndian.PutUint16(hdr[4:], Version)
	if _, err := w.blobsBuf.Write(hdr[:]); err != nil {
		return errors.Wrap(err, "write blobs header")
	}
	w.blobsSize = BlobFileHeaderSize

	w.line.Reset()
	w.line.WriteString(`{"seq":0,"ev":"hdr","v":`)
	w.line.WriteString(strconv.Itoa(Version))
	w.line.WriteString(`,"seg":`)
	w.line.WriteString(strconv.Itoa(index))
	if err := appendFields(&w.line, w.opt.Header); err != nil {
		return err
	}
	w.line.WriteString("}\n")
	return w.flushLine()
}

func (w *Writer) flushLine() error {
	n, err := w.eventsBuf.Write(w.line.Bytes())
	w.eventsSize += int64(n)
	if err != nil {
		return errors.Wrap(err, "write event")
	}
	return nil
}

func (w *Writer) closeSegment() error {
	if w.events == nil {
		return nil
	}
	var err error
	for _, f := range []func() error{
		w.eventsBuf.Flush, w.blobsBuf.Flush, w.events.Close, w.blobs.Close,
	} {
		if fErr := f(); fErr != nil && err == nil {
			err = fErr
		}
	}
	w.closed = append(w.closed, closedSegment{index: w.segment, bytes: w.eventsSize + w.blobsSize})
	w.events, w.blobs = nil, nil
	if err != nil {
		return errors.Wrap(err, "close segment")
	}
	return nil
}

func (w *Writer) enforceTotal() {
	total := w.eventsSize + w.blobsSize
	for _, s := range w.closed {
		total += s.bytes
	}
	// Segment 0 holds the handshake and the first sync, never delete it.
	for total > w.opt.TotalBytes {
		idx := -1
		for i, s := range w.closed {
			if s.index != 0 {
				idx = i
				break
			}
		}
		if idx < 0 {
			return
		}
		s := w.closed[idx]
		_ = os.Remove(filepath.Join(w.dir, EventsName(s.index)))
		_ = os.Remove(filepath.Join(w.dir, BlobsName(s.index)))
		total -= s.bytes
		w.closed = append(w.closed[:idx], w.closed[idx+1:]...)
	}
}

func appendFields(b *bytes.Buffer, fields []Field) error {
	for _, f := range fields {
		b.WriteString(`,`)
		key, err := json.Marshal(f.Key)
		if err != nil {
			return errors.Wrap(err, "encode key")
		}
		b.Write(key)
		b.WriteString(`:`)
		v, err := json.Marshal(f.Value)
		if err != nil {
			return errors.Wrapf(err, "encode %q", f.Key)
		}
		b.Write(v)
	}
	return nil
}

// Write appends an event and returns its seq. If blob is not nil, it is
// written to the blobs segment and referenced as "blob" (and inlined as
// "b64" if it is small enough).
func (w *Writer) Write(thread, kind string, fields []Field, blob []byte) (uint64, error) {
	w.mux.Lock()
	defer w.mux.Unlock()
	w.seq++
	return w.seq, w.write(w.seq, thread, kind, fields, blob)
}

// WriteDropped appends a "dropped" record, as the tdesktop writer does when
// its queue overflows.
func (w *Writer) WriteDropped(count uint64) (uint64, error) {
	return w.Write("trace", KindDropped, []Field{F("count", count)}, nil)
}

func (w *Writer) write(seq uint64, thread, kind string, fields []Field, blob []byte) error {
	if w.events == nil {
		return errors.New("writer is closed")
	}
	if w.eventsSize > w.opt.SegmentEventsBytes || w.blobsSize > w.opt.SegmentBlobsBytes {
		next := w.segment + 1
		if err := w.closeSegment(); err != nil {
			return err
		}
		if err := w.openSegment(next); err != nil {
			return err
		}
		w.enforceTotal()
	}

	w.line.Reset()
	w.line.WriteString(`{"seq":`)
	w.line.WriteString(strconv.FormatUint(seq, 10))
	w.line.WriteString(`,"t":`)
	w.line.WriteString(strconv.FormatFloat(w.opt.Now(), 'f', 6, 64))
	th, _ := json.Marshal(thread)
	w.line.WriteString(`,"th":`)
	w.line.Write(th)
	ev, _ := json.Marshal(kind)
	w.line.WriteString(`,"ev":`)
	w.line.Write(ev)
	if err := appendFields(&w.line, fields); err != nil {
		return err
	}
	if blob != nil {
		var hdr [BlobRecordHeaderSize]byte
		crc := CRC32C(blob)
		binary.LittleEndian.PutUint64(hdr[:], seq)
		binary.LittleEndian.PutUint32(hdr[8:], uint32(len(blob))) // #nosec G115
		binary.LittleEndian.PutUint32(hdr[12:], crc)
		off := w.blobsSize + BlobRecordHeaderSize
		if _, err := w.blobsBuf.Write(hdr[:]); err != nil {
			return errors.Wrap(err, "write blob header")
		}
		if _, err := w.blobsBuf.Write(blob); err != nil {
			return errors.Wrap(err, "write blob")
		}
		w.blobsSize += BlobRecordHeaderSize + int64(len(blob))
		sum := sha256.Sum256(blob)
		ref, err := json.Marshal(BlobRef{
			Seg:    w.segment,
			Off:    off,
			Len:    len(blob),
			CRC32C: crc,
			SHA256: hex.EncodeToString(sum[:]),
		})
		if err != nil {
			return errors.Wrap(err, "encode blob ref")
		}
		w.line.WriteString(`,"blob":`)
		w.line.Write(ref)
		if w.opt.InlineMax > 0 && len(blob) <= w.opt.InlineMax {
			w.line.WriteString(`,"b64":"`)
			w.line.WriteString(base64.StdEncoding.EncodeToString(blob))
			w.line.WriteString(`"`)
		}
	}
	w.line.WriteString("}\n")
	return w.flushLine()
}

// Flush flushes buffered data to the files.
func (w *Writer) Flush() error {
	w.mux.Lock()
	defer w.mux.Unlock()
	if w.events == nil {
		return nil
	}
	if err := w.eventsBuf.Flush(); err != nil {
		return errors.Wrap(err, "flush events")
	}
	if err := w.blobsBuf.Flush(); err != nil {
		return errors.Wrap(err, "flush blobs")
	}
	return nil
}

// Close flushes and closes the current segment.
func (w *Writer) Close() error {
	w.mux.Lock()
	defer w.mux.Unlock()
	return w.closeSegment()
}

// StructFields converts a struct with json tags (e.g. Msg) into ordered
// fields, keeping the field order of the encoded JSON object.
func StructFields(v any) ([]Field, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, errors.Wrap(err, "encode")
	}
	d := json.NewDecoder(bytes.NewReader(data))
	tok, err := d.Token()
	if err != nil {
		return nil, errors.Wrap(err, "decode")
	}
	if delim, ok := tok.(json.Delim); !ok || delim != '{' {
		return nil, errors.Errorf("expected object, got %v", tok)
	}
	var out []Field
	for d.More() {
		tok, err := d.Token()
		if err != nil {
			return nil, errors.Wrap(err, "decode key")
		}
		key, ok := tok.(string)
		if !ok {
			return nil, errors.Errorf("unexpected key %v", tok)
		}
		var raw json.RawMessage
		if err := d.Decode(&raw); err != nil {
			return nil, errors.Wrapf(err, "decode %q", key)
		}
		out = append(out, F(key, raw))
	}
	return out, nil
}

// WriteMsg appends a "msg" event with the plaintext body.
func (w *Writer) WriteMsg(thread string, m *Msg, body []byte) (uint64, error) {
	fields, err := StructFields(m)
	if err != nil {
		return 0, err
	}
	if body == nil {
		body = []byte{}
	}
	return w.Write(thread, KindMsg, fields, body)
}
