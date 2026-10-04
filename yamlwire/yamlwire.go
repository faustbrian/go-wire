package yamlwire

import (
	"bytes"
	"errors"
	"io"
	"math"
	"reflect"
	"regexp"
	"strings"

	"github.com/faustbrian/go-wire/v2"
	"github.com/faustbrian/go-wire/v2/internal/outputlimit"
	"github.com/faustbrian/go-wire/v2/internal/valuecheck"
	"go.yaml.in/yaml/v4"
	"go.yaml.in/yaml/v4/plugin/limit"
)

// DefaultMaxBytes is the default maximum YAML stream size.
const DefaultMaxBytes int64 = 1 << 20

// ErrPayloadTooLarge identifies YAML streams over the configured byte limit.
var ErrPayloadTooLarge = errors.New("payload exceeds size limit")

var (
	errAliasLimit       = errors.New("YAML alias limit exceeded")
	errAliasesDisabled  = errors.New("YAML aliases are disabled")
	errDepthLimit       = errors.New("YAML nesting depth limit exceeded")
	errMergeKeyDisabled = errors.New("YAML merge keys are disabled")
)

// DecodeOptions controls YAML parsing, interoperability, and resource limits.
// Zero limits retain safe library defaults. AllowMultipleDocuments requires a
// pointer to a slice and decodes every document into that slice.
type DecodeOptions struct {
	MaxBytes               int64
	MaxDepth               int
	MaxAliases             int
	DisallowUnknownFields  bool
	AllowDuplicateKeys     bool
	AllowMultipleDocuments bool
	DisallowAliases        bool
	DisallowMergeKeys      bool
}

// EncodeOptions controls deterministic YAML serialization.
type EncodeOptions struct {
	MaxBytes              int64
	Indent                int
	DefaultSequenceIndent bool
}

// Decode parses a bounded YAML stream into target.
func Decode(payload []byte, target any, options DecodeOptions) error {
	return DecodeReader(bytes.NewReader(payload), target, options)
}

// DecodeReader reads a bounded YAML stream and decodes it into target.
func DecodeReader(reader io.Reader, target any, options DecodeOptions) error {
	if err := validateOptions(options); err != nil {
		return wrap(wire.ErrorKindValidation, "decode options", err)
	}
	if err := validateTarget(target, options.AllowMultipleDocuments); err != nil {
		return wrap(wire.ErrorKindTarget, "decode", err)
	}
	payload, err := readBounded(reader, options.MaxBytes)
	if err != nil {
		return err
	}
	if err := preflightDocuments(payload, options); err != nil {
		return err
	}

	yamlOptions := decodeYAMLOptions(options)
	if err := yaml.Load(payload, target, yamlOptions...); err != nil {
		return classifyDecodeError(err)
	}
	return nil
}

// Encode serializes value deterministically using sorted mapping keys.
func Encode(value any, options EncodeOptions) ([]byte, error) {
	return encode(value, options, yaml.NewDumper)
}

