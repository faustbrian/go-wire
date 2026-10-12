package wire

import "bytes"

// Format identifies a supported structured wire format.
type Format string

const (
	// FormatJSON identifies JSON text, including scalar values when explicitly decoded.
	FormatJSON Format = "json"
	// FormatXML identifies XML documents, not a schema or canonicalization profile.
	FormatXML Format = "xml"
	// FormatSOAP identifies SOAP envelopes; detection deliberately reports them as XML.
	FormatSOAP Format = "soap"
	// FormatYAML identifies YAML streams, not a JSON-compatible value model.
	FormatYAML Format = "yaml"
	// FormatTOML identifies TOML documents.
	FormatTOML Format = "toml"
	// FormatMessagePack names the MessagePack binary format.
	FormatMessagePack Format = "msgpack"
	// FormatCBOR identifies CBOR items; select their deterministic profile explicitly.
	FormatCBOR Format = "cbor"
	// FormatBSON identifies complete BSON documents, not arbitrary binary values.
	FormatBSON Format = "bson"
)

var utf8BOM = []byte{0xef, 0xbb, 0xbf}

// DetectFormat distinguishes JSON from XML using the first significant byte.
// SOAP is returned as XML because reliably identifying a SOAP envelope requires
// parsing its namespace; use the soap package when SOAP is expected.
// This is not syntax validation: only {/[ and < are hints, so valid JSON scalars
// are unsupported. It trims whitespace and one UTF-8 BOM without mutating input;
// unsupported/empty payloads match ErrUnsupportedFormat. There is no byte quota.
func DetectFormat(payload []byte) (Format, error) {
	payload = bytes.TrimSpace(payload)
	payload = bytes.TrimPrefix(payload, utf8BOM)
	payload = bytes.TrimSpace(payload)

	if len(payload) == 0 {
		return "", unsupportedError("detect format", payload)
	}

	switch payload[0] {
	case '{', '[':
		return FormatJSON, nil
	case '<':
		return FormatXML, nil
	default:
		return "", unsupportedError("detect format", payload)
	}
}
