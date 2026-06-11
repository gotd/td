package examples

import (
	"bufio"
	"context"
	"io"
	"math/rand"
	"os"
	"os/exec"
	"time"

	"github.com/go-faster/errors"
	"github.com/pion/rtp"
)

// Opus is streamed in 20 ms frames at a 48 kHz clock, so each RTP packet
// advances the timestamp by 960 samples.
const (
	opusFrame      = 20 * time.Millisecond
	opusSamples    = 48000 / 1000 * 20 // 960
	opusHeaderPkts = 2                 // OpusHead + OpusTags
)

// StreamMP3 transcodes an MP3 file to Opus with ffmpeg and feeds it to write as
// RTP, paced in real time. ffmpeg must be installed.
//
// write is typically a track's WriteRTP or a call's WriteAudio; it may rewrite
// the SSRC and payload type, so only the sequence number and timestamp are set
// here.
func StreamMP3(ctx context.Context, write func(*rtp.Packet) error, path string) error {
	if _, err := os.Stat(path); err != nil {
		return errors.Wrap(err, "audio file")
	}

	cmd := exec.CommandContext(ctx, "ffmpeg",
		"-hide_banner", "-loglevel", "error",
		"-i", path,
		"-vn",
		"-ac", "2", "-ar", "48000",
		"-c:a", "libopus", "-b:a", "64k", "-application", "voip",
		"-frame_duration", "20",
		"-f", "ogg", "pipe:1",
	)
	cmd.Stderr = os.Stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return errors.Wrap(err, "ffmpeg stdout")
	}
	if err := cmd.Start(); err != nil {
		return errors.Wrap(err, "start ffmpeg (is it installed?)")
	}
	defer func() { _ = cmd.Wait() }()

	ogg := &oggDemuxer{r: bufio.NewReader(stdout)}
	// Skip the OpusHead and OpusTags metadata packets.
	for range opusHeaderPkts {
		if _, err := ogg.next(); err != nil {
			return errors.Wrap(err, "read opus header")
		}
	}

	ticker := time.NewTicker(opusFrame)
	defer ticker.Stop()

	seq := uint16(rand.Uint32()) //nolint:gosec // Non-cryptographic RTP sequence seed.
	ts := rand.Uint32()          //nolint:gosec // Non-cryptographic RTP timestamp seed.
	marker := true
	for {
		frame, err := ogg.next()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return errors.Wrap(err, "read opus frame")
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}

		seq++
		ts += opusSamples
		if err := write(&rtp.Packet{
			Header: rtp.Header{
				Version:        2,
				Marker:         marker,
				PayloadType:    111,
				SequenceNumber: seq,
				Timestamp:      ts,
			},
			Payload: frame,
		}); err != nil {
			if errors.Is(err, io.ErrClosedPipe) {
				return nil // Track closed: call ended.
			}
			return errors.Wrap(err, "write rtp")
		}
		marker = false
	}
}

// oggDemuxer extracts whole Opus packets from an Ogg bitstream, honouring the
// segment lacing table (so packets that span multiple pages are reassembled).
type oggDemuxer struct {
	r     io.Reader
	queue [][]byte
	cur   []byte
}

func (d *oggDemuxer) next() ([]byte, error) {
	for len(d.queue) == 0 {
		if err := d.readPage(); err != nil {
			return nil, err
		}
	}
	pkt := d.queue[0]
	d.queue = d.queue[1:]
	return pkt, nil
}

func (d *oggDemuxer) readPage() error {
	var header [27]byte
	if _, err := io.ReadFull(d.r, header[:]); err != nil {
		return err
	}
	if string(header[0:4]) != "OggS" {
		return errors.New("invalid ogg capture pattern")
	}

	segments := int(header[26])
	table := make([]byte, segments)
	if _, err := io.ReadFull(d.r, table); err != nil {
		return err
	}
	total := 0
	for _, n := range table {
		total += int(n)
	}
	data := make([]byte, total)
	if _, err := io.ReadFull(d.r, data); err != nil {
		return err
	}

	off := 0
	for _, n := range table {
		d.cur = append(d.cur, data[off:off+int(n)]...)
		off += int(n)
		// A lacing value below 255 terminates the current packet.
		if n < 255 {
			d.queue = append(d.queue, d.cur)
			d.cur = nil
		}
	}
	return nil
}
