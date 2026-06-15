// Binary bot-bigbuckbunny is a bot that replies to every message with the
// Big Buck Bunny video, streamed directly from the Blender download server.
//
// It demonstrates how to build a multi-connection invoker with client.Pool(N)
// and feed it to the uploader, spreading upload RPCs across N sub-connections
// to the current DC.
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ernado/ff/ffprobe"
	"github.com/ernado/ff/ffrun"
	"github.com/go-faster/errors"
	"go.uber.org/zap"

	"github.com/gotd/log/logzap"

	"github.com/gotd/td/examples"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/telegram/message/styling"
	"github.com/gotd/td/telegram/uploader"
	"github.com/gotd/td/tg"
)

// progressLogger is an uploader.Progress that logs upload progress at INFO
// level, throttled to at most once every 5 seconds (the final 100% state is
// always logged). Chunk may be called concurrently from multiple
// sub-connections, so access is serialized with a mutex.
type progressLogger struct {
	log *zap.Logger

	mu           sync.Mutex
	interval     time.Duration
	start        time.Time
	lastLog      time.Time
	lastUploaded int64
}

// Chunk implements uploader.Progress.
func (p *progressLogger) Chunk(_ context.Context, state uploader.ProgressState) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	now := time.Now()
	if p.start.IsZero() {
		p.start = now
	}

	done := state.Total > 0 && state.Uploaded >= state.Total
	if !done && now.Sub(p.lastLog) < p.interval {
		return nil
	}

	// Instantaneous speed since the last log, plus average since start.
	var speed float64
	if d := now.Sub(p.lastLog).Seconds(); d > 0 && !p.lastLog.IsZero() {
		speed = mib(state.Uploaded-p.lastUploaded) / d
	}
	var avg float64
	if d := now.Sub(p.start).Seconds(); d > 0 {
		avg = mib(state.Uploaded) / d
	}
	p.lastLog = now
	p.lastUploaded = state.Uploaded

	fields := []zap.Field{
		zap.String("name", state.Name),
		zap.Float64("uploaded_mib", mib(state.Uploaded)),
		zap.Float64("speed_mib_s", speed),
		zap.Float64("avg_mib_s", avg),
	}
	if state.Total > 0 {
		fields = append(fields,
			zap.Float64("total_mib", mib(state.Total)),
			zap.Float64("percent", float64(state.Uploaded)/float64(state.Total)*100),
		)
	}
	p.log.Info("Upload progress", fields...)
	return nil
}

func mib(b int64) float64 { return float64(b) / (1024 * 1024) }

// isTTY reports whether f is a terminal (character device), used to decide
// whether to emit ANSI color codes.
func isTTY(f *os.File) bool {
	fi, err := f.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

// colorizer wraps text in ANSI color codes when enabled (i.e. on a TTY).
type colorizer struct{ enabled bool }

func (c colorizer) wrap(code, s string) string {
	if !c.enabled {
		return s
	}
	return "\x1b[" + code + "m" + s + "\x1b[0m"
}

func (c colorizer) bold(s string) string  { return c.wrap("1", s) }
func (c colorizer) green(s string) string { return c.wrap("32", s) }
func (c colorizer) cyan(s string) string  { return c.wrap("36", s) }

// summary holds upload statistics for the final pretty-printed report.
type summary struct {
	name          string
	size          int64
	duration      time.Duration
	pool          bool
	conns         int64
	width, height int
	mime          string
	thumb         bool
}

// print writes a human-friendly, optionally colored summary to stdout.
func (s summary) print(c colorizer) {
	var avg float64
	if d := s.duration.Seconds(); d > 0 {
		avg = mib(s.size) / d
	}
	transport := "main connection"
	if s.pool {
		transport = fmt.Sprintf("pool (%d connections)", s.conns)
	}
	thumb := "no"
	if s.thumb {
		thumb = "yes"
	}

	var b strings.Builder
	fmt.Fprintln(&b, c.green(c.bold("✓ Upload complete")))
	row := func(label, value string) {
		fmt.Fprintf(&b, "  %s %s\n", c.cyan(fmt.Sprintf("%-11s", label+":")), value)
	}
	row("File", s.name)
	row("Size", fmt.Sprintf("%.1f MiB", mib(s.size)))
	if s.mime != "" {
		row("MIME", s.mime)
	}
	if s.width > 0 && s.height > 0 {
		row("Resolution", fmt.Sprintf("%dx%d", s.width, s.height))
	}
	row("Upload time", s.duration.Round(time.Millisecond).String())
	row("Avg speed", fmt.Sprintf("%.2f MiB/s", avg))
	row("Transport", transport)
	row("Thumbnail", thumb)
	fmt.Fprint(os.Stdout, b.String())
}

// defaultVideoURL is a 480p H.264 sample video served by the Blender Foundation,
// used as the default when no -url flag is provided.
const defaultVideoURL = "https://download.blender.org/peach/bigbuckbunny_movies/big_buck_bunny_480p_h264.mov"

// videoName derives a temp filename from a video URL, falling back to "video".
func videoName(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "video"
	}
	name := path.Base(u.Path)
	if name == "" || name == "." || name == "/" {
		return "video"
	}
	return name
}

