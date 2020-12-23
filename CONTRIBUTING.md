# Contributing

This project uses [Conventional commits](https://www.conventionalcommits.org/en/v1.0.0/).

Before creating pull request, please read [coding guidelines](https://wiki.crdb.io/wiki/spaces/CRDB/pages/181371303/Go+coding+guidelines) and
follow some existing [pull requests](https://github.com/gotd/td/pulls).

General tradeoffs:
* Less is more
* Maintainability > feature bloat
* Simplicity > speed
* Consistency > elegancy

## Optimizations

Please provide [benchcmp](https://godoc.org/golang.org/x/tools/cmd/benchcmp) output if your PR
tries to optimize something.

Note that in most cases readability is more important that speed.


## Features

Please check [projects](https://github.com/gotd/td/projects) page for features that
are on roadmap. If you have idea for new feature, please open feature request first.

Also, it will be great to [contact](.github/SUPPORT.md) developers to discuss implementation
details.

## Schema update

If new layer is released in [tdesktop](https://github.com/telegramdesktop/tdesktop) repo, one can
use it to update to latest schema:

```console
$ make download_schema generate
```

Please don't do it too early, because it is possible to have multiple versions of
layer.

## Fuzzing

This project uses fuzzing to increase overall stability and decrease
possibility of DOS attacks.

To start fuzzing locally, install [dvyukov/go-fuzz](https://github.com/dvyukov/go-fuzz):
```console
# Using temp directory to avoid modifying current go.mod.
$ mkdir /tmp/fuzz && cd /tmp/fuzz
$ GO111MODULE=on go get github.com/dvyukov/go-fuzz/go-fuzz github.com/dvyukov/go-fuzz/go-fuzz-build
```

After that, you will be able to prepare fuzzing target binary:
```console
$ make fuzz_telegram_build
```
Now you can start fuzzer locally:
```console
$ make fuzz_telegram
```
Please refer to [dvyukov/go-fuzz](https://github.com/dvyukov/go-fuzz) for advanced usage.
