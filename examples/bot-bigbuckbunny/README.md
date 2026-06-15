# bot-bigbuckbunny

A bot that replies to **every** incoming message with a video. By default it
sends the [Big Buck Bunny](https://peach.blender.org/) sample (480p H.264, served
by the Blender Foundation); pass `-url` to use any other video.

It demonstrates uploading a video through a dedicated **multi-connection upload
pool**, with progress logging.

## What it shows

- `client.Pool(N)` — a multi-connection invoker to the current DC, fed to the
  uploader so part uploads are balanced across `N` sub-connections.
- `uploader.WithThreads(N)` — per-upload goroutine limit, matched to the pool
  size so the part uploads actually fan out across the connections.
- `uploader.WithProgress` — a custom `uploader.Progress` that logs progress at
  INFO level, throttled to once every 5 seconds, including instantaneous and
  average upload speed.
- Local file caching: the video is downloaded to the OS temp dir once (atomic
  rename on success) and reused on subsequent runs.
- Optional `ffmpeg` integration via [`github.com/ernado/ff`](https://github.com/ernado/ff):
  the video is probed for width/height/duration (set on the video with
  `Resolution`/`Duration`), its MIME type is derived from the container format
  reported by `ffprobe`, and a JPEG thumbnail is extracted and attached with
  `UploadedDocumentBuilder.Thumb`. This is best-effort — if `ffmpeg`/`ffprobe`
  are not in `PATH` (or a step fails) a warning is logged and the video is sent
  with a default `video/mp4` MIME and without dimensions or thumbnail.
- Optional pre-upload processing with ffmpeg: `-cut` trims the video to a
  duration (ffmpeg `-t`), and `-max-size` re-encodes it to approximately fit a
  size budget (target bitrate is derived from the requested size and the
  possibly-cut duration; compression is skipped when the source already fits).
- `message.UploadedDocument(...).Video()` — sending the uploaded file as a video
  reply.
- A pretty-printed summary to stdout after each upload (duration, average speed,
  size, transport, thumbnail), colored when stdout is a TTY.

## Run

Set the standard environment variables (see the [examples README](../README.md)):
`BOT_TOKEN`, `APP_ID`, `APP_HASH`, and a `SESSION_FILE`/`SESSION_DIR`.

[`ffmpeg`](https://ffmpeg.org/) (and `ffprobe`) are optional dependencies, used
to probe the video dimensions and generate the thumbnail. Without them the bot
still works and logs a warning.

```console
go run ./bot-bigbuckbunny
```

Then send any message to the bot.

## Flags

| Flag       | Default            | Description                                                                       |
|------------|--------------------|-----------------------------------------------------------------------------------|
| `-url`      | Big Buck Bunny URL | Video URL to download and reply with.                                             |
| `-conns`    | `4`                | Number of sub-connections in the upload pool (also the upload thread limit).      |
| `-pool`     | `true`             | Use the dedicated upload pool. Set `-pool=false` to upload over the main connection. |
| `-cut`      | `0` (off)          | Cut the video to this duration via ffmpeg `-t`, e.g. `30s` (requires ffmpeg).      |
| `-max-size` | `0` (off)          | Compress the video to approximately fit within this many megabytes (requires ffmpeg). |

Examples:

```console
go run ./bot-bigbuckbunny -url https://example.com/clip.mp4   # custom video
go run ./bot-bigbuckbunny -conns 8                            # larger upload pool
go run ./bot-bigbuckbunny -pool=false                        # single (main) connection
go run ./bot-bigbuckbunny -cut 30s                           # send only first 30s
go run ./bot-bigbuckbunny -max-size 10                        # compress to ~10 MiB
```