// ensureFile downloads url into a file named name inside the OS temp dir,
// returning its path. If the file already exists it is reused as-is and no
// download is performed. The download goes to a temporary file that is renamed
// into place only on success, so a partial download is never mistaken for a
// complete one.
func ensureFile(ctx context.Context, log *zap.Logger, url, name string) (string, error) {
	path := filepath.Join(os.TempDir(), name)
	if _, err := os.Stat(path); err == nil {
		log.Info("Using cached file", zap.String("path", path))
		return path, nil
	} else if !os.IsNotExist(err) {
		return "", errors.Wrap(err, "stat")
	}

	log.Info("Downloading file", zap.String("url", url), zap.String("path", path))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", errors.Wrap(err, "new request")
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", errors.Wrap(err, "do request")
	}
	defer func() { _ = res.Body.Close() }()
	if res.StatusCode != http.StatusOK {
		return "", errors.Errorf("unexpected status %s", res.Status)
	}

	tmp, err := os.CreateTemp(os.TempDir(), name+".*.part")
	if err != nil {
		return "", errors.Wrap(err, "create temp")
	}
	tmpPath := tmp.Name()
	if _, err := io.Copy(tmp, res.Body); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
		return "", errors.Wrap(err, "download")
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return "", errors.Wrap(err, "close temp")
	}
	if err := os.Rename(tmpPath, path); err != nil {
		_ = os.Remove(tmpPath)
		return "", errors.Wrap(err, "rename")
	}
	return path, nil
}

// videoMeta holds video metadata probed with ffprobe and an optional thumbnail.
// All fields are best-effort: when ffmpeg/ffprobe are unavailable they stay
// zero/empty and the caller sends the video without dimensions or thumbnail.
type videoMeta struct {
	width, height int
	duration      time.Duration
	mime          string // derived from container format; empty if unknown
	thumbPath     string // empty if no thumbnail was generated
}

// mimeByFormat maps ffprobe format_name tokens to MIME types. The mov/mp4/m4a
// family (ISOBMFF) is reported as a single comma-separated format_name; it maps
// to video/mp4, which Telegram handles as a proper streamable video.
var mimeByFormat = map[string]string{
	"mp4":      "video/mp4",
	"mov":      "video/mp4",
	"m4v":      "video/mp4",
	"webm":     "video/webm",
	"matroska": "video/x-matroska",
	"avi":      "video/x-msvideo",
	"flv":      "video/x-flv",
	"mpegts":   "video/mp2t",
	"mpeg":     "video/mpeg",
	"asf":      "video/x-ms-asf",
	"ogg":      "video/ogg",
	"3gp":      "video/3gpp",
}

// videoMIME picks a MIME type from a comma-separated ffprobe format_name,
// e.g. "mov,mp4,m4a,3gp,3g2,mj2". Returns "" when no token is recognized.
func videoMIME(formatName string) string {
	for _, f := range strings.Split(formatName, ",") {
		if mime, ok := mimeByFormat[strings.TrimSpace(f)]; ok {
			return mime
		}
	}
	return ""
}

