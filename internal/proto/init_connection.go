package proto

import "github.com/gotd/td/bin"

// initConnection#c1cd5ea9 {X:Type} flags:# api_id:int device_model:string system_version:string
// app_version:string system_lang_code:string lang_pack:string lang_code:string
// proxy:flags.0?InputClientProxy params:flags.1?JSONValue query:!X = X;

type InitConnection struct {
	Flags          bin.Fields
	ID             int
	DeviceModel    string
	SystemVersion  string
	AppVersion     string
	SystemLangCode string
	LangPack       string
	LangCode       string

	Query bin.Encoder
}

func (i InitConnection) Encode(b *bin.Buffer) error {
	b.PutID(0xc1cd5ea9)
	if err := i.Flags.Encode(b); err != nil {
		return err
	}

	b.PutInt(i.ID)
	b.PutString(i.DeviceModel)
	b.PutString(i.SystemVersion)
	b.PutString(i.AppVersion)
	b.PutString(i.SystemLangCode)
	b.PutString(i.LangPack)
	b.PutString(i.LangCode)

	return i.Query.Encode(b)
}
