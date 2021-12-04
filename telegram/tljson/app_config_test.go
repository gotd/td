package tljson

import (
	"testing"

	"github.com/go-faster/jx"
	"github.com/stretchr/testify/require"

	"github.com/gotd/td/tg"
)

const example = `{
    "test": 1,
    "emojies_animated_zoom": 0.625,
    "emojies_send_dice": [
        "\ud83c\udfb2",
        "\ud83c\udfaf",
        "\ud83c\udfc0",
        "\u26bd",
        "\u26bd\ufe0f",
        "\ud83c\udfb0",
        "\ud83c\udfb3"
    ],
    "emojies_send_dice_success": {
        "\ud83c\udfaf": {
            "value": 6,
            "frame_start": 62
        },
        "\ud83c\udfc0": {
            "value": 5,
            "frame_start": 110
        },
        "\u26bd": {
            "value": 5,
            "frame_start": 110
        },
        "\u26bd\ufe0f": {
            "value": 5,
            "frame_start": 110
        },
        "\ud83c\udfb0": {
            "value": 64,
            "frame_start": 110
        },
        "\ud83c\udfb3": {
            "value": 6,
            "frame_start": 110
        }
    },
    "emojies_sounds": {
        "\ud83c\udf83": {
            "id": "4956223179606458539",
            "access_hash": "-2107001400913062971",
            "file_reference_base64": "AGFhvoKbftK5O9K9RpgN1ZtgSzWy"
        },
        "\u26b0": {
            "id": "4956223179606458540",
            "access_hash": "-1498869544183595185",
            "file_reference_base64": "AGFhvoJIm8Uz0qSMIdm3AsKlK7wJ"
        },
        "\ud83e\udddf\u200d\u2642": {
            "id": "4960929110848176331",
            "access_hash": "3986395821757915468",
            "file_reference_base64": "AGFhvoLtXSSIclmvfg6ePz3KsHQF"
        },
        "\ud83e\udddf": {
            "id": "4960929110848176332",
            "access_hash": "-8929417974289765626",
            "file_reference_base64": "AGFhvoImaz5Umt4GvMUD5nocIu0W"
        },
        "\ud83e\udddf\u200d\u2640": {
            "id": "4960929110848176333",
            "access_hash": "9161696144162881753",
            "file_reference_base64": "AGFhvoIm1QZsb48xlpRfh4Mq7EMG"
        },
        "\ud83c\udf51": {
            "id": "4963180910661861548",
            "access_hash": "-7431729439735063448",
            "file_reference_base64": "AGFhvoKLrwl_WKr5LR0Jjs7o3RyT"
        },
        "\ud83c\udf8a": {
            "id": "5094064004578410732",
            "access_hash": "8518192996098758509",
            "file_reference_base64": "AGFhvoKMNffRV2J3vKED0O6d8e42"
        },
        "\ud83c\udf84": {
            "id": "5094064004578410733",
            "access_hash": "-4142643820629256996",
            "file_reference_base64": "AGFhvoJ1ulPBbXEURlTZWwJFx6xZ"
        },
        "\ud83e\uddbe": {
            "id": "5094064004578410734",
            "access_hash": "-8934384022571962340",
            "file_reference_base64": "AGFhvoL4zdMRmYv9z3L8KPaX4JQL"
        }
    },
    "gif_search_branding": "tenor",
    "gif_search_emojies": [
        "\ud83d\udc4d",
        "\ud83d\ude18",
        "\ud83d\ude0d",
        "\ud83d\ude21",
        "\ud83e\udd73",
        "\ud83d\ude02",
        "\ud83d\ude2e",
        "\ud83d\ude44",
        "\ud83d\ude0e",
        "\ud83d\udc4e"
    ],
    "stickers_emoji_suggest_only_api": false,
    "stickers_emoji_cache_time": 86400,
    "qr_login_camera": false,
    "qr_login_code": "disabled",
    "dialog_filters_enabled": true,
    "dialog_filters_tooltip": false,
    "autoarchive_setting_available": false,
    "pending_suggestions": [
        "AUTOARCHIVE_POPULAR",
        "VALIDATE_PASSWORD",
        "VALIDATE_PHONE_NUMBER",
        "NEWCOMER_TICKS"
    ],
    "autologin_token": "string",
    "autologin_domains": [
        "instantview.telegram.org",
        "translations.telegram.org",
        "contest.dev",
        "contest.com",
        "bugs.telegram.org",
        "suggestions.telegram.org",
        "themes.telegram.org"
    ],
	"youtube_pip": "abc",
	"groupcall_video_participants_max": 1234,
    "url_auth_domains": [
        "somedomain.telegram.org"
    ],
    "round_video_encoding": {
        "diameter": 384,
        "video_bitrate": 1000,
        "audio_bitrate": 64,
        "max_size": 12582912
    },
    "chat_read_mark_size_threshold": 50,
    "chat_read_mark_expire_period": 604800,
	"unknown": null
}`