func encode(
	value any,
	options EncodeOptions,
	newDumper func(io.Writer, ...yaml.Option) (*yaml.Dumper, error),
) ([]byte, error) {
	if err := valuecheck.Validate(value); err != nil {
		return nil, wrap(wire.ErrorKindEncode, "encode", err)
	}
	if !validIndent(options.Indent) {
		return nil, wrap(wire.ErrorKindValidation, "encode options", errors.New("indent must be between 2 and 9"))
	}
	// Negative width becomes a finite int32 limit in the pinned emitter.
	// No addressable int column can exceed MaxInt, so plain continuations
	// cannot wrap into marker-like lines consumed by normalization.
	yamlOptions := []yaml.Option{yaml.WithV4Defaults(), yaml.WithLineWidth(math.MaxInt)}
	if options.Indent != 0 {
		yamlOptions = append(yamlOptions, yaml.WithIndent(options.Indent))
	}
	if options.DefaultSequenceIndent {
		yamlOptions = append(yamlOptions, yaml.WithCompactSeqIndent(false))
	}
	// The pinned folded emitter can insert an extra LF beside each authored
	// LF. Each removed byte has a retained break partner, so at most twice
	// the final quota is needed before normalization. Do not allocate from
	// this ceiling; the buffer grows only as the provider produces bytes.
	intermediateMax := intermediateOutputLimit(options.MaxBytes)
	output, err := outputlimit.New(intermediateMax, DefaultMaxBytes)
	if err != nil {
		return nil, wrap(wire.ErrorKindValidation, "encode options", err)
	}
	dumper, err := newDumper(output, yamlOptions...)
	if err != nil {
		return nil, wrap(wire.ErrorKindValidation, "encode options", err)
	}
	err = dumper.Dump(value)
	if err == nil {
		err = dumper.Close()
	}
	if err != nil {
		if errors.Is(err, outputlimit.ErrLimit) {
			return nil, wrap(wire.ErrorKindSizeLimit, "encode", errors.Join(ErrPayloadTooLarge, err))
		}
		kind := wire.ErrorKindEncode
		if strings.Contains(err.Error(), "cannot marshal") || strings.Contains(err.Error(), "cannot represent") || strings.Contains(err.Error(), "unsupported") {
			kind = wire.ErrorKindUnsupported
		}
		return nil, wrap(kind, "encode", err)
	}
	indent := options.Indent
	if indent == 0 {
		indent = 2
	}
	payload, err := addBlockIndentIndicators(output.Bytes(), indent, options.MaxBytes)
	if err != nil {
		return nil, wrap(wire.ErrorKindSizeLimit, "encode", errors.Join(ErrPayloadTooLarge, err))
	}
	return payload, nil
}

func addBlockIndentIndicators(payload []byte, indent int, configuredMax int64) ([]byte, error) {
	maxBytes := configuredMax
	if maxBytes == 0 {
		maxBytes = DefaultMaxBytes
	}
	// Authored comments and block bodies retain every YAML logical break,
	// not just LF. Keep the separators byte-for-byte in their source slices.
	var lines [][]byte
	start := 0
	for _, lineBreak := range dumperLineBreaks.FindAllIndex(payload, -1) {
		lines = append(lines, payload[start:lineBreak[1]])
		start = lineBreak[1]
	}
	lines = append(lines, payload[start:])
	type hint struct {
		line, index int
		replace     bool
		value       byte
	}
	var hints []hint
	additions, bodyIndent := 0, 0
	foldedChanged := false
	var context scalarTextContext
	for lineIndex, line := range lines {
		if len(bytes.TrimSpace(line)) == 0 {
			continue
		}
		if bodyIndent > 0 && leadingSpaces(line) >= bodyIndent {
			continue
		}
		bodyIndent = 0
		continued := context.quote != 0 || context.flowDepth > 0
		context.scan(line)
		if continued || context.quote != 0 || context.flowDepth > 0 {
			continue
		}
		header := blockScalarHeader(line)
		if !hasBlockScalar(header.index) {
			continue
		}
		// The pinned emitter aligns mapping content to the configured grid,
		// but advances only two columns after a compact sequence indicator.
		width := indent - header.parentIndent%indent
		if header.sequence {
			width = 2
		}
		bodyIndent = header.parentIndent + width
		if line[header.index] == '>' {
			delta, changed := preserveFoldedBreaks(lines[lineIndex+1:], bodyIndent)
			additions += delta
			foldedChanged = foldedChanged || changed
		}
		index, replace := header.index+1, false
		if header.digit > 0 {
			index, replace = header.digit, true
		}
		value := byte('0' + width)
		if replace && line[index] == value {
			continue
		}
		hints = append(hints, hint{lineIndex, index, replace, value})
		if !replace {
			additions++
		}
	}
	capacity := outputCapacity(len(payload), additions)
	if capacity == -1 || int64(capacity) > maxBytes {
		return nil, outputlimit.ErrLimit
	}
	if len(hints) == 0 && !foldedChanged {
		return payload, nil
	}
	result := make([]byte, 0, capacity)
	next := 0
	for lineIndex, line := range lines {
		if next == len(hints) || hints[next].line != lineIndex {
			result = append(result, line...)
			continue
		}
		edit := hints[next]
		result = append(result, line[:edit.index]...)
		result = append(result, edit.value)
		if edit.replace {
			edit.index++
		}
		result = append(result, line[edit.index:]...)
		next++
	}
	return result, nil
}

