package tljson

import (
	"encoding/base64"
	"strconv"

	"github.com/go-faster/errors"
	"github.com/go-faster/jx"

	"github.com/gotd/td/tg"
)

// EmojiSound represents emoji sound file location.
//
// See https://core.telegram.org/api/animated-emojis#emojis-with-sounds.
type EmojiSound struct {
	ID            int64  `json:"id"`
	AccessHash    int64  `json:"access_hash"`
	FileReference []byte `json:"file_reference_base64"`
}

// DecodeJSON decodes EmojiSound.
func (e *EmojiSound) DecodeJSON(d *jx.Decoder) error {
	return d.ObjBytes(func(d *jx.Decoder, key []byte) error {
		switch string(key) {
		case "id":
			v, err := d.Str()
			if err != nil {
				return errors.Wrap(err, "decode id")
			}

			id, err := strconv.ParseInt(v, 10, 64)
			if err != nil {
				return errors.Wrap(err, "parse id")
			}
			e.ID = id
		case "access_hash":
			v, err := d.Str()
			if err != nil {
				return errors.Wrap(err, "decode access_hash")
			}

			accessHash, err := strconv.ParseInt(v, 10, 64)
			if err != nil {
				return errors.Wrap(err, "parse access_hash")
			}
			e.AccessHash = accessHash
		case "file_reference_base64":
			encoded, err := d.StrBytes()
			if err != nil {
				return errors.Wrap(err, "decode file_reference_base64")
			}
			encoding := base64.RawURLEncoding

			e.FileReference = make([]byte, encoding.DecodedLen(len(encoded)))
			n, err := encoding.Decode(e.FileReference, encoded)
			if err != nil {
				return errors.Wrap(err, "un-base64 file_reference_base64")
			}
			e.FileReference = e.FileReference[:n]
		default:
			return d.Skip()
		}

		return nil
	})
}

// EmojiSendDiceSuccess represents the winning dice value and the final frame of the animated sticker.
//
// See https://core.telegram.org/api/dice.
type EmojiSendDiceSuccess struct {
	Value      int `json:"value"`
	FrameStart int `json:"frame_start"`
}

// DecodeJSON decodes EmojiSendDiceSuccess.
func (e *EmojiSendDiceSuccess) DecodeJSON(d *jx.Decoder) error {
	return d.ObjBytes(func(d *jx.Decoder, key []byte) error {
		switch string(key) {
		case "value":
			v, err := d.Int()
			if err != nil {
				return errors.Wrap(err, "decode value")
			}
			e.Value = v
		case "frame_start":
			v, err := d.Int()
			if err != nil {
				return errors.Wrap(err, "decode frame_start")
			}
			e.FrameStart = v
		default:
			return d.Skip()
		}

		return nil
	})
}

// RoundVideoEncoding represents a set of recommended codec parameters for round videos.
type RoundVideoEncoding struct {
	Diameter     int `json:"diameter"`
	VideoBitrate int `json:"video_bitrate"`
	AudioBitrate int `json:"audio_bitrate"`
	MaxSize      int `json:"max_size"`
}

// DecodeJSON decodes RoundVideoEncoding.
func (e *RoundVideoEncoding) DecodeJSON(d *jx.Decoder) error {
	return d.ObjBytes(func(d *jx.Decoder, key []byte) error {
		switch string(key) {
		case "diameter":
			v, err := d.Int()
			if err != nil {
				return errors.Wrap(err, "decode diameter")
			}
			e.Diameter = v
		case "video_bitrate":
			v, err := d.Int()
			if err != nil {
				return errors.Wrap(err, "decode video_bitrate")
			}
			e.VideoBitrate = v
		case "audio_bitrate":
			v, err := d.Int()
			if err != nil {
				return errors.Wrap(err, "decode audio_bitrate")
			}
			e.AudioBitrate = v
		case "max_size":
			v, err := d.Int()
			if err != nil {
				return errors.Wrap(err, "decode max_size")
			}
			e.MaxSize = v
		default:
			return d.Skip()
		}

		return nil
	})
}

