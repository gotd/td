// See https://github.com/tdlib/td/blob/70b71d569b81c66b25f8f77d9536b1d41fbf091d/td/generate/scheme/telegram_api.tl#L17-L20

ipPort#d433ad73 ipv4:int port:int = IpPort;
ipPortSecret#37982646 ipv4:int port:int secret:bytes = IpPort;
accessPointRule#4679b65f phone_prefix_rules:string dc_id:int ips:vector<IpPort> = AccessPointRule;
help.configSimple#5a592a6c date:int expires:int rules:vector<AccessPointRule> = help.ConfigSimple;
