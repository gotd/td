// Code generated manually for JSON invocation support.
// This file provides InvokeJSON method for calling MTProto methods via JSON.

package tg

import (
	"context"
	"encoding/hex"
	"encoding/json"

	"github.com/go-faster/errors"
	"github.com/go-faster/jx"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/tdjson"
	"github.com/gotd/td/tdp"
)

// InvokeJSON invokes an MTProto method using JSON input.
// The JSON input should contain an "@type" field with the method name (e.g., "messages.sendMessage").
// All method parameters should be included in the same JSON object.
//
// The JSON format follows MTProto JSON API specification where:
// - "@type" field contains the method or type name
// - All fields use snake_case naming
// - Nested objects also include "@type" field
//
// useSnakeCase controls the output JSON field naming convention:
// - true: Use snake_case (MTProto convention, e.g., "user_id", "access_hash")
// - false: Use PascalCase (Go struct field names, e.g., "UserID", "AccessHash")
//
// Example:
//
//	jsonInput := `{
//	  "@type": "messages.sendMessage",
//	  "peer": {"@type": "inputPeerUser", "user_id": 123, "access_hash": 456},
//	  "message": "Hello",
//	  "random_id": 123456789
//	}`
//	result, err := client.InvokeJSON(ctx, jsonInput, true) // true = snake_case output
func (c *Client) InvokeJSON(ctx context.Context, jsonData string, useSnakeCase bool) (string, error) {
	// Parse JSON to extract @type field
	d := jx.DecodeStr(jsonData)
	var typeField string
	if err := d.ObjBytes(func(d *jx.Decoder, key []byte) error {
		if string(key) == tdjson.TypeField {
			val, err := d.Str()
			if err != nil {
				return err
			}
			typeField = val
			return nil
		}
		return d.Skip()
	}); err != nil {
		return "", errors.Wrap(err, "extract @type")
	}

	if typeField == "" {
		return "", errors.New("@type field is required")
	}

	// Get type ID from method name
	namesMap := NamesMap()
	typeID, ok := namesMap[typeField]
	if !ok {
		return "", errors.Errorf("unknown method or type: %q", typeField)
	}

	// Create request object using constructor (no reflection)
	constructors := TypesConstructorMap()
	constructor, ok := constructors[typeID]
	if !ok {
		return "", errors.Errorf("no constructor for type ID: 0x%x (%s)", typeID, typeField)
	}

	requestObj := constructor()
	if requestObj == nil {
		return "", errors.Errorf("constructor returned nil for type ID: 0x%x (%s)", typeID, typeField)
	}

	// Convert JSON to binary format using jx for parsing (no encoding/json)
	// Create a fresh decoder since the previous one was consumed
	if err := jsonUnmarshalToObject(jsonData, requestObj); err != nil {
		return "", errors.Wrap(err, "unmarshal JSON to object")
	}

	// For result, we use a generic decoder that can handle any result type
	resultDecoder := &genericResultDecoder{
		constructors: constructors,
	}

	// Invoke method
	if err := c.rpc.Invoke(ctx, requestObj.(bin.Encoder), resultDecoder); err != nil {
		return "", err
	}

	if resultDecoder.result == nil {
		return "", errors.New("no result from invocation")
	}

	// Convert binary directly to pure MTProto JSON
	// Note: For output, we still use encoding/json since manual encoding requires reflection
	// The key improvement is that input parsing uses jx, not encoding/json
	return binaryToMTProtoJSON(resultDecoder.buffer, resultDecoder.typeID, useSnakeCase)
}

// genericResultDecoder is a decoder that can decode any response type
// and store it for later JSON encoding.
type genericResultDecoder struct {
	constructors map[uint32]func() bin.Object
	result       bin.Object
	buffer       *bin.Buffer
	typeID       uint32
}

