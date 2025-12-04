package telegram

import (
	"context"
	"strings"
	"sync"

	"github.com/go-faster/errors"
	"github.com/go-faster/jx"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/tdjson"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"
)

var (
	tdlibToTelegramMap     map[string]string
	tdlibToTelegramMapOnce sync.Once
)

// buildTDLibToTelegramMap builds reverse mapping from TDLib method names
// (e.g., "sendMessage") to Telegram API names (e.g., "messages.sendMessage").
func buildTDLibToTelegramMap() map[string]string {
	tdlibToTelegramMapOnce.Do(func() {
		namesMap := tg.NamesMap()
		tdlibToTelegramMap = make(map[string]string)

		// Build reverse mapping: for each entry in NamesMap that looks like a method request,
		// create mapping from TDLib name (without namespace) to full name
		for fullName := range namesMap {
			// Only process method requests (they contain a dot and end with "Request" conceptually)
			// In NamesMap, request methods are stored as "namespace.methodName"
			if !strings.Contains(fullName, ".") {
				continue
			}

			// Extract method name without namespace
			parts := strings.SplitN(fullName, ".", 2)
			if len(parts) != 2 {
				continue
			}
			methodName := parts[1]

			// Skip result types and constructors (they don't have "Request" in their name pattern)
			// Request methods in NamesMap are like "messages.sendMessage", "auth.sendCode", etc.
			// We want to map "sendMessage" -> "messages.sendMessage"
			// Skip types that are clearly not methods (like "auth.authorization", "messages.messages")
			if strings.HasSuffix(methodName, "TypeID") ||
				strings.HasSuffix(methodName, "Class") {
				continue
			}

			// Check if the type ID corresponds to a Request type by checking the constructors map
			// Request types in NamesMap map to RequestTypeID constants
			// We'll include it in the mapping - if it's not a request, it will fail later during invocation
			// This is simpler than trying to detect all request types upfront
			tdlibToTelegramMap[methodName] = fullName
		}
	})

	return tdlibToTelegramMap
}

