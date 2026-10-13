# Package map

[Documentation index](README.md) · [Repository entry](../README.md)

This map follows the current source modules, not a synchronized release.
Use the repository README for published installation versions; an unpublished
source major is not installable merely because its package appears here.
Internal packages and non-releasable fixtures are intentionally excluded.

## github.com/faustbrian/go-wire/v3

`wire` provides explicit, auditable JSON, XML, SOAP, YAML, TOML, MessagePack, CBOR, and BSON interoperability boundaries with bounded read and write APIs.

| Import path | Responsibility |
| --- | --- |
| [`github.com/faustbrian/go-wire/v3`](../doc.go) | Use for shared format identifiers, stable error categories, and explicit JSON-or-XML detection. |
| [`github.com/faustbrian/go-wire/v3/bsonwire`](../bsonwire/doc.go) | Use for bounded BSON document encoding and decoding with explicit duplicate-key and numeric policies. |
| [`github.com/faustbrian/go-wire/v3/cborwire`](../cborwire/doc.go) | Use for bounded CBOR processing with explicit deterministic profiles, tag policy, and structural limits. |
| [`github.com/faustbrian/go-wire/v3/jsonwire`](../jsonwire/doc.go) | Use for bounded JSON encoding, decoding, and deterministic normalization with explicit strictness options. |
| [`github.com/faustbrian/go-wire/v3/msgpackwire`](../msgpackwire/doc.go) | Use for bounded MessagePack encoding and decoding with explicit map, extension, and numeric policies. |
| [`github.com/faustbrian/go-wire/v3/soap`](../soap/doc.go) | Use for bounded SOAP 1.1 and 1.2 envelope, body, and fault processing independent of HTTP policy. |
| [`github.com/faustbrian/go-wire/v3/tomlwire`](../tomlwire/doc.go) | Use for bounded TOML document encoding and decoding with native datetime and strict-field options. |
| [`github.com/faustbrian/go-wire/v3/xmlwire`](../xmlwire/doc.go) | Use for bounded XML document encoding and decoding with namespace, charset, and depth policy. |
| [`github.com/faustbrian/go-wire/v3/yamlwire`](../yamlwire/doc.go) | Use for bounded YAML document encoding and decoding with explicit alias, merge, and duplicate-key policy. |

Choose the next guide in the [documentation index](README.md)
or return to the [repository entry](../README.md) for installation and ownership.
