package wiretrace

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"

	"github.com/go-faster/errors"
)

// ErrBlobMismatch is returned when a blob record does not match its reference.
var ErrBlobMismatch = errors.New("blob mismatch")

// MaxLineSize limits the size of one events line.
const MaxLineSize = 64 << 20

var segmentRe = regexp.MustCompile(`^events-(\d{4,})\.jsonl$`)

// EventsName returns the name of the events file of segment.
func EventsName(segment int) string {
	return fmt.Sprintf("events-%04d.jsonl", segment)
}

// BlobsName returns the name of the blobs file of segment.
func BlobsName(segment int) string {
	return fmt.Sprintf("blobs-%04d.bin", segment)
}

// Run is a trace run directory opened for reading.
type Run struct {
	// Dir is the run directory.
	Dir string
	// Segments are the indexes of present segments, sorted. Segments may
	// have gaps: the writer deletes the oldest segments except 0 when the
	// total size cap is reached.
	Segments []int
	// Header is the "hdr" record of the first present segment.
	Header *Header

	blobs map[int]*os.File
}

// Open opens a run directory.
func Open(dir string) (*Run, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, errors.Wrap(err, "read run dir")
	}
	r := &Run{Dir: dir, blobs: map[int]*os.File{}}
	for _, e := range entries {
		m := segmentRe.FindStringSubmatch(e.Name())
		if m == nil {
			continue
		}
		n, err := strconv.Atoi(m[1])
		if err != nil {
			continue
		}
		r.Segments = append(r.Segments, n)
	}
	if len(r.Segments) == 0 {
		return nil, errors.Errorf("no events-NNNN.jsonl in %s", dir)
	}
	sort.Ints(r.Segments)
	h, err := r.readHeader(r.Segments[0])
	if err != nil {
		return nil, err
	}
	r.Header = h
	return r, nil
}

func (r *Run) readHeader(segment int) (*Header, error) {
	f, err := os.Open(filepath.Join(r.Dir, EventsName(segment)))
	if err != nil {
		return nil, errors.Wrap(err, "open events")
	}
	defer func() { _ = f.Close() }()
	line, err := bufio.NewReaderSize(f, 4096).ReadBytes('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return nil, errors.Wrap(err, "read header")
	}
	ev, err := ParseEvent(bytes.TrimSpace(line))
	if err != nil {
		return nil, errors.Wrapf(err, "segment %d header", segment)
	}
	if ev.Kind != KindHeader {
		return nil, errors.Errorf("segment %d: first record is %q, not %q", segment, ev.Kind, KindHeader)
	}
	return ev.Header()
}

// Close releases opened blob files.
func (r *Run) Close() error {
	var err error
	for seg, f := range r.blobs {
		if cErr := f.Close(); cErr != nil && err == nil {
			err = cErr
		}
		delete(r.blobs, seg)
	}
	return err
}

// Events calls fn for every record of every segment in file order, including
// "hdr" records. Lines that can't be parsed are reported via onBad (if not
// nil) and skipped.
func (r *Run) Events(fn func(ev *Event) error, onBad func(segment, line int, err error)) error {
	for _, seg := range r.Segments {
		if err := r.segmentEvents(seg, fn, onBad); err != nil {
			return err
		}
	}
	return nil
}

func (r *Run) segmentEvents(seg int, fn func(ev *Event) error, onBad func(segment, line int, err error)) error {
	f, err := os.Open(filepath.Join(r.Dir, EventsName(seg)))
	if err != nil {
		return errors.Wrap(err, "open events")
	}
	defer func() { _ = f.Close() }()

	s := bufio.NewScanner(f)
	s.Buffer(make([]byte, 0, 64*1024), MaxLineSize)
	n := 0
	for s.Scan() {
		n++
		line := bytes.TrimSpace(s.Bytes())
		if len(line) == 0 {
			continue
		}
		ev, err := ParseEvent(append([]byte(nil), line...))
		if err != nil {
			if onBad != nil {
				onBad(seg, n, err)
			}
			continue
		}
		ev.Segment = seg
		ev.Line = n
		if err := fn(ev); err != nil {
			return err
		}
	}
	if err := s.Err(); err != nil {
		// A truncated last line of a live trace is fine, report the rest.
		return errors.Wrapf(err, "scan %s", EventsName(seg))
	}
	return nil
}