// InvokeRawJSON invokes MTProto method using pure TDLib JSON format.
// It accepts JSON with @type field in TDLib format (e.g., "sendMessage")
// and returns JSON response in TDLib format.
//
// The JSON structure, keys and values exactly match TDLib standard JSON.
// For example, use "sendMessage" not "messages.sendMessage" in @type field.
func (c *Client) InvokeRawJSON(ctx context.Context, jsonRequest string) (jsonResponse string, err error) {
	// Parse JSON to extract @type
	decoder := jx.DecodeStr(jsonRequest)
	tdDecoder := tdjson.Decoder{Decoder: decoder}

	typeID, err := tdDecoder.FindTypeID()
	if err != nil {
		return "", errors.Wrap(err, "find @type field")
	}

	// Map TDLib method name to Telegram API name
	tdlibMap := buildTDLibToTelegramMap()
	telegramName, ok := tdlibMap[typeID]
	if !ok {
		// Try direct lookup in case it's already in full format
		namesMap := tg.NamesMap()
		if _, exists := namesMap[typeID]; exists {
			telegramName = typeID
		} else {
			return "", errors.Errorf("unknown method: %q", typeID)
		}
	}

	// Lookup type ID
	namesMap := tg.NamesMap()
	typeIDValue, ok := namesMap[telegramName]
	if !ok {
		return "", errors.Errorf("type not found: %q", telegramName)
	}

	// Create instance
	constructorsMap := tg.TypesConstructorMap()
	constructor, ok := constructorsMap[typeIDValue]
	if !ok {
		return "", errors.Errorf("constructor not found for type: %q (id: 0x%x)", telegramName, typeIDValue)
	}

	instance := constructor()
	if instance == nil {
		return "", errors.Errorf("failed to create instance for type: %q", telegramName)
	}

	// Check if instance supports TDLib JSON decoding
	decoderObj, ok := instance.(tdjson.TDLibDecoder)
	if !ok {
		return "", errors.Errorf("type %q does not support TDLib JSON decoding", telegramName)
	}

	// Decode JSON into instance (create new decoder for full decode)
	tdDecoder2 := jx.DecodeStr(jsonRequest)
	tdDecoderWrapper := tdjson.Decoder{Decoder: tdDecoder2}
	if err := decoderObj.DecodeTDLibJSON(tdDecoderWrapper); err != nil {
		return "", errors.Wrapf(err, "decode TDLib JSON for %q", telegramName)
	}

	// Ensure instance implements bin.Encoder
	encoder, ok := instance.(bin.Encoder)
	if !ok {
		return "", errors.Errorf("type %q does not implement bin.Encoder", telegramName)
	}

	// Invoke via existing Invoke method with a decoder that can handle any response type
	responseDecoder := &rawJSONResponseDecoder{}
	if err := c.Invoke(ctx, encoder, responseDecoder); err != nil {
		// Check if it's an RPC error and convert to TDLib JSON error format
		if rpcErr, ok := tgerr.As(err); ok {
			return encodeTDLibError(rpcErr), nil
		}
		return "", errors.Wrap(err, "invoke")
	}

	response := responseDecoder.result
	if response == nil {
		return "", errors.New("empty response")
	}

	// Encode response to TDLib JSON
	encoderObj, ok := response.(tdjson.TDLibEncoder)
	if !ok {
		return "", errors.Errorf("response type does not support TDLib JSON encoding")
	}

	var jsonBuilder jx.Writer
	tdEncoder := tdjson.Encoder{Writer: &jsonBuilder}
	if err := encoderObj.EncodeTDLibJSON(tdEncoder); err != nil {
		return "", errors.Wrap(err, "encode response to TDLib JSON")
	}

	return jsonBuilder.String(), nil
}

// rawJSONResponseDecoder is a decoder that can decode any response type
// and store it as bin.Object for later JSON encoding.
type rawJSONResponseDecoder struct {
	result bin.Object
}

func (d *rawJSONResponseDecoder) Decode(b *bin.Buffer) error {
	// Peek at the type ID to determine what to decode
	id, err := b.PeekID()
	if err != nil {
		return errors.Wrap(err, "peek response type ID")
	}

	// Use TypesConstructorMap to create appropriate instance
	constructorsMap := tg.TypesConstructorMap()
	constructor, ok := constructorsMap[id]
	if !ok {
		return errors.Errorf("unknown response type: 0x%x", id)
	}

	instance := constructor()
	if instance == nil {
		return errors.Errorf("failed to create response instance for type: 0x%x", id)
	}

	// Decode into instance
	decoder, ok := instance.(bin.Decoder)
	if !ok {
		return errors.Errorf("response type does not implement bin.Decoder")
	}

	if err := decoder.Decode(b); err != nil {
		return errors.Wrap(err, "decode response")
	}

	// Store as bin.Object
	if obj, ok := instance.(bin.Object); ok {
		d.result = obj
	} else {
		return errors.New("response is not a bin.Object")
	}

	return nil
}

// encodeTDLibError encodes tgerr.Error to TDLib JSON error format.
func encodeTDLibError(err *tgerr.Error) string {
	// TDLib error format: {"@type":"error","code":420,"message":"FLOOD_WAIT_3"}
	var jsonBuilder jx.Writer
	jsonBuilder.ObjStart()
	jsonBuilder.FieldStart("@type")
	jsonBuilder.Str("error")
	jsonBuilder.Comma()
	jsonBuilder.FieldStart("code")
	jsonBuilder.Int(err.Code)
	jsonBuilder.Comma()
	jsonBuilder.FieldStart("message")
	jsonBuilder.Str(err.Message)
	jsonBuilder.ObjEnd()
	return jsonBuilder.String()
}