// probeVideo probes videoPath for dimensions/duration and extracts a JPEG
// thumbnail, both via ffmpeg (github.com/ernado/ff). It is optional: if ffmpeg
// or ffprobe is not installed (or a step fails) it logs a warning and returns
// whatever it managed to gather, so the bot keeps working without them.
func probeVideo(ctx context.Context, log *zap.Logger, videoPath, thumbName string) videoMeta {
	var meta videoMeta

	if !ffmpegAvailable() {
		log.Warn("ffmpeg/ffprobe not found in PATH; sending video without dimensions or thumbnail")
		return meta
	}

	ff := ffrun.New(ffrun.Options{})

	// Probe for dimensions and duration.
	probe, err := ff.Probe(ctx, videoPath)
	if err != nil {
		log.Warn("Failed to probe video; sending without dimensions", zap.Error(err))
		return meta
	}
	meta.mime = videoMIME(probe.Format.FormatName)
	if s, err := ffprobe.ParseSummary(probe); err != nil {
		log.Warn("Failed to parse probe summary", zap.Error(err))
	} else {
		meta.width, meta.height, meta.duration = s.Width, s.Height, s.Duration
		log.Info("Probed video",
			zap.Int("width", s.Width),
			zap.Int("height", s.Height),
			zap.Duration("duration", s.Duration),
			zap.String("format", probe.Format.FormatName),
			zap.String("mime", meta.mime),
		)
	}

	// Extract a thumbnail: a single frame at 5s, scaled to a max width of 320px
	// (the Telegram thumbnail size limit; height is computed to keep ratio).
	thumbPath := filepath.Join(os.TempDir(), thumbName)
	if _, err := os.Stat(thumbPath); err == nil {
		log.Info("Using cached thumbnail", zap.String("path", thumbPath))
		meta.thumbPath = thumbPath
		return meta
	}

	log.Info("Generating thumbnail", zap.String("path", thumbPath))
	if err := ff.Run(ctx, ffrun.RunOptions{
		Input:     videoPath,
		Output:    thumbPath,
		Probe:     probe,
		InputArgs: []string{"-ss", "5"},
		Args:      []string{"-frames:v", "1", "-vf", "scale=320:-1"},
	}); err != nil {
		log.Warn("Failed to generate thumbnail; sending video without it", zap.Error(err))
		return meta
	}
	meta.thumbPath = thumbPath
	return meta
}

// ffmpegAvailable reports whether both ffmpeg and ffprobe are in PATH.
func ffmpegAvailable() bool {
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		return false
	}
	if _, err := exec.LookPath("ffprobe"); err != nil {
		return false
	}
	return true
}

// secs formats a duration as fractional seconds for ffmpeg arguments.
func secs(d time.Duration) string {
	return strconv.FormatFloat(d.Seconds(), 'f', -1, 64)
}