// preserveFoldedBreaks corrects the pinned emitter's fixed-origin LF
// lookahead using its actual folded body. It never visits the represented
// graph or re-evaluates a Marshaler. Blank logical lines belong to this body;
// the first nonblank line outside its indentation ends it.
func preserveFoldedBreaks(lines [][]byte, indent int) (int, bool) {
	end, first := len(lines), -1
	for i, line := range lines {
		content := bytes.TrimRight(line, "\r\n\u0085\u2028\u2029")
		if len(content) == 0 {
			continue
		}
		if leadingSpaces(content) < indent {
			end = i
			break
		}
		// The emitter writes indentation only immediately before an
		// authored non-break character. Empty body breaks are unindented.
		if first < 0 {
			first = i
		}
	}
	if first < 0 {
		return 0, false
	}
	leading := lines[first][indent] == ' ' || lines[first][indent] == '\t'
	delta, changed := 0, false
	for i := first; i < end; i++ {
		line := lines[i]
		content := bytes.TrimRight(line, "\r\n\u0085\u2028\u2029")
		if len(content) == 0 || content[indent] == ' ' || content[indent] == '\t' || !bytes.HasSuffix(line, []byte{'\n'}) {
			continue
		}
		// Every intervening blank line is crossed only once: the outer loop
		// ignores them, and no two content lines share their following run.
		next := i + 1
		for next < end && len(bytes.TrimRight(lines[next], "\r\n\u0085\u2028\u2029")) == 0 {
			next++
		}
		ordinary := next < end && lines[next][indent] != ' ' && lines[next][indent] != '\t'
		if leading && ordinary {
			lines[i] = append(append([]byte(nil), line...), '\n')
			delta++
			changed = true
		} else if !leading && !ordinary && bytes.Equal(lines[i+1], []byte{'\n'}) {
			// A second LF proves a compensating break is present. The
			// splitter retains a following slice for every LF; a sibling
			// ending this body is nonblank and cannot match the second LF.
			// A lone closing LF may represent no final authored break.
			lines[i] = line[:len(line)-1]
			delta--
			changed = true
		}
	}
	return delta, changed
}

type scalarHeader struct {
	index, digit, parentIndent int
	sequence                   bool
}

func leadingSpaces(line []byte) int {
	return len(line) - len(bytes.TrimLeft(line, " "))
}

// These expressions recognize only the pinned emitter's prefixes and hints.
// Explicit-key/value and sequence indicators compose; hints put width first.
var (
	dumperLineBreaks   = regexp.MustCompile(`\r\n|[\r\n\x{0085}\x{2028}\x{2029}]`)
	dumperScalarPrefix = regexp.MustCompile(`^ *(?:[-?:] +)*`)
	dumperBlockHints   = regexp.MustCompile(`^(?:([2-9]))?[+-]?(?: +(?:#.*)?)?$`)
)

