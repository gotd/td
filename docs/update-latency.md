# Measuring update latency

How to find where time is lost between a message being posted and your
`UpdateHandler` seeing it, and how to share the result without sensitive data.

## What is measured

The client times every batch of updates pushed by the server:

```
message date ──► msg_id time ──► read from socket ──► handler called ──► handler returned
       server.delay     receive.delay      process.duration      handle.duration
```

| Metric (histogram, seconds) | What it is | Whose time |
| --- | --- | --- |
| `gotd.client.updates.server.delay` | Message date to the creation of the server message that pushed it. Attribute `kind`: `channel` or `other`. | Telegram |
| `gotd.client.updates.receive.delay` | Creation of the server message to its read from the socket, on top of the fastest delivery seen. | Network |
| `gotd.client.updates.process.duration` | Read from the socket to the handler call: decryption, gzip, decoding, scheduling. | gotd, CPU, GC |
| `gotd.client.updates.handle.duration` | The `UpdateHandler` call. Attribute `result`: `ok` or `error`. | Your code |

The same numbers go to the debug log as `Update timing`, one record per batch.

Limits:

- The message date is in whole seconds, so `server.delay` is up to a second
  larger than the real delay. Compare distributions, not single values.
- `receive.delay` is not the whole network time. The client learns the
  difference between its clock and the server's from the fastest message it
  has received, so that message counts as zero and the rest are measured
  against it. A wrong local clock does not distort it. The fixed part of the
  path is about half of a ping round trip to the DC.
- Only pushed updates are timed. Messages fetched with `updates.getDifference`
  or `updates.getChannelDifference` are not, and a message the server never
  pushed is not there at all. Compare the count of `server.delay` with the
  number of messages posted.
- Edits are not counted in `server.delay`: they carry the date of the original
  message.

## Collect

Metrics are off by default. Pass a provider to the client. This one keeps
everything in memory and writes one JSON file; `resource.Empty()` keeps the
host and binary names out of it.

```go
reader := sdkmetric.NewManualReader()
provider := sdkmetric.NewMeterProvider(
	sdkmetric.WithReader(reader),
	sdkmetric.WithResource(resource.Empty()),
)
client := telegram.NewClient(appID, appHash, telegram.Options{
	MeterProvider: provider,
	// ...
})

// Call on shutdown, or from a timer. Counts are cumulative, so the last
// file has everything.
func dump(ctx context.Context, reader *sdkmetric.ManualReader, path string) error {
	var rm metricdata.ResourceMetrics
	if err := reader.Collect(ctx, &rm); err != nil {
		return err
	}
	data, err := json.MarshalIndent(rm, "", " ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}
```

Imports: `go.opentelemetry.io/otel/sdk/metric` as `sdkmetric`,
`go.opentelemetry.io/otel/sdk/metric/metricdata`, `go.opentelemetry.io/otel/sdk/resource`.

If you already export OpenTelemetry metrics, skip this and query the four
histograms in your backend.

Histograms have fixed buckets. For exact values, also keep the `Update timing`
log records. This logger passes them through as info and drops the rest of the
debug output:

```go
type timingLogger struct{ log.Logger }

func (timingLogger) Enabled(context.Context, log.Level) bool { return true }

func (l timingLogger) Log(ctx context.Context, level log.Level, msg string, attrs ...log.Attr) {
	switch {
	case msg == "Update timing":
		l.Logger.Log(ctx, log.LevelInfo, msg, attrs...)
	case level >= log.LevelInfo:
		l.Logger.Log(ctx, level, msg, attrs...)
	}
}

// telegram.Options{Logger: timingLogger{logslog.New(slog.Default())}}
```

`log` is `github.com/gotd/log`.

Run the process with `TZ=UTC`. Otherwise timestamps in the file and in the
log carry your UTC offset.

Run it for at least a day, across your busiest hours.

## Read

| Symptom | Meaning |
| --- | --- |
| `server.delay` is large, the rest is small | The server pushed the message late. No client change helps; look at the account, its DC and how it follows the channel. |
| Fewer `server.delay` points than messages posted | Some messages were never pushed to this session. |
| `receive.delay` is large or spread wide | Unstable network path to the DC: some updates arrive much later than the fastest ones. |
| `process.duration` above a few milliseconds | The client is slow: CPU starvation, GC pauses, or a gotd problem. Report it. |
| `handle.duration` is large | Your handler. Each batch is handled in its own goroutine, so a slow handler does not delay the next read, but it delays that batch. |

## Share

Safe to send:

- the JSON file: it has metric names, bucket counts, sums and timestamps. It
  includes the other `gotd.client.*` metrics too: DC numbers, error codes and
  TL method names of timed out requests;
- `Update timing` log records: durations and message counts;
- the gotd version, the DC of the account, the transport, the number of
  channels and how long the measurement ran.

Neither the metrics nor these log records contain peers, message IDs, text,
phone numbers or keys.

Do not send the session file, app ID and hash, phone numbers, channel names or
IDs, or full debug logs. To compare channels, measure in your handler and
label them yourself (`channel A`, `channel B`).
