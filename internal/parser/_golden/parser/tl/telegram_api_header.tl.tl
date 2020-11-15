int#a8509bda ? = Int;

long#22076cba ? = Long;

double#2210c154 ? = Double;

string#b5286e24 ? = String;

bytes#e937bb82 = Bytes;

true#3fedd339 = True;

boolFalse#bc799737 = Bool;

boolTrue#997275b5 = Bool;

error#c4b9f9bb code:int text:string = Error;

ipPort#d433ad73 ipv4:int port:int = IpPort;

ipPortSecret#37982646 ipv4:int port:int secret:bytes = IpPort;

accessPointRule#4679b65f phone_prefix_rules:string dc_id:int ips:vector<IpPort> = AccessPointRule;

help.configSimple#5a592a6c date:int expires:int rules:vector<AccessPointRule> = help.ConfigSimple;


---functions---

test.useError#ee75af01 = Error;

test.useConfigSimple#f9b7b23d = help.ConfigSimple;