// blockScalarHeader recognizes scalar syntax in dumper-produced lines,
// not marker-like suffixes within a plain scalar. Block bodies are skipped
// by the caller, so their text is never interpreted as another header.
func blockScalarHeader(line []byte) scalarHeader {
	absent := scalarHeader{index: -1}
	line = bytes.TrimRight(line, "\r\n\u0085\u2028\u2029")
	prefix := dumperScalarPrefix.Find(line)
	start := len(prefix)
	parent, sequence := start, false
	if indicators := bytes.TrimRight(prefix, " "); len(indicators) > 0 {
		parent = len(indicators) - 1
		sequence = indicators[parent] == '-'
	}
	if start == len(line) || line[start] == '#' {
		return absent
	}
	// A quoted mapping key can contain colons and escaped quotes. Only a
	// colon followed by whitespace outside that key introduces its value.
	var quote byte
	keyStart := scalarPropertiesEnd(line, start)
	if keyStart < len(line) && line[keyStart] != '|' && line[keyStart] != '>' {
		// Nonempty flow collections cannot be simple keys in this emitter.
		// Empty collection keys can precede genuine block values on one line.
		// All other flow lines contain scalar data, not block headers.
		if line[keyStart] == '[' || line[keyStart] == '{' {
			if !bytes.HasPrefix(line[keyStart:], []byte("[]: ")) && !bytes.HasPrefix(line[keyStart:], []byte("{}: ")) {
				return absent
			}
			keyStart += 2
		}
		skipNext := false
		for offset, value := range line[keyStart:] {
			index := keyStart + offset
			if skipNext {
				skipNext = false
				continue
			}
			if quote != 0 {
				if quote == '"' && value == '\\' {
					skipNext = true
					continue
				}
				if value == quote {
					if quote == '\'' && index+1 < len(line) && line[index+1] == '\'' {
						skipNext = true
						continue
					}
					quote = 0
				}
				continue
			}
			// A comment ends the key search; its colons and markers are data.
			if value == '#' && line[index-1] == ' ' {
				return absent
			}
			if index == keyStart && (value == '\'' || value == '"') {
				quote = value
				continue
			}
			if value == ':' && index+1 < len(line) && line[index+1] == ' ' {
				parent, sequence = start, false
				start = index + 2
				break
			}
		}
	}
	start = scalarPropertiesEnd(line, start)
	if start == len(line) || (line[start] != '|' && line[start] != '>') {
		return absent
	}
	header := scalarHeader{index: start, parentIndent: parent, sequence: sequence}
	// The pinned dumper emits original explicit widths only in 2..9.
	// Normalization may write width 1 at an odd grid boundary, but never
	// feeds its own output back through this recognizer.
	hints := dumperBlockHints.FindSubmatchIndex(line[start+1:])
	if hints == nil {
		return absent
	}
	if hints[2] >= 0 {
		header.digit = start + 1
	}
	return header
}

// scalarTextContext follows quoted and flow continuations emitted by the
// pinned dumper. Block bodies are excluded before scanning: their characters
// are data, including unmatched quotes and collection indicators.
// Completed emission places delimiters, whitespace or logical breaks after
// syntactic quotes and indicators; this is not an arbitrary-input parser.
type scalarTextContext struct {
	quote     byte
	flowDepth int
}

func (context *scalarTextContext) scan(line []byte) {
	start, skipNext, property := true, false, false
	for index, value := range line {
		if skipNext {
			skipNext = false
			continue
		}
		if property {
			if value == ' ' {
				property = false
			}
			continue
		}
		if context.quote != 0 {
			if context.quote == '"' && value == '\\' {
				skipNext = true
				continue
			}
			if value == context.quote {
				if value == '\'' && line[index+1] == '\'' {
					skipNext = true
					continue
				}
				context.quote = 0
			}
			start = false
			continue
		}
		if value == ' ' || value == '\r' || value == '\n' {
			continue
		}
		if value == '#' && (index == 0 || line[index-1] == ' ') {
			return
		}
		if start && (value == '!' || value == '&') {
			property = true
			continue
		}
		if start && (value == '\'' || value == '"') {
			context.quote = value
			continue
		}
		if start && (value == '[' || value == '{') {
			context.flowDepth++
			continue
		}
		if context.flowDepth > 0 && (value == ']' || value == '}') {
			context.flowDepth--
			start = false
			continue
		}
		if context.flowDepth > 0 && value == ',' {
			start = true
			continue
		}
		if value == ':' && line[index+1] == ' ' {
			start = true
			continue
		}
		if start && (value == '-' || value == '?' || value == ':') && line[index+1] == ' ' {
			continue
		}
		start = false
	}
}