// processVideo optionally cuts (ffmpeg -t) and/or compresses srcPath to fit
// within maxSizeMB megabytes, returning the path to use for upload. If neither
// option is requested (or ffmpeg is unavailable), srcPath is returned unchanged.
//
// The target bitrate for compression is approximated from the requested size
// and the (possibly cut) duration; the source's current bitrate is used to skip
// compression when the video already fits the budget.
func processVideo(ctx context.Context, log *zap.Logger, srcPath, name string, cut time.Duration, maxSizeMB int64) (string, error) {
	if cut <= 0 && maxSizeMB <= 0 {
		return srcPath, nil
	}
	if !ffmpegAvailable() {
		log.Warn("ffmpeg/ffprobe not found in PATH; skipping cut/compress")
		return srcPath, nil
	}

	ff := ffrun.New(ffrun.Options{})
	probe, err := ff.Probe(ctx, srcPath)
	if err != nil {
		return "", errors.Wrap(err, "probe")
	}
	summary, err := ffprobe.ParseSummary(probe)
	if err != nil {
		return "", errors.Wrap(err, "summary")
	}

	// Effective duration after an optional cut, used for the bitrate budget.
	duration := summary.Duration
	if cut > 0 && cut < duration {
		duration = cut
	}

	var args []string
	if cut > 0 {
		args = append(args, "-t", secs(cut))
	}

	compress := false
	if maxSizeMB > 0 && duration > 0 {
		// Total bitrate (bits/s) that fits maxSizeMB over the duration.
		targetTotal := int64(float64(maxSizeMB*1024*1024*8) / duration.Seconds())
		current, _ := strconv.ParseInt(probe.Format.BitRate, 10, 64)
		switch {
		case current > 0 && current <= targetTotal:
			log.Info("Video already within size budget; skipping compression",
				zap.Int64("current_bitrate", current),
				zap.Int64("target_bitrate", targetTotal),
			)
		default:
			compress = true
			// Reserve a fixed audio budget, give the rest to video.
			const audioBitrate = 128 * 1024
			videoBitrate := targetTotal - audioBitrate
			if videoBitrate < 64*1024 {
				videoBitrate = 64 * 1024 // floor for very small budgets
			}
			args = append(args,
				"-b:v", strconv.FormatInt(videoBitrate, 10),
				"-maxrate", strconv.FormatInt(videoBitrate, 10),
				"-bufsize", strconv.FormatInt(videoBitrate*2, 10),
				"-b:a", strconv.FormatInt(audioBitrate, 10),
			)
			log.Info("Compressing video",
				zap.Int64("max_size_mb", maxSizeMB),
				zap.Int64("video_bitrate", videoBitrate),
				zap.Int64("current_bitrate", current),
			)
		}
	}

	// Fast path: cut only (no re-encode) can copy streams.
	if !compress && cut > 0 {
		args = append(args, "-c", "copy")
	}
	if len(args) == 0 {
		return srcPath, nil
	}

	// Always produce an MP4 container for broad Telegram compatibility.
	out := filepath.Join(os.TempDir(), "processed_"+strings.TrimSuffix(name, filepath.Ext(name))+".mp4")
	log.Info("Processing video with ffmpeg", zap.String("output", out), zap.Strings("args", args))
	if err := ff.Run(ctx, ffrun.RunOptions{
		Input:  srcPath,
		Output: out,
		Probe:  probe,
		Args:   args,
	}); err != nil {
		return "", errors.Wrap(err, "run ffmpeg")
	}
	return out, nil
}

