// Package outputlimit supplies New for private retained-output quotas, not codec
// allocation or atomic publication. Buffers are mutable and not concurrency-safe;
// Bytes aliases their storage. No external handles or workers are owned.
package outputlimit

import (
	"bytes"
	"errors"
)

// ErrLimit identifies encoded output that exceeds its configured byte limit.
var ErrLimit = errors.New("encoded output exceeds size limit")

// Buffer is an io.Writer that retains at most its configured byte limit.
type Buffer struct {
	buffer bytes.Buffer
	max    int64
}

// New rejects negative configured limits and uses fallback when configured is
// zero. Callers must provide a nonnegative fallback; it is not validated here.
// Construction does not preallocate the entire quota.
func New(configured, fallback int64) (*Buffer, error) {
	if configured < 0 {
		return nil, errors.New("max bytes must not be negative")
	}
	if configured == 0 {
		configured = fallback
	}
	return &Buffer{max: configured}, nil
}

// Write retains p when it fits and otherwise returns ErrLimit without
// retaining bytes beyond the configured maximum.
func (b *Buffer) Write(payload []byte) (int, error) {
	remaining := b.max - int64(b.buffer.Len())
	if int64(len(payload)) <= remaining {
		return b.buffer.Write(payload)
	}
	if remaining == 0 {
		return 0, ErrLimit
	}
	written, _ := b.buffer.Write(payload[:int(remaining)])
	return written, ErrLimit
}

// Bytes returns an alias of retained output; later writes can invalidate or
// modify it. Clone when an independent snapshot is required.
func (b *Buffer) Bytes() []byte {
	return b.buffer.Bytes()
}

// Len returns the number of retained bytes.
func (b *Buffer) Len() int {
	return b.buffer.Len()
}
