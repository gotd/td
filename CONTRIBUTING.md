# Contributing

See [architecture page](ARCHITECTURE.md) for understanding gotd.

This project follows [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/).

Before creating pull requests, please read the [coding guidelines](https://github.com/uber-go/guide/blob/master/style.md) and
follow some existing [pull requests](https://github.com/nnqq/td/pulls).

General tradeoffs:

* Less is more
* Maintainability > feature bloat
* Simplicity > speed
* Consistency > elegance

## Testing

Use a **test server**! Don't test on production servers!

Each phone number is limited to only a certain amount of logins per day (e.g. 5, but this is subject to change) after which the API will return a flood error until the next day.
This might not be enough for testing the implementation of User Authorization
flows in client applications.

### Test servers

You can use `AddrTest` with `TestAppID` and `TestAppHash` to connect to test servers.

It is also possible to use [test phone numbers](https://core.telegram.org/api/auth#test-phone-numbers) directly when connecting or
via `TestAuth` helper.

### Our testing group

The [@gotd_test](https://t.me/gotd_test) group can be used to test clients
on production servers, it should be relatively safe to test updates handling (i.e. passive) functions like that.

As always, we recommend test servers instead.

## Optimizations

Please provide [benchcmp](https://godoc.org/golang.org/x/tools/cmd/benchcmp) output if your PR
tries to optimize something.

Note that in most cases, readability is more important than speed.

## Features

Please check the [projects](https://github.com/nnqq/td/projects) page for features that
are on the roadmap. If you have an idea for a new feature, please open a feature request first.

Also, it will be great to [contact](.github/SUPPORT.md) the developers to discuss implementation
details.

## Updating schema

If a new layer is released in [tdesktop](https://github.com/telegramdesktop/tdesktop) repo, you can do this to update it:

```console
$ make download_schema generate
```

Please don't do it too early, because it is possible to have multiple versions of
layers.

## Fuzzing

This project uses fuzzing to increase overall stability and decrease
possibility of DOS attacks.

To prepare the fuzzing binary, do this:

```console
$ make fuzz_telegram_build
```

Now you can start the fuzzer locally:

```console
$ make fuzz_telegram
```

Please refer to [dvyukov/go-fuzz](https://github.com/dvyukov/go-fuzz) for advanced usage.

## Testing allocations

Please test that hot paths are not allocating too much.

```go
func TestBuffer_ResetN(t *testing.T) {
    var b Buffer
    testutil.ZeroAlloc(t, func() {
        b.ResetN(1024)
    })
}

func TestAllocs(t *testing.T) {
    const allocThreshold = 512

    testutil.MaxAlloc(t, allocThreshold, func() {
        _ = c.handleMessage(&bin.Buffer{Buf: data})
    })
}
```

## Coding guidance

Please read [Uber code style](https://github.com/uber-go/guide/blob/master/style.md).

### Newlines

Don't move the first argument to next line if it is not grouped
with other arguments.

<table>
<thead><tr><th>Bad</th><th>Good</th></tr></thead>
<tbody>
<tr><td>

```go
log.Info(
    "init_config",
    zap.Duration("retry_interval", cfg.RetryInterval),
    zap.Int("max_retries", cfg.MaxRetries),
)
```

</td><td>

```go
log.Info("init_config",
    zap.Duration("retry_interval", cfg.RetryInterval),
    zap.Int("max_retries", cfg.MaxRetries),
)
```

</td></tr>
</tbody></table>