func main() {
	// Environment variables:
	//	BOT_TOKEN:     token from BotFather
	// 	APP_ID:        app_id of Telegram app.
	// 	APP_HASH:      app_hash of Telegram app.
	// 	SESSION_FILE:  path to session file
	// 	SESSION_DIR:   path to session directory, if SESSION_FILE is not set
	videoURL := flag.String("url", defaultVideoURL, "video URL to download and reply with")
	conns := flag.Int64("conns", 4, "number of sub-connections in the upload pool")
	usePool := flag.Bool("pool", true, "use a dedicated multi-connection pool for uploads; if false, upload over the main connection")
	cut := flag.Duration("cut", 0, "if non-zero, cut the video to this duration (ffmpeg -t), e.g. 30s")
	maxSize := flag.Int64("max-size", 0, "if non-zero, compress the video to approximately fit within this many megabytes")
	flag.Parse()

	examples.Run(func(ctx context.Context, log *zap.Logger) error {
		// Dispatcher handles incoming updates.
		dispatcher := tg.NewUpdateDispatcher()
		opts := telegram.Options{
			Logger:        logzap.New(log),
			UpdateHandler: dispatcher,
		}
		// Note: the upload pool must be created inside the run callback (below),
		// not in the setup callback, because the pool relies on the client's run
		// context which is only available once client.Run has started.
		return telegram.BotFromEnvironment(ctx, opts, nil, func(ctx context.Context, client *telegram.Client) error {
			// Pick the invoker for uploads: either a dedicated multi-connection
			// pool to the current DC, or the main client connection.
			uploadInvoker := tg.NewClient(client)
			if *usePool {
				// client.Pool manages up to N sub-connections internally and
				// balances calls across them.
				pool, err := client.Pool(*conns)
				if err != nil {
					return errors.Wrap(err, "create pool")
				}
				defer func() { _ = pool.Close() }()
				uploadInvoker = tg.NewClient(pool)
				log.Info("Using upload pool", zap.Int64("conns", *conns))
			} else {
				log.Info("Using main connection for uploads")
			}

			// Feed the invoker to the uploader. WithThreads sets the per-upload
			// goroutine limit so part uploads fan out across the pool's
			// sub-connections; WithProgress logs progress as INFO.
			u := uploader.NewUploader(uploadInvoker).
				WithThreads(int(*conns)).
				WithProgress(&progressLogger{log: log, interval: 5 * time.Second})

			// Sender for replies uses the primary client connection. Attach the
			// uploader so UploadedDocument media is resolved through it.
			sender := message.NewSender(tg.NewClient(client)).WithUploader(u)

			// Download the video to the OS temp dir once (reused on restart if
			// it is already there), then upload from the local file.
			name := videoName(*videoURL)
			path, err := ensureFile(ctx, log, *videoURL, name)
			if err != nil {
				return errors.Wrap(err, "ensure file")
			}

			// Optionally cut and/or compress the video with ffmpeg before upload.
			path, err = processVideo(ctx, log, path, name, *cut, *maxSize)
			if err != nil {
				return errors.Wrap(err, "process video")
			}
			name = filepath.Base(path)

			// Optionally probe dimensions/duration and generate a thumbnail with
			// ffmpeg. Fields stay zero/empty when ffmpeg is unavailable.
			meta := probeVideo(ctx, log, path, name+".thumb.jpg")

			// Colored summary output only when stdout is a terminal.
			color := colorizer{enabled: isTTY(os.Stdout)}

			dispatcher.OnNewMessage(func(ctx context.Context, e tg.Entities, msg *tg.UpdateNewMessage) error {
				m, ok := msg.Message.(*tg.Message)
				if !ok || m.Out {
					// Outgoing message, not interesting.
					return nil
				}

				log.Info("Uploading video", zap.String("name", name))
				uploadStart := time.Now()
				f, err := u.FromPath(ctx, path)
				if err != nil {
					return errors.Wrap(err, "upload from path")
				}
				uploadDuration := time.Since(uploadStart)

				// MIME derived from the file contents by ffprobe; fall back to
				// video/mp4 when ffmpeg is unavailable or the format is unknown.
				mime := meta.mime
				if mime == "" {
					mime = "video/mp4"
				}
				docBuilder := message.UploadedDocument(f,
					styling.Bold(name),
				).
					MIME(mime).
					Filename(name)

				// Attach the thumbnail if one was generated. Thumb must be set
				// before Video(), which switches to the video builder.
				if meta.thumbPath != "" {
					thumb, err := u.FromPath(ctx, meta.thumbPath)
					if err != nil {
						return errors.Wrap(err, "upload thumbnail")
					}
					docBuilder = docBuilder.Thumb(thumb)
				}

				// Set probed dimensions and duration on the video, when known.
				video := docBuilder.Video()
				if meta.width > 0 && meta.height > 0 {
					video = video.Resolution(meta.width, meta.height)
				}
				if meta.duration > 0 {
					video = video.Duration(meta.duration)
				}

				log.Info("Sending video reply")
				if _, err := sender.Reply(e, msg).Media(ctx, video); err != nil {
					return errors.Wrap(err, "send")
				}

				// Pretty-printed summary to stdout.
				var size int64
				if fi, err := os.Stat(path); err == nil {
					size = fi.Size()
				}
				summary{
					name:     filepath.Base(path),
					size:     size,
					duration: uploadDuration,
					pool:     *usePool,
					conns:    *conns,
					width:    meta.width,
					height:   meta.height,
					mime:     mime,
					thumb:    meta.thumbPath != "",
				}.print(color)
				return nil
			})

			// Block until the context is canceled, handling updates meanwhile.
			<-ctx.Done()
			return ctx.Err()
		})
	})
}
