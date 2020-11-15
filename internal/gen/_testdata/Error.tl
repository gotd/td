int32 = Int32;
string ? = String;
boolFalse#bc799737 = Bool;
boolTrue#997275b5 = Bool;

///@description An object of this type can be returned on every function call, in case of an error
//@code Error code; subject to future changes. If the error code is 406, the error message must not be processed in any way and must not be displayed to the user
//@message Error message; subject to future changes
error code:int32 message:string temporary:Bool = Error;

//@description can be returned by functions as result.
ok = Ok;

message err:Error = Message;

---functions---

//@description check that server is live
ping id:int32 = Ok;
