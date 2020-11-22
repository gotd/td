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

getUpdatesResp updates:vector<AbstractMessage> = GetUpdatesResp;

fieldsMessage#947225b5 flags:# escape:flags.0?true ttl_seconds:flags.1?int = AbstractMessage;

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

---functions---

//@description check that server is live
ping id:int32 = Ok;

send msg:SMS = SMS;

sendMultipleSMS messages:vector<SMS> = Ok;
