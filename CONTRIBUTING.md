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

Also it will be great to [contact](.github/SUPPORT.md) developers to discuss implementation
defailts.
