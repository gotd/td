# Examples

Most examples use environment variable client builders.
You can do it manually, see `bot-auth-manual` for example.

1. Go to [https://my.telegram.org/apps](https://my.telegram.org/apps) and grab `APP_ID`, `APP_HASH`
2. Set `SESSION_FILE` to something like `~/session.yourbot.json` for persistent auth
3. Run example.

Please don't share `APP_ID` or `APP_HASH`, it can't be easily rotated.

## Environment variables

| Name             | Description
|------------------|---------------
| `BOT_TOKEN`      | Token from [BotFather](https://telegram.me/BotFather)
| `APP_ID`         | **api_id** of Telegram app from [my.telegram.org](https://my.telegram.org/apps)
| `APP_HASH`       | **api_hash** of Telegram app from [my.telegram.org](https://my.telegram.org/apps)
| `SESSION_FILE`   | Path to session file, like `/home/super-bot/.gotd/session.super-bot.json`
| `SESSION_DIR`    | Path to session directory, if `SESSION_FILE` is not set, like `/home/super-bot/.gotd`