// scalarPropertiesEnd skips emitted tags and anchors, but not scalar data.
func scalarPropertiesEnd(line []byte, index int) int {
	rest := bytes.TrimLeft(line[index:], " ")
	for len(rest) > 0 && (rest[0] == '!' || rest[0] == '&') {
		_, after, _ := bytes.Cut(rest, []byte{' '})
		rest = bytes.TrimLeft(after, " ")
	}
	return len(line) - len(rest)
}

// EncodeWriter serializes value and writes one complete YAML document.
func EncodeWriter(writer io.Writer, value any, options EncodeOptions) error {
	payload, err := Encode(value, options)
	if err != nil {
		return err
	}
	if writer == nil {
		return wrap(wire.ErrorKindValidation, "write", errors.New("writer must not be nil"))
	}
	if _, err := io.Copy(writer, bytes.NewReader(payload)); err != nil {
		return wrap(wire.ErrorKindWrite, "write", err)
	}
	return nil
}

func validateOptions(options DecodeOptions) error {
	if options.MaxBytes < 0 {
		return errors.New("max bytes must not be negative")
	}
	if options.MaxDepth < 0 {
		return errors.New("max depth must not be negative")
	}
	if options.MaxAliases < 0 {
		return errors.New("max aliases must not be negative")
	}
	return nil
}

func validateTarget(target any, multiple bool) error {
	if target == nil {
		return errors.New("target must be a non-nil pointer")
	}
	value := reflect.ValueOf(target)
	if value.Kind() != reflect.Pointer || value.IsNil() {
		return errors.New("target must be a non-nil pointer")
	}
	if multiple && value.Elem().Kind() != reflect.Slice {
		return errors.New("multi-document target must point to a slice")
	}
	return nil
}

func readBounded(reader io.Reader, configuredMax int64) ([]byte, error) {
	if reader == nil {
		return nil, wrap(wire.ErrorKindValidation, "read", errors.New("reader must not be nil"))
	}
	maxBytes := configuredMax
	if maxBytes == 0 {
		maxBytes = DefaultMaxBytes
	}
	readLimit := maxBytes + 1
	if maxBytes == math.MaxInt64 {
		readLimit = maxBytes
	}
	payload, err := io.ReadAll(io.LimitReader(reader, readLimit))
	if err != nil {
		return nil, wrap(wire.ErrorKindParse, "read", err)
	}
	if exceedsLimit(len(payload), maxBytes) {
		return nil, wrap(wire.ErrorKindSizeLimit, "read", ErrPayloadTooLarge)
	}
	return payload, nil
}

func decodeYAMLOptions(options DecodeOptions) []yaml.Option {
	yamlOptions := baseDecodeYAMLOptions(options)
	if options.AllowMultipleDocuments {
		return append(yamlOptions, yaml.WithAllDocuments())
	}
	return append(yamlOptions, yaml.WithSingleDocument())
}

func baseDecodeYAMLOptions(options DecodeOptions) []yaml.Option {
	yamlOptions := []yaml.Option{
		yaml.WithV4Defaults(),
		yaml.WithKnownFields(options.DisallowUnknownFields),
		yaml.WithUniqueKeys(!options.AllowDuplicateKeys),
	}
	if needsLimitPlugin(options) {
		limitOptions := make([]limit.Option, 0, 2)
		if depthLimitEnabled(options.MaxDepth) {
			limitOptions = append(limitOptions, limit.DepthFunc(func(depth int, _ *limit.DepthContext) error {
				if exceedsDepth(depth, options.MaxDepth) {
					return errDepthLimit
				}
				return nil
			}))
		}
		if aliasLimitEnabled(options.DisallowAliases, options.MaxAliases) {
			limitOptions = append(limitOptions, limit.AliasFunc(func(aliasCount, _ int) error {
				if aliasesDisabled(options.DisallowAliases, aliasCount) {
					return errAliasesDisabled
				}
				if aliasesExceeded(options.MaxAliases, aliasCount) {
					return errAliasLimit
				}
				return nil
			}))
		}
		yamlOptions = append(yamlOptions, yaml.WithPlugin(limit.New(limitOptions...)))
	}
	return yamlOptions
}

