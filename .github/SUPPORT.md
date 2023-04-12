This project is fully open-source and support is done voluntarily
by the community, so no SLA is provided.

Important news, major updates or security issues are posted in:

* [@gotd_news](https://t.me/gotd_news) — The gotd news channel

You can also join our groups:

* [@gotd_en](https://t.me/gotd_en) (English)
* [@gotd_ru](https://t.me/gotd_ru) (Russian)

Development and user support are provided in the groups above.

While we recommend using test servers, you can join [@gotd_test](https://t.me/gotd_test) for testing in production.

## How to not get banned?

**Do not share your app's ID and hash!**
They cannot be regenerated and are bound to your Telegram account.

> Note that all clients are strictly monitored to prevent abuse.

> If you try to use a Telegram client for flooding, spamming, faking subscribers and view counts you will be banned permanently.

> Due to excessive abuse of the Telegram API, **all accounts that sign up or
> log in using unofficial Telegram clients are automatically
> put under observation** to avoid violations of the [Terms of Service](https://core.telegram.org/api/terms).
> &mdash; <cite>[Official documentation][1]</cite>

[1]: https://core.telegram.org/api/obtaining_api_id
[2]: https://docs.madelineproto.xyz/docs/LOGIN.html

## Readme from MadelineProto

There is a useful quote from [MadelineProto][2] docs:

### Getting permission to use the telegram API
Before you start using MadelineProto with the Telegram API, you have to understand that Telegram strictly controls userbots created on their platform.

If you use MadelineProto or any other Telegram client, **including official clients**, for flooding, spamming, botting channels, you **will be banned forever**.

Due to excessive abuse of the Telegram API, **all phone numbers** that sign up or log in using **even official or approved by telegram API clients like MadelineProto** are automatically put **under observation**, and **may** get banned **even if you did nothing wrong**, simply because some internal flags are triggered in the **telegram servers**.

To avoid this, you must let telegram know that you will use your account with a userbot:
When or before you first sign up or login you have to **send an email with the phone number to recover@telegram.org**; explaining **what will your userbot do**.
**DO NOT LIE**, just tell them what will you do with the userbot you intend to create, asking them to not ban your account.

If your account does get banned, write to recover@telegram.org asking to unban it.

**DO NOT ABUSE OF THIS OR ANY OTHER API FOR FLOODING; SPAMMING; BOTTING**; the consequences will fall not only on you, but on all other users of this and other libraries, and even normal users.

There were cases when several **normal user accounts that did nothing wrong** were banned when telegram deployed a new spambot detection system: this is bad for the community, and it is bad for Telegram, so please do not abuse.


So, here's a summary:

1. This client is unofficial, Telegram treats such clients suspiciously, especially fresh ones.
2. Use bots whenever possible.
3. If you still want to automate things with a user, use it passively (i.e. receive more than sending).
4. When using it with a user:
   * Do not use QR code login, this will result in permaban.
   * Do it with extreme care.
   * Do not use VoIP numbers.
   * Do not abuse, spam or use it for other suspicious activities.
   * Implement a rate limiting system.
   * _Generally_, this is bad idea if you're not 100% sure what you're doing.

Bad usages of the API can trigger Telegram's anti-abuse system and ban all your accounts forever.

## What to do if I got banned?

First of all, there's no reason to panic. The automated anti-abuse system makes incorrect bans often.

See [discussions](https://github.com/lonamiwebs/telethon/issues/824#issuecomment-432182634) in other Telegram API libraries
for more context.

Second, write to [recover@telegram.org](mailto:recover@telegram.org) explaining what you intend to do with the API,
asking to unban your account.

Third, if you follow the "How to not get banned?" recommendations and suspect that
something in this library can trigger anti-abuse system, create an issue with
detailed description of what you were doing.
