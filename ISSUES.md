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

## Progress

- **Done (PR):** #1474 → PR #1743 (Spoiler on media builders), #1510 → PR #1744
  (InvertMedia + WebPage link-preview builder), #884 → PR #1745 (GetMediaGroup helper),
  #615 → password recovery helpers (`RequestPasswordRecovery`/`CheckRecoveryPassword`/`RecoverPassword`
  in `telegram/auth/password.go`), #883 → NTP network clock (`clock/ntp`, isolated subpackage so
  `beevik/ntp` stays out of the core dependency tree), #689 → updates state-load callbacks
  (`OnLoadUserStateFailed`/`OnLoadChannelStateFailed` in `telegram/updates`), #1030/#1021 →
  connection-loss recovery: requests not processed by the server (not sent, or sent but not
  acknowledged) are transparently retried on the new connection (`rpc` close cause,
  `telegram.Client.invokeConn` wait-for-reconnect, `pool.DC.Invoke` re-acquire), #1308 →
  primary connection state hook (`Options.OnConnectionState` with
  connecting/ready/disconnected `ConnectionState`), #816 → uploader auto part-size
  (grows part size from file size to keep parts within the 3999 limit, preventing
  `FILE_PARTS_INVALID` on large files; explicit `WithPartSize` still respected), #217 →
  get-messages-by-ID helper (`messages.QueryBuilder.GetMessages`, channel-aware;
  users/channels already covered by `telegram/peers`), #825 → tdesktop modern `key_datas`
  layout: `session/tdesktop` now prefers the modern `s`-suffixed file over legacy `0`/`1`
  (matching Telegram Desktop), so a stale legacy file no longer shadows the current one;
  added an end-to-end modern-layout read test, #376 → parameterized `As`-mappers: the
  code generator now exposes target fields that can't be derived from the source
  constructor as method parameters instead of skipping the mapping. `thumb_size` on
  file-location inputs is supplied by the caller, so `Document`/`InputDocument` gain
  `AsInputDocumentFileLocation(thumbSize string)` and `Photo`/`InputPhoto` get a new
  `AsInputPhotoFileLocation(thumbSize string)` (previously dropped because `thumb_size`
  is required); `telegram/query/messages` now uses these mappers instead of building the
  location structs by hand.
- **Done (partial — tgtest server epic):** #199 → personal-messages service
  (`tgtest/services/messages`): implements `messages.sendMessage` (stores the
  dialog, echoes the sent message back as `UpdateMessageID`+`UpdateNewMessage`)
  and `messages.getHistory`, and delivers messages to connected recipients as
  `UpdateNewMessage` updates. Only user peers (`InputPeerSelf`/`InputPeerUser`)
  are supported. The other epic checklist items were already covered by later
  work (key exchange → `exchange` pkg, up/downloads → `services/file`, multi-DC
  → `tgtest/cluster`).
- **Done (partial — admin helpers):** #188 → channel/supergroup admin helpers in
  `telegram/peers`: username management (`Channel.SetUsername`/`CheckUsername`/`DeactivateAllUsernames`),
  photo (`Channel.SetPhoto`/`DeletePhoto`), supergroup toggles
  (`Supergroup.TogglePreHistoryHidden`/`ToggleJoinToSend`/`ToggleJoinRequest`), recent-actions admin
  log (`Channel.AdminLog()` query builder with `Search`/`Filter`/`Admins`/`ForEach`/`Fetch`), and
  member convenience wrappers in `telegram/peers/members`
  (`ChannelMembers.Promote`/`Demote`/`Ban`/`Unban`). Builds on the already-existing title/description/
  reactions/slow-mode/signatures/sticker-set/discussion-group helpers; creator transfer
  (`channels.editCreator`, SRP-gated) intentionally left out.
- **Done:** #1267 → channel recommendations helper
  (`Channel.RecommendedChannels` / `Manager.RecommendedChannels` in `telegram/peers`):
  the raw `channels.getChannelRecommendations` has no offset/hash, so true pagination
  isn't an API feature — instead the helper surfaces the total `Count` from
  `messages.chatsSlice` (Premium gets the full list; others a capped subset) so callers
  can detect that more recommendations exist.
- **Closed as already-addressed:** #824 (`tgerr.Error` already extracts `Type`/`Argument`).
- **Found already implemented** (should be closed, not built): #214 Markdown styling
  (`telegram/message/markdown`), #189 sticker helpers (`telegram/query/cached` generates all 8).

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
| 825 | feat(tdesktop): support key_datas | **done — see Progress** (modern `s`-suffixed file now preferred over legacy `0`/`1`). |
| 199 | e2e: improve server (epic) | Largely superseded; remaining gap (personal messages) **done** — `tgtest/services/messages`. |

---

## 🔵 Can be fixed

Legitimate open bugs and actionable enhancements. Tracked in the backlog issue.

### Bugs / correctness

| # | Title | Status |
|---|---|---|
| 1725 | `*tg.UploadGetFileRequest` get dc5 video time out (wrapped transport timeouts not retried) | in-progress branch |
| 1658 | DEADLOCK in updates.Manager when starting with many channels w/ unread messages | open |
| 1623 | Some `UpdateNewMessage` updates are not received in handler | open |
| 1572 | `rpc_result`/`msg_container` decoded as gzip when not compressed | open |
| 1382 | Update postponed and handled only with next update | open |
| 1322 | Repeated errors in Updates Recovery for channels becoming private | open |
| 1030 | client can't recover from connection loss (help wanted) | **done — see Progress** |

