package tdjson

// TDLibEncoder represents TDLib JSON API encoder.
type TDLibEncoder interface {
	EncodeTDLibJSON(Encoder) error
}

// TDLibDecoder represents TDLib JSON API decoder.
type TDLibDecoder interface {
	DecodeTDLibJSON(Decoder) error
}