func TestDecodeAppConfig(t *testing.T) {
	a := require.New(t)

	var appConfig AppConfig

	obj, err := Decode(jx.DecodeStr(example))
	a.NoError(err)
	a.NoError(appConfig.DecodeJSONValue(obj))
	a.Equal(AppConfig{
		Test:                1,
		EmojiesAnimatedZoom: 0.625000,
		EmojiesSendDice: []string{
			"🎲",
			"🎯",
			"🏀",
			"⚽",
			"⚽️",
			"🎰",
			"🎳",
		},
		EmojiesSendDiceSuccess: map[string]EmojiSendDiceSuccess{
			"⚽": {
				Value:      5,
				FrameStart: 110,
			},
			"⚽️": {
				Value:      5,
				FrameStart: 110,
			},
			"🎯": {
				Value:      6,
				FrameStart: 62,
			},
			"🎰": {
				Value:      64,
				FrameStart: 110,
			},
			"🎳": {
				Value:      6,
				FrameStart: 110,
			},
			"🏀": {
				Value:      5,
				FrameStart: 110,
			},
		},
		EmojiesSounds: map[string]EmojiSound{
			"⚰": {
				ID:         4956223179606458540,
				AccessHash: -1498869544183595185,
				FileReference: []uint8{
					0x00, 0x61, 0x61, 0xbe, 0x82, 0x48, 0x9b, 0xc5, 0x33, 0xd2, 0xa4, 0x8c, 0x21, 0xd9, 0xb7, 0x02,
					0xc2, 0xa5, 0x2b, 0xbc, 0x09,
				},
			},
			"🍑": {
				ID:         4963180910661861548,
				AccessHash: -7431729439735063448,
				FileReference: []uint8{
					0x00, 0x61, 0x61, 0xbe, 0x82, 0x8b, 0xaf, 0x09, 0x7f, 0x58, 0xaa, 0xf9, 0x2d, 0x1d, 0x09, 0x8e,
					0xce, 0xe8, 0xdd, 0x1c, 0x93,
				},
			},
			"🎃": {
				ID:         4956223179606458539,
				AccessHash: -2107001400913062971,
				FileReference: []uint8{
					0x00, 0x61, 0x61, 0xbe, 0x82, 0x9b, 0x7e, 0xd2, 0xb9, 0x3b, 0xd2, 0xbd, 0x46, 0x98, 0x0d, 0xd5,
					0x9b, 0x60, 0x4b, 0x35, 0xb2,
				},
			},
			"🎄": {
				ID:         5094064004578410733,
				AccessHash: -4142643820629256996,
				FileReference: []uint8{
					0x00, 0x61, 0x61, 0xbe, 0x82, 0x75, 0xba, 0x53, 0xc1, 0x6d, 0x71, 0x14, 0x46, 0x54, 0xd9, 0x5b,
					0x02, 0x45, 0xc7, 0xac, 0x59,
				},
			},
			"🎊": {
				ID:         5094064004578410732,
				AccessHash: 8518192996098758509,
				FileReference: []uint8{
					0x00, 0x61, 0x61, 0xbe, 0x82, 0x8c, 0x35, 0xf7, 0xd1, 0x57, 0x62, 0x77, 0xbc, 0xa1, 0x03, 0xd0,
					0xee, 0x9d, 0xf1, 0xee, 0x36,
				},
			},
			"🦾": {
				ID:         5094064004578410734,
				AccessHash: -8934384022571962340,
				FileReference: []uint8{
					0x00, 0x61, 0x61, 0xbe, 0x82, 0xf8, 0xcd, 0xd3, 0x11, 0x99, 0x8b, 0xfd, 0xcf, 0x72, 0xfc, 0x28,
					0xf6, 0x97, 0xe0, 0x94, 0x0b,
				},
			},
			"🧟": {
				ID:         4960929110848176332,
				AccessHash: -8929417974289765626,
				FileReference: []uint8{
					0x00, 0x61, 0x61, 0xbe, 0x82, 0x26, 0x6b, 0x3e, 0x54, 0x9a, 0xde, 0x06, 0xbc, 0xc5, 0x03, 0xe6,
					0x7a, 0x1c, 0x22, 0xed, 0x16,
				},
			},
			"🧟\u200d♀": {
				ID:         4960929110848176333,
				AccessHash: 9161696144162881753,
				FileReference: []uint8{
					0x00, 0x61, 0x61, 0xbe, 0x82, 0x26, 0xd5, 0x06, 0x6c, 0x6f, 0x8f, 0x31, 0x96, 0x94, 0x5f, 0x87,
					0x83, 0x2a, 0xec, 0x43, 0x06,
				},
			},
			"🧟\u200d♂": {
				ID:         4960929110848176331,
				AccessHash: 3986395821757915468,
				FileReference: []uint8{
					0x00, 0x61, 0x61, 0xbe, 0x82, 0xed, 0x5d, 0x24, 0x88, 0x72, 0x59, 0xaf, 0x7e, 0x0e, 0x9e, 0x3f,
					0x3d, 0xca, 0xb0, 0x74, 0x05,
				},
			},
		},
		GIFSearchBranding: "tenor",
		GIFSearchEmojies: []string{
			"👍",
			"😘",
			"😍",
			"😡",
			"🥳",
			"😂",
			"😮",
			"🙄",
			"😎",
			"👎",
		},
		StickersEmojiSuggestOnlyAPI:   false,
		StickersEmojiCacheTime:        86400,
		GroupCallVideoParticipantsMax: 1234,
		YoutubePIP:                    "abc",
		QRLoginCamera:                 false,
		QRLoginCode:                   "disabled",
		DialogFiltersEnabled:          true,
		DialogFiltersTooltip:          false,
		AutoArchiveSettingAvailable:   false,
		PendingSuggestions: []string{
			"AUTOARCHIVE_POPULAR",
			"VALIDATE_PASSWORD",
			"VALIDATE_PHONE_NUMBER",
			"NEWCOMER_TICKS",
		},
		AutologinToken: "string",
		AutologinDomains: []string{
			"instantview.telegram.org",
			"translations.telegram.org",
			"contest.dev",
			"contest.com",
			"bugs.telegram.org",
			"suggestions.telegram.org",
			"themes.telegram.org",
		},
		URLAuthDomains: []string{
			"somedomain.telegram.org",
		},
		RoundVideoEncoding: RoundVideoEncoding{
			Diameter:     384,
			VideoBitrate: 1000,
			AudioBitrate: 64,
			MaxSize:      12582912,
		},
		ChatReadMarkExpirePeriod:  604800,
		ChatReadMarkSizeThreshold: 50,
		Unparsed: map[string]tg.JSONValueClass{
			"unknown": &tg.JSONNull{},
		},
	}, appConfig)
}
