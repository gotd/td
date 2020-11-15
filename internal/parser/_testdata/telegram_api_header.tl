int ? = Int;
long ? = Long;
double ? = Double;
string ? = String;

bytes = Bytes;

true#3fedd339 = True;

boolFalse#bc799737 = Bool;
boolTrue#997275b5 = Bool;

vector#1cb5c415 {t:Type} # [ t ] = Vector t;

error#c4b9f9bb code:int text:string = Error;

ipPort#d433ad73 ipv4:int port:int = IpPort;
ipPortSecret#37982646 ipv4:int port:int secret:bytes = IpPort;
accessPointRule#4679b65f phone_prefix_rules:string dc_id:int ips:vector<IpPort> = AccessPointRule;
help.configSimple#5a592a6c date:int expires:int rules:vector<AccessPointRule> = help.ConfigSimple;

---functions---

test.useError = Error;
test.useConfigSimple = help.ConfigSimple;
