This project is fully opensource and support is done voluntarily
by community, so no SLA is provided.

**Only** important news, major updates or security issues are posted here:
* [@gotd_news](https://t.me/gotd_news) — gotd news telegram channel

You can also use following telegram groups:

* **[@gotd_en](https://t.me/gotd_en) (English)**
* [@gotd_ru](https://t.me/gotd_ru) (Russian)

Both development and user support is done in [@gotd_en](https://t.me/gotd_en).
Use [@gotd_test](https://t.me/gotd_test) if you want to test client,
but doing it on staging server is better.

## How not to get banned

**Do not share your application id and hash!**
They cannot be rotated and are bound to your Telegram account.

> Before using the MTProto Telegram API, please note that all API client
> libraries are strictly monitored to prevent abuse.

> If you use the Telegram API for flooding, spamming, faking subscriber and
> view counters of channels, you will be banned forever.

> Due to excessive abuse of the Telegram API, **all accounts that sign up or
> log in using unofficial Telegram API clients are automatically
> put under observation** to avoid violations of the [Terms of Service](https://core.telegram.org/api/terms).
> &mdash; <cite>[Official documentation][1]</cite>

[1]: https://core.telegram.org/api/obtaining_api_id

So, some summary:

1) This client is unofficial, Telegram treats such clients suspiciously, especially fresh ones.
2) Use it only for Bots (but don't abuse).
3) If you still want "userbot", use it passively (i.e. just receive updates and not send anything).
4) If you want to implement active "userbot"
   * Do not use QR code login, this will result in permaban
   * Do it with extreme care
   * Do not use voip accounts
   * Do not abuse, spam or use it for other suspicious activities
   * Implement rate limiting
   * But *in general* it is bad idea if you are not 100% know what you are doing

Other usages can trigger Telegram anti-abuse system and ban all your accounts forever.

## What to do if I'm banned

First, try not to panic, anti-abuse system often makes false-positive bans.
See [discussions](https://github.com/lonamiwebs/telethon/issues/824#issuecomment-432182634) in other Telegram API libraries
for more context.

Second, write to [recover@telegram.org](mailto:recover@telegram.org) explaining what you intend to do with the API,
asking to unban your account.

Third, if you follow "How not to get banned" recommendations and suspect that
something in this library can trigger anti-abuse system, create issue with
detailed description of what you were doing.
