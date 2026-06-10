# Open Issue Triage

Triage of all open issues as of 2026-06-11. Grouped into **Invalid**, **Already fixed**,
**Outdated**, and **Can be fixed**. Schema/code references verified against the current tree
(`tg/`, `_schema/`, `telegram/`).

| Group | Count |
|---|---|
| [Invalid](#-invalid) | 24 |
| [Already fixed](#-already-fixed) | 7 |
| [Outdated](#-outdated) | 9 |
| [Can be fixed](#-can-be-fixed) | 36 |

---

## 🔴 Invalid

Not library bugs — usage errors, Telegram-API behavior/errors, or pure questions.

| # | Title | Reason |
|---|---|---|
| 1737 | Is anyone still maintaining this project? | Question, answered by maintainer. |
| 1612 | how can i use sendResoldGift method? | Usage question. |
| 1582 | Implement the server with scheme.tl? | Open-ended question. |
| 1553 | OnNewChannelMessage `CHANNEL_INVALID` | Needs valid channel access hash; Telegram-side usage. |
| 1544 | Update User offline status not work | Server-side offline semantics, not a lib bug. |
| 1528 | 32bit android cpu, no OTP | Device/Telegram-side, not reproducible as a lib bug. |
| 1495 | Can't read messages in old groups | Telegram update/migration behavior, no lib defect shown. |
| 1493 | `KeyboardButtonSimpleWebView` `BUTTON_TYPE_INVALID` | Button type only valid in specific contexts; Telegram-side. |
| 1456 | `AUTH_KEY_UNREGISTERED` on README example | Session/credential/env issue, not a code bug. |
| 1454 | SignIn fail | Incomplete repro, usage. |
| 1451 | Can I set a proxy login? | Supported via custom Resolver/dialer; question. |
| 1389 | `MessagesForwardMessages` `CHANNEL_INVALID` | Missing access hash; Telegram-side usage. |
| 1369 | phone_code missing error | Usage error. |
| 1340 | `MessagesSearch` always returns nil | Telegram pinned-filter behavior; usage. |
| 1279 | `GetMessageLink` `INPUT_METHOD_INVALID` | Calling a TDLib-only method over raw MTProto; misuse. |
| 1154 | How to get participants? | Question. |
| 1138 | My account has been banned | Telegram-side, out of scope. |
| 1098 | invokeWithTakeout | Already possible via raw `InvokeWithTakeout`; question. |
| 946 | `CHAT_ID_INVALID` on `MessagesAddChatUser` | Telegram-side usage. |
| 923 | Can't pass ResolveUsername peer to GetHistory | Type mismatch in user code (helper req is #283). |
| 790 | support socks5 and mtproto proxy | Supported via dialer/contrib; out of scope here. |
| 568 | Who is using gotd? | Meta/marketing thread, not actionable. |
| 544 | connect gotd to pion | Open question, out of scope. |
| 1451 | (see above) | |

---

## 🟢 Already fixed

Verified against current `tg/` and `_schema/`.

| # | Title | Resolution |
|---|---|---|
| 1583 | `payments.getUserStarGifts` not exists | Method renamed; `payments.getSavedStarGifts#a319e569` now in schema (`tg/tl_payments_get_saved_star_gifts_gen.go`). |
| 1027 | CDN downloader implementation | Implemented: `telegram/downloader/cdn.go`, `cdn_plan.go`, `cdn_state_machine.go`, `cdn_verify.go`. |
| 1375 | `MessagesInvitedUsers` does not implement `UpdatesClass` | Type now carries `Updates UpdatesClass` + `GetUpdates()` (`tl_messages_invited_users_gen.go`); was a stale-dependency mismatch. |
| 1166 | decode `Vector<User>` unexpected id `0x8f97c628` | Schema was behind; constructor no longer present, schema kept at latest layer. |
| 1548 | decode `userFull` unexpected id `0x979d2376` | Same root cause; resolved by ongoing schema updates. |
| 288 | FromID same for both parties in 1:1 chat | Very old (v0.33.3); message-peer handling reworked since. |
| 789 | examples are missing (broken link) | `examples/` directory exists in repo today. |

---

## 🟡 Outdated

Possibly valid once, but tied to old versions/context and now stale or unverifiable as written.

| # | Title | Note |
|---|---|---|
| 1689 | `ChannelsEditAdmin` broken in v0.141 (`USER_CREATOR`) | Likely Telegram-side rule (can't edit creator's rights); no repro since. |
| 1479 | Supergroup messages not received + wrong ChatID | Stale; ChatID confusion is documented MTProto behavior. |
| 1385 | Blocks when using `bg.connect` | Overlaps connection-recovery work; no recent repro. |
| 1363 | `MessagesSendMessage` freezes forever | Same family as #1030; stale, superseded. |
| 1203 | Client ping panics on direct call | `internal/mtproto` restructured since v0.88; needs re-confirm. |
| 731 | client: bg-run failed | Ancient (v0.55.2, 2022). |
| 704 | client: rpc not responding | Incomplete, ancient (v0.55.2, 2022). |
| 825 | feat(tdesktop): support key_datas | "May be no key_data in latest format"; format moved on. |
| 199 | e2e: improve server (epic) | Largely superseded by later server work; stale checklist. |

---

## 🔵 Can be fixed

Legitimate open bugs and actionable enhancements. Tracked in the backlog issue.

### Bugs / correctness

| # | Title |
|---|---|
| 1725 | `*tg.UploadGetFileRequest` get dc5 video time out (wrapped transport timeouts not retried) |
| 1658 | DEADLOCK in updates.Manager when starting with many channels w/ unread messages |
| 1623 | Some `UpdateNewMessage` updates are not received in handler |
| 1572 | `rpc_result`/`msg_container` decoded as gzip when not compressed |
| 1382 | Update postponed and handled only with next update |
| 1322 | Repeated errors in Updates Recovery for channels becoming private |
| 1030 | client can't recover from connection loss (help wanted) |

### Enhancements / helpers

| # | Title |
|---|---|
| 1527 | `MessagesSendMultiMedia` `MEDIA_INVALID` when sending multiple media |
| 1510 | Link Preview Options helper |
| 1474 | Set Spoiler via `UploadedPhotoBuilder` |
| 1406 | Update FakeTLS ClientHello to match modern clients |
| 1362 | Phone call function |
| 1308 | Handling `UpdateConnectionState` |
| 1267 | Channel recommendations pagination |
| 884 | helper: support messages/GetMediaGroup |
| 883 | clock: support network clock |
| 824 | feat: errors with placeholders like `%d` |
| 816 | uploader: compute part size automatically |
| 788 | invites: support `tg.ChatInvitePublicJoinRequests` |
| 755 | auth: allow safer password passing |
| 689 | Callback if user/channel state fails to load |
| 615 | auth: helpers for (re)setting/updating/recovering password |
| 597 | bot: fix inspection of service messages |
| 392 | mtproto: containerize small messages |
| 376 | gen: derive mappings with parameters |
| 283 | query: generate resolve helpers when query needs peer parameter |
| 263 | client: improve FLOOD_WAIT handling |
| 217 | client: get-by-id helpers |
| 214 | message: Markdown styling for text messages |
| 189 | message: sticker helpers |
| 188 | client: admin helpers |
| 172 | client: add OpenTelemetry tracing |
| 166 | doc: add examples for every feature |
| 164 | proto: sequential calls using `invokeAfterMsg(s)` |
| 163 | proto: full service message support |
| 128 | Accept interface instead of `*zap.Logger` |

> Note: #1527 is filed as a Telegram `MEDIA_INVALID` usage error, but the long comment thread
> suggests a real ergonomics/helper gap around multi-media sending, so it is kept here for review.
