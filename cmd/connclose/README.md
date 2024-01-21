This project is an attempt to reproduce connection timeout that causes gotd to hang.

> [!WARNING]
> Currently I'm unable to reproduce the issue.

Referenced issue: https://github.com/gotd/td/issues/1030

Dependencies:
- docker
- kind
- helm
- cilium cli
- go

## Running

Start cluster:

```bash
make up
```

Deny connections to telegram:
```bash
make deny
```

Update binary
```bash
make update
```

Restore connections
```bash
make allow
```

## Logs

```bash
make logs
```

## Cleanup

```bash
make down
```
