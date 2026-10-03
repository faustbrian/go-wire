package cborwire_test

import (
	"bytes"
	"errors"
	"reflect"
	"testing"

	"github.com/faustbrian/go-wire"
	"github.com/faustbrian/go-wire/cborwire"
	"github.com/fxamacker/cbor/v2"
)

type namedMapKey any
type structMapKey struct{ X any }

func TestDecodeDynamicMapKeyContracts(t *testing.T) {
	// Literal CBOR maps provide an oracle independent of the upgraded encoder.
	testDynamicMapKey(t, "named interface", []byte{0xa1, 0x81, 0x01, 0x02}, []byte{0xa1, 0x01, 0x02}, namedMapKey(uint64(1)), "[]interface {}", cborwire.DecodeOptions{})
	testDynamicMapKey(t, "array", []byte{0xa1, 0x81, 0x81, 0x01, 0x02}, []byte{0xa1, 0x81, 0x01, 0x02}, [1]any{uint64(1)}, "[1]interface {}", cborwire.DecodeOptions{})
	testDynamicMapKey(t, "struct", []byte{0xa1, 0xa1, 0x61, 'X', 0x81, 0x01, 0x02}, []byte{0xa1, 0xa1, 0x61, 'X', 0x01, 0x02}, structMapKey{X: uint64(1)}, reflect.TypeFor[structMapKey]().String(), cborwire.DecodeOptions{})
	testDynamicMapKey(t, "generic tag", []byte{0xa1, 0xd8, 0x8b, 0x81, 0x01, 0x02}, []byte{0xa1, 0xd8, 0x8b, 0x01, 0x02}, cbor.Tag{Number: 139, Content: uint64(1)}, "cbor.Tag", cborwire.DecodeOptions{AllowTags: true})
}

func testDynamicMapKey[K comparable](t *testing.T, name string, bad, good []byte, key K, badType string, options cborwire.DecodeOptions) {
	t.Helper()
	for _, reader := range []bool{false, true} {
		entry := "Decode"
		if reader {
			entry = "DecodeReader"
		}
		t.Run(name+"/"+entry, func(t *testing.T) {
			t.Run("uncomparable", func(t *testing.T) {
				// A recovered panic remains a failing assertion, allowing every family
				// and entry point to be observed in the old-dependency replay.
				defer func() {
					if p := recover(); p != nil {
						t.Errorf("decode panicked: %v", p)
					}
				}()
				var target map[K]any
				err := decodeMapKey(bad, &target, options, reader)
				if !errors.Is(err, wire.ErrParse) {
					t.Fatalf("error = %v, want parse failure", err)
				}
				var detail *wire.Error
				if !errors.As(err, &detail) || detail.Kind != wire.ErrorKindParse || detail.Format != wire.FormatCBOR || detail.Op != "decode" {
					t.Fatalf("wire error context = %#v", detail)
				}
				var cause *cbor.InvalidMapKeyTypeError
				if !errors.As(err, &cause) || cause.GoType != badType {
					t.Fatalf("map key cause = %#v, want %q", cause, badType)
				}
			})
			t.Run("comparable", func(t *testing.T) {
				var target map[K]any
				if err := decodeMapKey(good, &target, options, reader); err != nil {
					t.Fatal(err)
				}
				want := map[K]any{key: uint64(2)}
				if !reflect.DeepEqual(target, want) {
					t.Fatalf("decoded map = %#v, want %#v", target, want)
				}
			})
		})
	}
}

func TestDecodeTypedTagMapKeysRetainDefaultPolicy(t *testing.T) {
	for _, payload := range [][]byte{{0xa1, 0xd8, 0x8b, 0x81, 0x01, 0x02}, {0xa1, 0xd8, 0x8b, 0x01, 0x02}} {
		for _, reader := range []bool{false, true} {
			var target map[cbor.Tag]any
			err := decodeMapKey(payload, &target, cborwire.DecodeOptions{}, reader)
			if !errors.Is(err, wire.ErrUnsupportedFormat) || errors.Is(err, wire.ErrParse) {
				t.Fatalf("default tag policy error = %v", err)
			}
			var cause *cbor.TagsMdError
			if !errors.As(err, &cause) {
				t.Fatalf("tag policy cause = %T, want TagsMdError", err)
			}
		}
	}
}

func decodeMapKey(payload []byte, target any, options cborwire.DecodeOptions, reader bool) error {
	if reader {
		return cborwire.DecodeReader(bytes.NewReader(payload), target, options)
	}
	return cborwire.Decode(payload, target, options)
}
