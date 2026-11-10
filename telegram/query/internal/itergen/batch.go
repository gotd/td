package main

// fallbackBatchSize is the default page size of methods missing from
// defaultBatchSizes. It is the server-side limit of most paginated methods.
const fallbackBatchSize = 100

// defaultBatchSizes maps tg.Client method names to the default iterator page size.
//
// Values follow the official clients and never exceed the server-side limit:
// iterators stop when a page is shorter than the requested limit, so a default
// above the server cap would silently truncate iteration.
var defaultBatchSizes = map[string]int{
	// tdlib DialogManager::MAX_GET_DIALOGS = 100 (server-side limit).
	"MessagesGetDialogs": 100,
	// tdlib MessagesManager::MAX_GET_HISTORY = 100 (server-side limit).
	"MessagesGetHistory":                100,
	"MessagesGetReplies":                100,
	"MessagesGetSavedHistory":           100,
	"MessagesGetUnreadMentions":         100,
	"MessagesGetUnreadReactions":        100,
	"MessagesGetUnreadPollVotes":        100,
	"MessagesGetPersonalChannelHistory": 100,
	// tdlib MessageQueryManager::MAX_SEARCH_MESSAGES = 100 (server-side limit).
	"MessagesSearch":             100,
	"MessagesSearchGlobal":       100,
	"MessagesSearchSentMedia":    100,
	"MessagesGetRecentLocations": 100,
	"ChannelsSearchPosts":        100,
	// tdlib DialogParticipantManager::MAX_GET_CHANNEL_PARTICIPANTS = 200 (server-side limit).
	"ChannelsGetParticipants": 200,
	// tdlib UserManager::MAX_GET_PROFILE_PHOTOS = 100; tdesktop kSharedMediaLimit = 100.
	"PhotosGetUserPhotos": 100,
	// tdesktop api_blocked_peers.cpp kBlockedPerPage = 40.
	"ContactsGetBlocked": 40,
	// tdesktop stickers_list_widget.cpp kOfficialLoadLimit = 40.
	"MessagesGetOldFeaturedStickers": 40,
}

// defaultBatchSize returns the default iterator page size of the given tg.Client method.
func defaultBatchSize(method string) int {
	if v, ok := defaultBatchSizes[method]; ok {
		return v
	}
	return fallbackBatchSize
}
