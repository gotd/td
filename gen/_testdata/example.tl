int32 = Int32;
string ? = String;
false#bc799737 = Bool;
true#997275b5 = Bool;
bytes#e937bb82 = Bytes;

vector#1cb5c415 {t:Type} # [ t ] = Vector t;

///@description An object of this type can be returned on every function call, in case of an error
//@code Error code; subject to future changes. If the error code is 406, the error message must not be processed in any way and must not be displayed to the user
//@message Error message; subject to future changes
error code:int32 message:string temporary:Bool = Error;

//@description can be returned by functions as result.
ok = Ok;

message err:Error = Message;

sms text:string = SMS;

responseID id:int32 = Response;
responseText text:string = Response;

//@description Message
bigMessage id:int32 count:int32 targetId:int32 escape:bool summary:bool = AbstractMessage;

noMessage = AbstractMessage;

targetsMessage targets:vector<int32> = AbstractMessage;

update msg:AbstractMessage delay:int32 = Update;

getUpdatesResp updates:Vector<AbstractMessage> = GetUpdatesResp;

fieldsMessage#947225b5 flags:# escape:flags.0?Bool ttl_seconds:flags.1?int = AbstractMessage;

bytesMessage data:bytes = AbstractMessage;

//@class TextEntityType @description Represents a part of the text which must be formatted differently

//@description A mention of a user by their username
textEntityTypeMention = TextEntityType;

//@description A hashtag text, beginning with "#"
textEntityTypeHashtag = TextEntityType;

//@description A cashtag text, beginning with "$" and consisting of capital english letters (i.e. "$USD")
textEntityTypeCashtag = TextEntityType;

//@description A bot command, beginning with "/". This shouldn't be highlighted if there are no bots in the chat
textEntityTypeBotCommand = TextEntityType;

//@description An HTTP URL
textEntityTypeUrl = TextEntityType;

//@description An email address
textEntityTypeEmailAddress = TextEntityType;

//@description A phone number
textEntityTypePhoneNumber = TextEntityType;

//@description A bank card number. The getBankCardInfo method can be used to get information about the bank card
textEntityTypeBankCardNumber = TextEntityType;

//@description A bold text
textEntityTypeBold = TextEntityType;

//@description An italic text
textEntityTypeItalic = TextEntityType;

//@description An underlined text
textEntityTypeUnderline = TextEntityType;

//@description A strikethrough text
textEntityTypeStrikethrough = TextEntityType;

//@description Text that must be formatted as if inside a code HTML tag
textEntityTypeCode = TextEntityType;

//@description Text that must be formatted as if inside a pre HTML tag
textEntityTypePre = TextEntityType;

//@description Text that must be formatted as if inside pre, and code HTML tags @language Programming language of the code; as defined by the sender
textEntityTypePreCode language:string = TextEntityType;

//@description A text description shown instead of a raw URL @url HTTP or tg:// URL to be opened when the link is clicked
textEntityTypeTextUrl url:string = TextEntityType;

//@description A text shows instead of a raw mention of the user (e.g., when the user has no username) @user_id Identifier of the mentioned user
textEntityTypeMentionName user_id:int32 = TextEntityType;

//@description Represents a part of the text that needs to be formatted in some unusual way @offset Offset of the entity, in UTF-16 code units @length Length of the entity, in UTF-16 code units @type Type of the entity
textEntity offset:int32 length:int32 type:TextEntityType = TextEntity;

//@description Contains a list of text entities @entities List of text entities
textEntities entities:vector<textEntity> = TextEntities;


//@description A simple object containing a number; for testing only @value Number
testInt value:int32 = TestInt;
//@description A simple object containing a string; for testing only @value String
testString value:string = TestString;
//@description A simple object containing a sequence of bytes; for testing only @value Bytes
testBytes value:bytes = TestBytes;
//@description A simple object containing a vector of numbers; for testing only @value Vector of numbers
testVectorInt value:vector<int32> = TestVectorInt;
//@description A simple object containing a vector of objects that hold a number; for testing only @value Vector of objects
testVectorIntObject value:vector<testInt> = TestVectorIntObject;
//@description A simple object containing a vector of strings; for testing only @value Vector of strings
testVectorString value:vector<string> = TestVectorString;
//@description A simple object containing a vector of objects that hold a string; for testing only @value Vector of objects
testVectorStringObject value:vector<testString> = TestVectorStringObject;

//@description A simple object containing a vector of bytes.
testVectorBytes value:vector<bytes> = TestVectorBytes;

testVectorVector value:vector<vector<string>> = TestVectorVector;

client_DH_inner_data#6643b654 nonce:int128 server_nonce:int128 retry_id:long g_b:string = Client_DH_Inner_Data;

dcOption#18b7a10d flags:# ipv6:flags.0?true media_only:flags.1?true tcpo_only:flags.2?true cdn:flags.3?true static:flags.4?true id:int ip_address:string port:int secret:flags.10?bytes = DcOption;
config#330b4067 flags:# phonecalls_enabled:flags.1?true default_p2p_contacts:flags.3?true preload_featured_stickers:flags.4?true ignore_phone_entities:flags.5?true revoke_pm_inbox:flags.6?true blocked_mode:flags.8?true pfs_enabled:flags.13?true date:int expires:int test_mode:Bool this_dc:int dc_options:Vector<DcOption> dc_txt_domain_name:string chat_size_max:int megagroup_size_max:int forwarded_count_max:int online_update_period_ms:int offline_blur_timeout_ms:int offline_idle_timeout_ms:int online_cloud_timeout_ms:int notify_cloud_delay_ms:int notify_default_delay_ms:int push_chat_period_ms:int push_chat_limit:int saved_gifs_limit:int edit_time_limit:int revoke_time_limit:int revoke_pm_time_limit:int rating_e_decay:int stickers_recent_limit:int stickers_faved_limit:int channels_read_media_period:int tmp_sessions:flags.0?int pinned_dialogs_count_max:int pinned_infolder_count_max:int call_receive_timeout_ms:int call_ring_timeout_ms:int call_connect_timeout_ms:int call_packet_timeout_ms:int me_url_prefix:string autoupdate_url_prefix:flags.7?string gif_search_username:flags.9?string venue_search_username:flags.10?string img_search_username:flags.11?string static_maps_provider:flags.12?string caption_length_max:int message_length_max:int webfile_dc_id:int suggested_lang_code:flags.2?string lang_pack_version:flags.2?int base_lang_pack_version:flags.2?int = Config;

invokeWithLayer#da9b0d0d {X:Type} layer:int query:!X = X;

auth name:string = Auth;

authPassword name:string password:string = Auth;

user.auth foo:string = user.Auth;
user.authPassword pwd:string = user.Auth;

theme#28f1114 name:string = Theme;

account.themesNotModified#f41eb622 = account.Themes;
account.themes#7f676421 hash:int themes:Vector<Theme> = account.Themes;

---functions---

//@description check that server is live
ping id:int32 = Ok;

send msg:SMS = SMS;

sendMultipleSMS messages:vector<SMS> = Ok;

doAuth = Auth;

echoVector ids:Vector<int> = Vector<int>;
