package proto

import (
	"fmt"

	"github.com/gotd/td/bin"
)

// initConnection#c1cd5ea9 {X:Type} flags:# api_id:int device_model:string system_version:string
// app_version:string system_lang_code:string lang_pack:string lang_code:string
// proxy:flags.0?InputClientProxy params:flags.1?JSONValue query:!X = X;

// InitConnection is initConnection#c1cd5ea9 function.
type InitConnection struct {
	Flags          bin.Fields
	ID             int
	DeviceModel    string
	SystemVersion  string
	AppVersion     string
	SystemLangCode string
	LangPack       string
	LangCode       string

	Query TType
}

// InitConnectionID is TL type id of initConnection#c1cd5ea9.
const InitConnectionID = 0xc1cd5ea9

// Encode implements bin.Encoder.
func (i InitConnection) Encode(b *bin.Buffer) error {
	b.PutID(InitConnectionID)
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

// Decode implements bin.Decoder.
func (i InitConnection) Decode(b *bin.Buffer) (err error) {
	if err := b.ConsumeID(InitConnectionID); err != nil {
		return fmt.Errorf("unable to decode initConnection#c1cd5ea9: %w", err)
	}

	if err := i.Flags.Decode(b); err != nil {
		return fmt.Errorf("unable to decode initConnection#c1cd5ea9: %w", err)
	}

	i.ID, err = b.Int()
	if err != nil {
		return fmt.Errorf("unable to decode initConnection#c1cd5ea9: %w", err)
	}

	i.DeviceModel, err = b.String()
	if err != nil {
		return fmt.Errorf("unable to decode initConnection#c1cd5ea9: %w", err)
	}

	i.SystemVersion, err = b.String()
	if err != nil {
		return fmt.Errorf("unable to decode initConnection#c1cd5ea9: %w", err)
	}

	i.AppVersion, err = b.String()
	if err != nil {
		return fmt.Errorf("unable to decode initConnection#c1cd5ea9: %w", err)
	}

	i.SystemLangCode, err = b.String()
	if err != nil {
		return fmt.Errorf("unable to decode initConnection#c1cd5ea9: %w", err)
	}

	i.LangPack, err = b.String()
	if err != nil {
		return fmt.Errorf("unable to decode initConnection#c1cd5ea9: %w", err)
	}

	i.LangCode, err = b.String()
	if err != nil {
		return fmt.Errorf("unable to decode initConnection#c1cd5ea9: %w", err)
	}

	if err := i.Query.Decode(b); err != nil {
		return fmt.Errorf("unable to decode initConnection#c1cd5ea9: %w", err)
	}
	return nil
}
