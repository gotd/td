# Examples

Most examples use environment variable client builders.
You can do it manually, see `bot-auth-manual` for example.

1. Go to [https://my.telegram.org/apps](https://my.telegram.org/apps) and grab `APP_ID`, `APP_HASH`
2. Set `SESSION_FILE` to something like `~/session.yourbot.json` for persistent auth
3. Run example.

Please don't share `APP_ID` or `APP_HASH`, it can't be easily rotated.