// All reads all records of the run.
func (r *Run) All() ([]*Event, error) {
	var out []*Event
	if err := r.Events(func(ev *Event) error {
		out = append(out, ev)
		return nil
	}, nil); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *Run) blobFile(seg int) (*os.File, error) {
	if f, ok := r.blobs[seg]; ok {
		return f, nil
	}
	f, err := os.Open(filepath.Join(r.Dir, BlobsName(seg)))
	if err != nil {
		return nil, errors.Wrap(err, "open blobs")
	}
	var hdr [BlobFileHeaderSize]byte
	if _, err := io.ReadFull(f, hdr[:]); err != nil {
		_ = f.Close()
		return nil, errors.Wrapf(err, "read %s header", BlobsName(seg))
	}
	if string(hdr[:4]) != BlobMagic {
		_ = f.Close()
		return nil, errors.Errorf("%s: bad magic %q", BlobsName(seg), hdr[:4])
	}
	if v := binary.LittleEndian.Uint16(hdr[4:]); v != Version {
		_ = f.Close()
		return nil, errors.Errorf("%s: unsupported version %d", BlobsName(seg), v)
	}
	r.blobs[seg] = f
	return f, nil
}

// Blob returns the body of the event, verifying the record header (seq, len,
// crc32c) and the sha256 of the reference.
//
// If the blob segment is missing but the body was inlined ("b64"), the
// inline body is returned. Returns nil, nil for events without a body.
func (r *Run) Blob(ev *Event) ([]byte, error) {
	ref := ev.Blob
	if ref == nil {
		if ev.Inline != nil {
			return ev.Inline, nil
		}
		return nil, nil
	}
	f, err := r.blobFile(ref.Seg)
	if err != nil {
		if ev.Inline != nil {
			return ev.Inline, nil
		}
		return nil, err
	}
	if ref.Off < BlobFileHeaderSize+BlobRecordHeaderSize || ref.Len < 0 {
		return nil, errors.Wrapf(ErrBlobMismatch, "event %d: bad offset %d", ev.Seq, ref.Off)
	}
	buf := make([]byte, BlobRecordHeaderSize+ref.Len)
	if _, err := f.ReadAt(buf, ref.Off-BlobRecordHeaderSize); err != nil {
		return nil, errors.Wrapf(err, "event %d: read blob", ev.Seq)
	}
	body := buf[BlobRecordHeaderSize:]
	if err := VerifyBlob(ev.Seq, ref, buf[:BlobRecordHeaderSize], body); err != nil {
		return nil, err
	}
	return body, nil
}

// VerifyBlob checks a blob record header and body against its reference.
func VerifyBlob(seq uint64, ref *BlobRef, header, body []byte) error {
	if len(header) != BlobRecordHeaderSize {
		return errors.Wrapf(ErrBlobMismatch, "event %d: short record header", seq)
	}
	if s := binary.LittleEndian.Uint64(header); s != seq {
		return errors.Wrapf(ErrBlobMismatch, "event %d: record seq %d", seq, s)
	}
	if l := binary.LittleEndian.Uint32(header[8:]); int(l) != ref.Len || len(body) != ref.Len {
		return errors.Wrapf(ErrBlobMismatch, "event %d: record len %d, ref len %d", seq, l, ref.Len)
	}
	crc := CRC32C(body)
	if c := binary.LittleEndian.Uint32(header[12:]); c != crc {
		return errors.Wrapf(ErrBlobMismatch, "event %d: record crc32c %d, body %d", seq, c, crc)
	}
	if ref.CRC32C != 0 && ref.CRC32C != crc {
		return errors.Wrapf(ErrBlobMismatch, "event %d: ref crc32c %d, body %d", seq, ref.CRC32C, crc)
	}
	if ref.SHA256 != "" {
		sum := sha256.Sum256(body)
		if got := hex.EncodeToString(sum[:]); got != ref.SHA256 {
			return errors.Wrapf(ErrBlobMismatch, "event %d: sha256 %s, ref %s", seq, got, ref.SHA256)
		}
	}
	return nil
}