// AppConfig represents app config structure.
//
// See https://core.telegram.org/api/config#client-configuration.
type AppConfig struct {
	Test int `json:"test"`
	// Animated emojis and animated dice should be scaled by this factor before being shown to the user.
	EmojiesAnimatedZoom float64 `json:"emojies_animated_zoom"`
	// A list of supported animated dice stickers.
	EmojiesSendDice []string `json:"emojies_send_dice"`
	// For animated dice emojis other than the basic 🎲, indicates the winning dice value
	// and the final frame of the animated sticker, at which to show the fireworks.
	EmojiesSendDiceSuccess map[string]EmojiSendDiceSuccess `json:"emojies_send_dice_success"`
	// A map of soundbites to be played when the user clicks on the specified animated emoji.
	//
	// The file reference field should be base64-decoded before downloading the file.
	EmojiesSounds map[string]EmojiSound `json:"emojies_sounds"`
	// Specifies the name of the service providing GIF search through gif_search_username.
	GIFSearchBranding string `json:"gif_search_branding"`
	// Specifies a list of emojies that should be suggested as search term in a bar above the GIF search box.
	GIFSearchEmojies []string `json:"gif_search_emojies"`
	// Specifies that the app should not display local sticker suggestions
	// for emojis at all and just use the result of messages.getStickers.
	StickersEmojiSuggestOnlyAPI bool `json:"stickers_emoji_suggest_only_api"`
	// Specifies the validity period of the local cache of messages.getStickers,
	// also relevant when generating the pagination hash when invoking the method.
	StickersEmojiCacheTime        int    `json:"stickers_emoji_cache_time"`
	GroupCallVideoParticipantsMax int    `json:"groupcall_video_participants_max"`
	YoutubePIP                    string `json:"youtube_pip"`
	// Whether the Settings->Devices menu should show an option to scan a QR login code.
	QRLoginCamera bool `json:"qr_login_camera"`
	// Whether the login screen should show a QR code login option, possibly
	// as default login method ("disabled", "primary" or "secondary")
	QRLoginCode string `json:"qr_login_code"`
	// Whether clients should show an option for managing dialog filters AKA folders.
	DialogFiltersEnabled bool `json:"dialog_filters_enabled"`
	// Whether clients should actively show a tooltip, inviting the user to configure dialog filters AKA folders.
	//
	// Typically, this happens when the chat list is long enough to start getting cluttered.
	DialogFiltersTooltip bool `json:"dialog_filters_tooltip"`
	// Whether clients can invoke account.setGlobalPrivacySettings
	// with globalPrivacySettings.archive_and_mute_new_noncontact_peers = boolTrue,
	// to automatically archive and mute new incoming chats from non-contacts.
	AutoArchiveSettingAvailable bool `json:"autoarchive_setting_available"`
	// Contains a list of suggestions that should be actively shown as a tooltip to the user.
	PendingSuggestions []string `json:"pending_suggestions"`
	// Autologin token.
	//
	// See https://core.telegram.org/api/url-authorization#link-url-authorization.
	AutologinToken string `json:"autologin_token"`
	// A list of Telegram domains that support automatic login with no user confirmation.
	//
	// See https://core.telegram.org/api/url-authorization#link-url-authorization.
	AutologinDomains []string `json:"autologin_domains"`
	// A list of domains that support automatic login with manual user confirmation.
	//
	// See https://core.telegram.org/api/url-authorization#link-url-authorization.
	URLAuthDomains []string `json:"url_auth_domains"`
	// Contains a set of recommended codec parameters for round videos.
	RoundVideoEncoding RoundVideoEncoding `json:"round_video_encoding"`
	// To protect user privacy, read receipts are only stored for
	// chat_read_mark_expire_period seconds after the message was sent.
	ChatReadMarkExpirePeriod int `json:"chat_read_mark_expire_period"`
	// Per-user read receipts, fetchable using messages.getMessageReadParticipants
	// will be available in groups with less than chat_read_mark_size_threshold participants.
	ChatReadMarkSizeThreshold int `json:"chat_read_mark_size_threshold"`

	// Unparsed is map of unknown unparsed fields.
	Unparsed map[string]tg.JSONValueClass
}

func decodeStringArray(d *jx.Decoder, array *[]string) error {
	return d.Arr(func(d *jx.Decoder) error {
		v, err := d.Str()
		if err != nil {
			return errors.Wrapf(err, "decode %d", len(*array))
		}
		*array = append(*array, v)
		return nil
	})
}