### Enhancements / helpers

| # | Title | Status |
|---|---|---|
| 1527 | `MessagesSendMultiMedia` `MEDIA_INVALID` when sending multiple media | open |
| 1510 | Link Preview Options helper | **done — PR #1744** |
| 1474 | Set Spoiler via `UploadedPhotoBuilder` | **done — PR #1743** |
| 1406 | Update FakeTLS ClientHello to match modern clients | open |
| 1362 | Phone call function | **in progress — `telegram/calls` (1:1 VoIP, tgcalls v2)** |
| 1308 | Handling `UpdateConnectionState` | **done — see Progress** |
| 1267 | Channel recommendations pagination | **done — `Channel.RecommendedChannels` (+ total Count)** |
| 884 | helper: support messages/GetMediaGroup | **done — PR #1745** |
| 883 | clock: support network clock | **done — `clock/ntp`** |
| 824 | feat: errors with placeholders like `%d` | **closed — already addressed** |
| 816 | uploader: compute part size automatically | **done — see Progress** |
| 788 | invites: support `tg.ChatInvitePublicJoinRequests` | open |
| 755 | auth: allow safer password passing | **done — `PasswordWith` + `telegram/auth/srpguard` (memguard)** |
| 689 | Callback if user/channel state fails to load | **done — load callbacks added** |
| 615 | auth: helpers for (re)setting/updating/recovering password | **done — recovery helpers added** |
| 597 | bot: fix inspection of service messages | open (lives in `gotd/bot`) |
| 392 | mtproto: containerize small messages | open |
| 376 | gen: derive mappings with parameters | **done — see Progress** |
| 283 | query: generate resolve helpers when query needs peer parameter | open |
| 263 | client: improve FLOOD_WAIT handling | open |
| 217 | client: get-by-id helpers | **done — see Progress** |
| 214 | message: Markdown styling for text messages | **already implemented** (`telegram/message/markdown`) |
| 189 | message: sticker helpers | **already implemented** (`telegram/query/cached`) |
| 188 | client: admin helpers | **done (partial) — see Progress** (username/photo/toggles/admin-log/promote-demote-ban in `telegram/peers`; creator transfer left out) |
| 172 | client: add OpenTelemetry tracing | open |
| 166 | doc: add examples for every feature | open |
| 164 | proto: sequential calls using `invokeAfterMsg(s)` | open |
| 163 | proto: full service message support | open (mostly done) |
| 128 | Accept interface instead of `*zap.Logger` | open |

> Note: #1527 is filed as a Telegram `MEDIA_INVALID` usage error, but the long comment thread
> suggests a real ergonomics/helper gap around multi-media sending, so it is kept here for review.

### Difficulty analysis (remaining, open items)

Effort estimates from probing the current tree. Two items above turned out to be already
implemented (#214, #189) and one was closed (#824); they are excluded here.

**Tier 1 — easiest** (small, localized, mock-testable offline, no new concepts):

| # | Title | Why easy |
|---|---|---|
| 615 | `auth.recoverPassword` helper | **done** — `RequestPasswordRecovery`/`CheckRecoveryPassword`/`RecoverPassword` added to `telegram/auth/password.go`. |
| 689 | updates state-load callbacks | Add `OnLoadUserStateFailed`/`OnLoadChannelStateFailed` to `updates.Config` + invoke at load sites, mirroring the existing `OnChannelTooLong`. |
| 883 | network clock (NTP) | `clock/` already defines the `Clock` interface; NTP impl is self-contained. Friction: adds a `beevik/ntp` dependency. |

**Tier 2 — moderate** (new code, contained):

| # | Title | Note |
|---|---|---|
| 217 | get-by-id helpers | **done** — `messages.QueryBuilder.GetMessages` (by-ID, channel-aware). Optional hash-based caching left out as a "nice to have". |
| 788 | `ChatInvitePublicJoinRequests` | Requires loosening the `InviteLink` type, currently hardwired to `ChatInviteExported`. |
| 1308 | connection-state updates | `OnSession` exists; add a connect/disconnect hook (no TDLib-style state machine). |
| 755 | safer password passing | Callback-based password hashing in `telegram/auth`. |
| 1267 | recommendations pagination | **done** — helper added (`telegram/peers`). No offset/hash on the raw method, so not truly paginable; helper exposes the total `Count` (Premium-gated full list). |

**Tier 3 — hard** (cross-cutting, breaking, or deep):

- Bugs #1658 / #1030 / #1382 / #1322 / #1623 — updates/connection concurrency and state-machine; hard to reproduce, risky.
- #1572 gzip decode — deep MTProto investigation, unclear repro.
- #128 logger interface — touches 38 non-test files, breaking API change.
- #392 outbound containerization — write-path perf, concurrency-sensitive.
- #1725 timeout — already has an in-progress branch.
- #1362 phone calls (VoIP), #1406 FakeTLS modernization (security-sensitive), #376 / #283 code-generator changes, #172 OpenTelemetry, #166 doc-examples epic.
- #597 — lives in the `gotd/bot` repo, not this one.
