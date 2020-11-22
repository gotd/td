int32 = Int32;
string ? = String;
boolFalse#bc799737 = Bool;
boolTrue#997275b5 = Bool;

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

fieldsMessage#947225b5 flags:# escape:flags.0?bool ttl_seconds:flags.1?int = AbstractMessage;

---functions---

//@description check that server is live
ping id:int32 = Ok;

send msg:SMS = SMS;