func validIndent(indent int) bool {
	return indent == 0 || indent >= 2 && indent <= 9
}

func hasBlockScalar(index int) bool {
	return index >= 0
}

func exceedsLimit(length int, maximum int64) bool {
	return int64(length) > maximum
}

// outputCapacity computes the normalized allocation size without overflowing.
// length is a materialized slice length; delta may grow or shrink its body.
func outputCapacity(length, delta int) int {
	if delta > math.MaxInt-length || delta < -length {
		return -1
	}
	return length + delta
}

// intermediateOutputLimit leaves invalid negative limits for validation and
// saturates the doubled scratch ceiling without a quota-sized allocation.
func intermediateOutputLimit(maxBytes int64) int64 {
	if maxBytes < 0 {
		return maxBytes
	}
	if maxBytes == 0 {
		maxBytes = DefaultMaxBytes
	}
	if maxBytes > math.MaxInt64/2 {
		return math.MaxInt64
	}
	return maxBytes * 2
}

func needsLimitPlugin(options DecodeOptions) bool {
	return options.DisallowAliases || options.MaxAliases > 0 || options.MaxDepth > 0
}

func depthLimitEnabled(maximum int) bool {
	return maximum > 0
}

func exceedsDepth(depth, maximum int) bool {
	return depth > maximum
}

func aliasLimitEnabled(disallow bool, maximum int) bool {
	return disallow || maximum > 0
}

func aliasesDisabled(disallow bool, count int) bool {
	return disallow && count > 0
}

func aliasesExceeded(maximum, count int) bool {
	return maximum > 0 && count > maximum
}

func preflightDocuments(payload []byte, options DecodeOptions) error {
	var documents []yaml.Node
	yamlOptions := baseDecodeYAMLOptions(options)
	yamlOptions = append(yamlOptions, yaml.WithAllDocuments())
	if err := yaml.Load(payload, &documents, yamlOptions...); err != nil {
		return classifyDecodeError(err)
	}
	if !options.AllowMultipleDocuments && len(documents) != 1 {
		return wrap(wire.ErrorKindParse, "decode", errors.New("YAML stream must contain exactly one document"))
	}
	if options.DisallowMergeKeys {
		for i := range documents {
			if hasMergeKey(&documents[i]) {
				return wrap(wire.ErrorKindUnsupported, "decode", errMergeKeyDisabled)
			}
		}
	}
	return nil
}

func hasMergeKey(node *yaml.Node) bool {
	if node.Tag == "!!merge" {
		return true
	}
	for _, child := range node.Content {
		if hasMergeKey(child) {
			return true
		}
	}
	return false
}

func classifyDecodeError(err error) error {
	message := err.Error()
	switch {
	case errors.Is(err, errAliasLimit), errors.Is(err, errDepthLimit),
		strings.Contains(message, errAliasLimit.Error()), strings.Contains(message, errDepthLimit.Error()),
		strings.Contains(message, "document contains excessive aliasing"),
		strings.Contains(message, "exceeded max depth"):
		return wrap(wire.ErrorKindSizeLimit, "decode", err)
	case errors.Is(err, errAliasesDisabled), strings.Contains(message, errAliasesDisabled.Error()):
		return wrap(wire.ErrorKindUnsupported, "decode", err)
	case strings.Contains(message, "field "):
		return wrap(wire.ErrorKindValidation, "decode", err)
	default:
		return wrap(wire.ErrorKindParse, "decode", err)
	}
}

func wrap(kind wire.ErrorKind, op string, err error) error {
	return &wire.Error{Kind: kind, Format: wire.FormatYAML, Op: op, Err: err}
}