// DecodeJSON decodes AppConfig.
func (e *AppConfig) DecodeJSON(d *jx.Decoder) error {
	// TODO(tdakkota): use schema-based parser
	return d.ObjBytes(func(d *jx.Decoder, key []byte) error {
		switch string(key) {
		case "test":
			v, err := d.Int()
			if err != nil {
				return errors.Wrap(err, "decode test")
			}
			e.Test = v
		case "emojies_animated_zoom":
			v, err := d.Float64()
			if err != nil {
				return errors.Wrap(err, "decode emojies_animated_zoom")
			}
			e.EmojiesAnimatedZoom = v
		case "emojies_send_dice":
			if err := decodeStringArray(d, &e.EmojiesSendDice); err != nil {
				return errors.Wrap(err, "decode emojies_send_dice")
			}
		case "emojies_send_dice_success":
			if e.EmojiesSendDiceSuccess == nil {
				e.EmojiesSendDiceSuccess = map[string]EmojiSendDiceSuccess{}
			}
			if err := d.Obj(func(d *jx.Decoder, key string) error {
				var v EmojiSendDiceSuccess
				if err := v.DecodeJSON(d); err != nil {
					return errors.Wrapf(err, "decode %q", key)
				}
				e.EmojiesSendDiceSuccess[key] = v
				return nil
			}); err != nil {
				return errors.Wrap(err, "decode emojies_send_dice_success")
			}
		case "emojies_sounds":
			if e.EmojiesSounds == nil {
				e.EmojiesSounds = map[string]EmojiSound{}
			}
			if err := d.Obj(func(d *jx.Decoder, key string) error {
				var v EmojiSound
				if err := v.DecodeJSON(d); err != nil {
					return errors.Wrapf(err, "decode %q", key)
				}
				e.EmojiesSounds[key] = v
				return nil
			}); err != nil {
				return errors.Wrap(err, "decode emojies_sounds")
			}
		case "gif_search_branding":
			v, err := d.Str()
			if err != nil {
				return errors.Wrap(err, "decode gif_search_branding")
			}
			e.GIFSearchBranding = v
		case "gif_search_emojies":
			if err := decodeStringArray(d, &e.GIFSearchEmojies); err != nil {
				return errors.Wrap(err, "decode gif_search_emojies")
			}
		case "stickers_emoji_suggest_only_api":
			v, err := d.Bool()
			if err != nil {
				return errors.Wrap(err, "decode stickers_emoji_suggest_only_api")
			}
			e.StickersEmojiSuggestOnlyAPI = v
		case "stickers_emoji_cache_time":
			v, err := d.Int()
			if err != nil {
				return errors.Wrap(err, "decode stickers_emoji_cache_time")
			}
			e.StickersEmojiCacheTime = v
		case "groupcall_video_participants_max":
			v, err := d.Int()
			if err != nil {
				return errors.Wrap(err, "decode groupcall_video_participants_max")
			}
			e.GroupCallVideoParticipantsMax = v
		case "youtube_pip":
			v, err := d.Str()
			if err != nil {
				return errors.Wrap(err, "decode youtube_pip")
			}
			e.YoutubePIP = v
		case "qr_login_camera":
			v, err := d.Bool()
			if err != nil {
				return errors.Wrap(err, "decode qr_login_camera")
			}
			e.QRLoginCamera = v
		case "qr_login_code":
			v, err := d.Str()
			if err != nil {
				return errors.Wrap(err, "decode qr_login_code")
			}
			e.QRLoginCode = v
		case "dialog_filters_enabled":
			v, err := d.Bool()
			if err != nil {
				return errors.Wrap(err, "decode dialog_filters_enabled")
			}
			e.DialogFiltersEnabled = v
		case "dialog_filters_tooltip":
			v, err := d.Bool()
			if err != nil {
				return errors.Wrap(err, "decode dialog_filters_tooltip")
			}
			e.DialogFiltersTooltip = v
		case "autoarchive_setting_available":
			v, err := d.Bool()
			if err != nil {
				return errors.Wrap(err, "decode autoarchive_setting_available")
			}
			e.AutoArchiveSettingAvailable = v
		case "pending_suggestions":
			if err := decodeStringArray(d, &e.PendingSuggestions); err != nil {
				return errors.Wrap(err, "decode pending_suggestions")
			}
		case "autologin_token":
			v, err := d.Str()
			if err != nil {
				return errors.Wrap(err, "decode autologin_token")
			}
			e.AutologinToken = v
		case "autologin_domains":
			if err := decodeStringArray(d, &e.AutologinDomains); err != nil {
				return errors.Wrap(err, "decode autologin_domains")
			}
		case "url_auth_domains":
			if err := decodeStringArray(d, &e.URLAuthDomains); err != nil {
				return errors.Wrap(err, "decode url_auth_domains")
			}
		case "round_video_encoding":
			if err := e.RoundVideoEncoding.DecodeJSON(d); err != nil {
				return errors.Wrap(err, "decode round_video_encoding")
			}
		case "chat_read_mark_expire_period":
			v, err := d.Int()
			if err != nil {
				return errors.Wrap(err, "decode chat_read_mark_expire_period")
			}
			e.ChatReadMarkExpirePeriod = v
		case "chat_read_mark_size_threshold":
			v, err := d.Int()
			if err != nil {
				return errors.Wrap(err, "decode chat_read_mark_size_threshold")
			}
			e.ChatReadMarkSizeThreshold = v
		default:
			return d.Skip()
		}

		return nil
	})
}

// DecodeJSONValue decodes AppConfig from tg.JSONValueClass.
func (e *AppConfig) DecodeJSONValue(val tg.JSONValueClass) error {
	v, ok := val.(*tg.JSONObject)
	if !ok {
		return errors.Errorf("unexpected type %T", val)
	}

	// TODO(tdakkota): decode directly
	encoder := jx.GetEncoder()
	Encode(v, encoder)

	return e.DecodeJSON(jx.DecodeBytes(encoder.Bytes()))
}
