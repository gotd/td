boolFalse#bc799737 = Bool;
boolTrue#997275b5 = Bool;
true#3fedd339 = True;
vector#1cb5c415 {t:Type} # [ t ] = Vector t;

// Minimal Telegram-like schema with the types used by generated handlers.

userEmpty#d3bc4b7a id:long = User;
user#b1b8cc83 flags:# self:flags.10?true id:long access_hash:flags.0?long = User;

chatEmpty#29562865 id:long = Chat;
chat#41cbf256 id:long title:string = Chat;
channel#7ad15240 flags:# min:flags.12?true id:long access_hash:flags.13?long title:string = Chat;

updateNewMessage#1f2b0afd text:string pts:int pts_count:int = Update;
updateDeleteMessages#a20db0e5 messages:Vector<int> pts:int pts_count:int = Update;

updatesTooLong#e317af7e = Updates;
updateShort#78d4dec1 update:Update date:int = Updates;
updatesCombined#725b04c3 updates:Vector<Update> users:Vector<User> chats:Vector<Chat> date:int seq_start:int seq:int = Updates;
updates#74ae4240 updates:Vector<Update> users:Vector<User> chats:Vector<Chat> date:int seq:int = Updates;
