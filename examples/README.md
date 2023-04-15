# Examples

Most examples use environment variable client builders.
You can do it manually, see `bot-auth-manual` for example.

1. Obtain [api_id and api_hash](https://core.telegram.org/api/obtaining_api_id) for your application and set as `APP_ID`, `APP_HASH`
2. Set `SESSION_FILE` to something like `~/session.yourbot.json` for persistent auth
3. Run example.

Please don't share `APP_ID` or `APP_HASH`, it can't be easily rotated.

| Name                                       | Description                                                 | Features                                                                                                   |
|--------------------------------------------|-------------------------------------------------------------|------------------------------------------------------------------------------------------------------------|
| [userbot](userbot/main.go)                 | Userbot example with peer storage and flood wait middelware | Custom auth flow, `session.Storage`, `PeerStorage`, `ResolverCache`                                        |
| [bot-auth-manual](bot-auth-manual/main.go) | Bot authentication                                          | `session.Storage`, setup without environment variables                                                     |
| [bot-echo](bot-echo/main.go)               | Echo bot                                                    | UpdateDispatcher, message sender                                                                           |
| [bot-upload](bot-upload/main.go)           | One-shot uploader for bot                                   | NoUpdates flag, uploads with MIME, custom file name and as audio, resolving peer by username, HTML message |
| [gif-download](gif-download/main.go)       | Saved gif backup (and restore) for user                     | Download, upload, middlewares with rate limit, unpack                                                      |
| [bg-run](bg-run/main.go)                   | Using client without Run                                    | contrib/bg package                                                                                         |
| [pretty-print](pretty-print/main.go)       | Pretty-print requests, responses and updates                | The tgp package, middleware and custom UpdateHandler for all updates                                       |
| [updates](updates/main.go)                 | Updates engine example                                      | The `updates` package that recovers missed updates                                                         |

## Environment variables

| Name           | Description                                                                           |
|----------------|---------------------------------------------------------------------------------------|
| `BOT_TOKEN`    | Token from [BotFather](https://core.telegram.org/bots#6-botfather)                    |
| `APP_ID`       | **api_id** of Telegram app from [my.telegram.org](https://my.telegram.org/apps)       |
| `APP_HASH`     | **api_hash** of Telegram app from [my.telegram.org](https://my.telegram.org/apps)     |
| `SESSION_FILE` | Path to session file, like `/home/super-bot/.gotd/session.super-bot.json`             |
| `SESSION_DIR`  | Path to session directory, if `SESSION_FILE` is not set, like `/home/super-bot/.gotd` |

## Support

Still don't know how to use specific features? See [user support](../.github/SUPPORT.md).
