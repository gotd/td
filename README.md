# td

Telegram protocol implementation based on [TL parser](https://github.com/ernado/tl) and inspired by
[grammers](https://github.com/Lonami/grammers).

## Status

Work in progress. Only go1.15 is supported.

Goal of this project is to implement Telegram client while
providing building blocks for the other client or even server
implementation without performance bottlenecks.

## Reference

The MTProto protocol description is [hosted](https://core.telegram.org/mtproto#general-description) by Telegram.

Most important parts for client impelemtations:
* [Security guidelines](https://core.telegram.org/mtproto/security_guidelines) for client software developers

Current implementation does not conform to security guidelines and should be used only
as reference or for testing.