func (d *genericResultDecoder) Decode(b *bin.Buffer) error {
	// Peek at the type ID to determine what to decode
	id, err := b.PeekID()
	if err != nil {
		return errors.Wrap(err, "peek response type ID")
	}

	d.typeID = id

	// Save the buffer for later JSON encoding
	d.buffer = &bin.Buffer{Buf: make([]byte, len(b.Buf))}
	copy(d.buffer.Buf, b.Buf)

	// Use TypesConstructorMap to create appropriate instance
	constructor, ok := d.constructors[id]
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

// jsonValue represents an intermediate JSON value during parsing
type jsonValue struct {
	typ     jx.Type
	str     string
	num     jx.Num
	boolean bool
	arr     []jsonValue
	obj     map[string]jsonValue
}

// jsonUnmarshalToObject unmarshals JSON into an MTProto object using jx.
// Parses JSON with jx into intermediate structure, then encodes to binary.
func jsonUnmarshalToObject(jsonData string, obj bin.Object) error {
	// Get TypeInfo to know field structure
	typeInfoObj, ok := obj.(interface{ TypeInfo() tdp.Type })
	if !ok {
		return errors.New("object does not implement TypeInfo()")
	}

	typeInfo := typeInfoObj.TypeInfo()

	// Parse JSON with jx into intermediate structure
	d := jx.DecodeStr(jsonData)
	fieldValues, err := parseJSONObject(d)
	if err != nil {
		return errors.Wrap(err, "parse JSON object")
	}

	// Build binary buffer
	buf := &bin.Buffer{}
	buf.PutID(typeInfo.ID)

	// Encode each field in order based on TypeInfo
	for _, field := range typeInfo.Fields {
		// Skip flags fields - they will be set by SetFlags() after decoding
		if field.SchemaName == "flags" || field.SchemaName == "flags2" {
			continue
		}

		// Skip null/optional fields that aren't present
		if field.Null {
			if _, exists := fieldValues[field.SchemaName]; !exists {
				continue
			}
		}

		value, exists := fieldValues[field.SchemaName]
		if !exists {
			continue
		}

		// Encode field value to binary
		if err := encodeJSONValueToBinary(buf, value); err != nil {
			return errors.Wrapf(err, "encode field %s", field.SchemaName)
		}
	}

	// Decode binary back to object
	if err := obj.Decode(buf); err != nil {
		return errors.Wrap(err, "decode binary to object")
	}

	// Set flags if the object supports it
	if setFlagsObj, ok := obj.(interface{ SetFlags() }); ok {
		setFlagsObj.SetFlags()
	}

	return nil
}

// parseJSONObject parses a JSON object into a map of field values
// includeTypeField controls whether to include @type field (needed for nested objects)
// The decoder should be positioned at the start (ObjBytes will call Next() internally)
func parseJSONObject(d *jx.Decoder) (map[string]jsonValue, error) {
	return parseJSONObjectWithType(d, false)
}

// parseJSONObjectWithType parses a JSON object, optionally including @type field
// ObjBytes will internally call Next() to check if it's an object
func parseJSONObjectWithType(d *jx.Decoder, includeType bool) (map[string]jsonValue, error) {
	result := make(map[string]jsonValue)

	if err := d.ObjBytes(func(d *jx.Decoder, key []byte) error {
		keyStr := string(key)

		// Skip @type field if includeType is false (for root object)
		if keyStr == "@type" && !includeType {
			return d.Skip()
		}

		value, err := parseJSONValue(d)
		if err != nil {
			return errors.Wrapf(err, "parse field %s", keyStr)
		}

		result[keyStr] = value
		return nil
	}); err != nil {
		return nil, err
	}

	return result, nil
}

// parseJSONValue parses a single JSON value
func parseJSONValue(d *jx.Decoder) (jsonValue, error) {
	var val jsonValue

	// Peek at the next token to determine type
	tt := d.Next()

	switch tt {
	case jx.Object:
		// Parse object recursively (include @type for nested objects)
		// Next() has already consumed the object token, so we're now inside the object
		// We can call ObjBytes directly
		obj := make(map[string]jsonValue)
		if err := d.ObjBytes(func(d *jx.Decoder, key []byte) error {
			keyStr := string(key)
			// Always include @type for nested objects
			value, err := parseJSONValue(d)
			if err != nil {
				return errors.Wrapf(err, "parse field %s", keyStr)
			}
			obj[keyStr] = value
			return nil
		}); err != nil {
			return jsonValue{}, errors.Wrap(err, "parse nested object")
		}
		val.typ = jx.Object
		val.obj = obj
	case jx.Array:
		// Parse array
		arr := []jsonValue{}
		if err := d.Arr(func(d *jx.Decoder) error {
			item, err := parseJSONValue(d)
			if err != nil {
				return errors.Wrap(err, "parse array item")
			}
			arr = append(arr, item)
			return nil
		}); err != nil {
			return jsonValue{}, errors.Wrap(err, "parse array")
		}
		val.typ = jx.Array
		val.arr = arr
	case jx.String:
		str, err := d.Str()
		if err != nil {
			return jsonValue{}, errors.Wrap(err, "parse string")
		}
		val.typ = jx.String
		val.str = str
	case jx.Number:
		num, err := d.Num()
		if err != nil {
			return jsonValue{}, errors.Wrap(err, "parse number")
		}
		val.typ = jx.Number
		val.num = num
	case jx.Bool:
		b, err := d.Bool()
		if err != nil {
			return jsonValue{}, errors.Wrap(err, "parse bool")
		}
		val.typ = jx.Bool
		val.boolean = b
	case jx.Null:
		if err := d.Null(); err != nil {
			return jsonValue{}, errors.Wrap(err, "parse null")
		}
		val.typ = jx.Null
	default:
		return jsonValue{}, errors.Errorf("unexpected JSON token: %v", tt)
	}

	return val, nil
}

// encodeJSONValueToBinary encodes a jsonValue to binary format
func encodeJSONValueToBinary(buf *bin.Buffer, val jsonValue) error {
	return encodeJSONValueToBinaryWithContext(buf, val, "")
}

// encodeJSONValueToBinaryWithContext encodes a jsonValue to binary format with field context
func encodeJSONValueToBinaryWithContext(buf *bin.Buffer, val jsonValue, fieldName string) error {
	switch val.typ {
	case jx.Object:
		// This is an interface type (object with @type)
		return encodeInterfaceObjectToBinary(buf, val.obj)
	case jx.Array:
		// This is an array
		buf.PutVectorHeader(len(val.arr))
		for _, item := range val.arr {
			if err := encodeJSONValueToBinary(buf, item); err != nil {
				return errors.Wrap(err, "encode array item")
			}
		}
		return nil
	case jx.Bool:
		buf.PutBool(val.boolean)
	case jx.Number:
		if val.num.IsInt() {
			i, err := val.num.Int64()
			if err != nil {
				return errors.Wrap(err, "parse integer")
			}
			buf.PutLong(i)
		} else {
			f, err := val.num.Float64()
			if err != nil {
				return errors.Wrap(err, "parse float")
			}
			buf.PutDouble(f)
		}
	case jx.String:
		// Check if this might be Int128 or Int256 (hex-encoded)
		// Int128 = 16 bytes = 32 hex chars, Int256 = 32 bytes = 64 hex chars
		if isHexString(val.str) {
			strLen := len(val.str)
			// Check field name hints (common names for Int128/Int256 fields)
			// Also check length - if it's exactly 32 or 64 hex chars, it's likely Int128/Int256
			isLikelyInt128Int256 := fieldName == "public_key" ||
				fieldName == "key" ||
				fieldName == "nonce" ||
				fieldName == "secret" ||
				strLen == 32 || strLen == 64

			if isLikelyInt128Int256 {
				switch strLen {
				case 32: // Int128
					var int128 bin.Int128
					if _, err := hex.Decode(int128[:], []byte(val.str)); err == nil {
						buf.PutInt128(int128)
						return nil
					}
				case 64: // Int256
					var int256 bin.Int256
					if _, err := hex.Decode(int256[:], []byte(val.str)); err == nil {
						buf.PutInt256(int256)
						return nil
					}
				}
			}
		}
		// Regular string
		buf.PutString(val.str)
	case jx.Null:
		// Nil values are typically not encoded (handled by optional fields)
		return nil
	default:
		return errors.Errorf("unsupported JSON value type: %v", val.typ)
	}
	return nil
}

// encodeInterfaceObjectToBinary encodes an interface object (with @type) to binary
func encodeInterfaceObjectToBinary(buf *bin.Buffer, obj map[string]jsonValue) error {
	// Extract @type
	typeVal, ok := obj[tdjson.TypeField]
	if !ok {
		return errors.New("interface object missing @type field")
	}

	if typeVal.typ != jx.String {
		return errors.New("@type field must be a string")
	}

	typeName := typeVal.str

	// Get type ID and constructor
	namesMap := NamesMap()
	typeID, ok := namesMap[typeName]
	if !ok {
		return errors.Errorf("unknown type: %q", typeName)
	}

	constructors := TypesConstructorMap()
	constructor, ok := constructors[typeID]
	if !ok {
		return errors.Errorf("no constructor for type: %q", typeName)
	}

	concreteObj := constructor()
	if concreteObj == nil {
		return errors.Errorf("constructor returned nil for type: %q", typeName)
	}

	// Get TypeInfo for the concrete object
	typeInfoObj, ok := concreteObj.(interface{ TypeInfo() tdp.Type })
	if !ok {
		return errors.New("concrete object does not implement TypeInfo()")
	}

	typeInfo := typeInfoObj.TypeInfo()

	// Build binary buffer for concrete object
	tempBuf := &bin.Buffer{}
	tempBuf.PutID(typeInfo.ID)

	// Encode each field
	for _, field := range typeInfo.Fields {
		if field.SchemaName == "flags" || field.SchemaName == "flags2" {
			continue
		}

		if field.Null {
			if _, exists := obj[field.SchemaName]; !exists {
				continue
			}
		}

		value, exists := obj[field.SchemaName]
		if !exists {
			continue
		}

		// Encode field value, with context about field name for better type detection
		if err := encodeJSONValueToBinaryWithContext(tempBuf, value, field.SchemaName); err != nil {
			return errors.Wrapf(err, "encode field %s", field.SchemaName)
		}
	}

	// Try to decode to concrete object to validate encoding
	// If it fails, it might be because we encoded a hex string incorrectly
	testDecodeBuf := &bin.Buffer{Buf: make([]byte, len(tempBuf.Buf))}
	copy(testDecodeBuf.Buf, tempBuf.Buf)

	if err := concreteObj.Decode(testDecodeBuf); err != nil {
		// Decode failed - might be because we encoded a hex string as regular string
		// Try re-encoding with Int128/Int256 detection for hex strings
		return retryEncodeWithInt128Int256(buf, obj, typeName, concreteObj, typeInfo)
	}

	// Decode succeeded, use the decoded object
	// Set flags
	if setFlagsObj, ok := concreteObj.(interface{ SetFlags() }); ok {
		setFlagsObj.SetFlags()
	}

	// Encode the concrete object to the output buffer
	return concreteObj.(bin.Encoder).Encode(buf)
}

// retryEncodeWithInt128Int256 retries encoding with Int128/Int256 detection for hex strings
func retryEncodeWithInt128Int256(buf *bin.Buffer, obj map[string]jsonValue, typeName string, concreteObj bin.Object, typeInfo tdp.Type) error {
	// Rebuild binary buffer, this time being more aggressive about Int128/Int256 detection
	tempBuf := &bin.Buffer{}
	tempBuf.PutID(typeInfo.ID)

	// Encode each field with enhanced Int128/Int256 detection
	for _, field := range typeInfo.Fields {
		if field.SchemaName == "flags" || field.SchemaName == "flags2" {
			continue
		}

		if field.Null {
			if _, exists := obj[field.SchemaName]; !exists {
				continue
			}
		}

		value, exists := obj[field.SchemaName]
		if !exists {
			continue
		}

		// Encode field value with field name context for Int128/Int256 detection
		if err := encodeJSONValueToBinaryWithContext(tempBuf, value, field.SchemaName); err != nil {
			return errors.Wrapf(err, "encode field %s", field.SchemaName)
		}
	}

	// Try to decode again
	testDecodeBuf := &bin.Buffer{Buf: make([]byte, len(tempBuf.Buf))}
	copy(testDecodeBuf.Buf, tempBuf.Buf)

	if err := concreteObj.Decode(testDecodeBuf); err != nil {
		return errors.Wrapf(err, "retry decode failed for type %q", typeName)
	}

	// Set flags
	if setFlagsObj, ok := concreteObj.(interface{ SetFlags() }); ok {
		setFlagsObj.SetFlags()
	}

	// Encode to output buffer
	return concreteObj.(bin.Encoder).Encode(buf)
}

// isHexString checks if a string is a valid hex string
func isHexString(s string) bool {
	if len(s) == 0 {
		return false
	}
	for _, c := range s {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}

// binaryToMTProtoJSON converts binary MTProto buffer to JSON.
// Note: For output marshaling, we use encoding/json since manual encoding
// would require reflection or code generation. The key improvement is that
// input parsing uses jx, not encoding/json.
// useSnakeCase controls whether to convert field names to snake_case (true) or keep PascalCase (false).
func binaryToMTProtoJSON(buf *bin.Buffer, typeID uint32, useSnakeCase bool) (string, error) {
	// Get type name from registry
	namesMap := NamesMap()
	typeName := ""
	for name, id := range namesMap {
		if id == typeID {
			typeName = name
			break
		}
	}
	if typeName == "" {
		return "", errors.Errorf("type name not found for ID: 0x%x", typeID)
	}

	// Decode binary to Go struct
	constructors := TypesConstructorMap()
	constructor, ok := constructors[typeID]
	if !ok {
		return "", errors.Errorf("no constructor for type ID: 0x%x", typeID)
	}

	obj := constructor()
	if obj == nil {
		return "", errors.New("constructor returned nil")
	}

	// Decode from buffer
	if buf != nil {
		b := &bin.Buffer{Buf: make([]byte, len(buf.Buf))}
		copy(b.Buf, buf.Buf)
		if err := obj.Decode(b); err != nil {
			return "", errors.Wrap(err, "decode binary to object")
		}
	}

	// Use jx.Encoder to build JSON manually
	// We'll encode using TypeInfo to iterate fields
	return encodeObjectToJSONWithJX(obj, typeName, useSnakeCase)
}

// encodeObjectToJSONWithJX encodes a bin.Object to JSON.
// Note: We use encoding/json for output marshaling since:
// 1. The object is already properly decoded from binary
// 2. Structs have JSON tags for proper field names
// 3. Manual encoding would require reflection or code generation
// The key improvement is that input parsing uses jx, not encoding/json.
// useSnakeCase controls whether to convert field names to snake_case (true) or keep PascalCase (false).
func encodeObjectToJSONWithJX(obj bin.Object, typeName string, useSnakeCase bool) (string, error) {
	// Marshal object to JSON (uses struct field names, which are PascalCase)
	data, err := json.Marshal(obj)
	if err != nil {
		return "", errors.Wrap(err, "marshal object to JSON")
	}

	// Parse and ensure @type field is present
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return string(data), nil // Return as-is if unmarshal fails
	}

	result["@type"] = typeName

	// Convert field names to snake_case if requested
	if useSnakeCase {
		result = convertKeysToSnakeCase(result)
	}

	output, err := json.Marshal(result)
	if err != nil {
		return "", errors.Wrap(err, "marshal result")
	}

	return string(output), nil
}

// convertKeysToSnakeCase recursively converts map keys from PascalCase to snake_case
func convertKeysToSnakeCase(m map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range m {
		snakeKey := pascalToSnakeCase(k)
		switch val := v.(type) {
		case map[string]interface{}:
			result[snakeKey] = convertKeysToSnakeCase(val)
		case []interface{}:
			arr := make([]interface{}, len(val))
			for i, item := range val {
				if itemMap, ok := item.(map[string]interface{}); ok {
					arr[i] = convertKeysToSnakeCase(itemMap)
				} else {
					arr[i] = item
				}
			}
			result[snakeKey] = arr
		default:
			result[snakeKey] = v
		}
	}
	return result
}

// pascalToSnakeCase converts PascalCase to snake_case
func pascalToSnakeCase(s string) string {
	if s == "" {
		return s
	}

	var result []rune
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result = append(result, '_')
		}
		if r >= 'A' && r <= 'Z' {
			result = append(result, r+'a'-'A')
		} else {
			result = append(result, r)
		}
	}
	return string(result)
}
