package bilibili

import (
	"encoding/binary"
	"errors"
)

// The APP API speaks protobuf. Only a handful of small messages are exchanged, so instead of a protobuf library
// there is this wire-format writer and reader; the message layouts live next to their use in app.go.

// Wire types.
const (
	wireVarint  = 0
	wireFixed64 = 1
	wireBytes   = 2
	wireFixed32 = 5
)

// pbWriter encodes a message field by field. Fields are written whether or not they hold their default, the way
// proto2 serializes fields that were set: the servers read "download = 0" and an empty buvid as sent.
type pbWriter struct{ buf []byte }

func (w *pbWriter) tag(field, wire int) {
	w.buf = binary.AppendUvarint(w.buf, uint64(field)<<3|uint64(wire))
}

func (w *pbWriter) uint(field int, v uint64) *pbWriter {
	w.tag(field, wireVarint)
	w.buf = binary.AppendUvarint(w.buf, v)
	return w
}

// int encodes int32, int64 and enums: negative numbers as ten-byte two's complement, as protobuf does.
func (w *pbWriter) int(field int, v int64) *pbWriter { return w.uint(field, uint64(v)) }

func (w *pbWriter) bool(field int, v bool) *pbWriter {
	if v {
		return w.uint(field, 1)
	}
	return w.uint(field, 0)
}

func (w *pbWriter) bytes(field int, b []byte) *pbWriter {
	w.tag(field, wireBytes)
	w.buf = binary.AppendUvarint(w.buf, uint64(len(b)))
	w.buf = append(w.buf, b...)
	return w
}

func (w *pbWriter) string(field int, s string) *pbWriter { return w.bytes(field, []byte(s)) }

func (w *pbWriter) message(field int, m *pbWriter) *pbWriter { return w.bytes(field, m.buf) }

// pbField is one decoded field: a varint (fixed-width numbers are kept as their bits) or length-delimited bytes.
type pbField struct {
	num   int
	wire  int
	value uint64
	data  []byte
}

// pbMessage is a decoded message, its fields in wire order. Accessors take the last occurrence of a singular
// field, as protobuf does, and yield zero values for missing ones.
type pbMessage []pbField

var errTruncated = errors.New("protobuf: truncated message")

// decodeProto splits a message into fields. Groups (long deprecated) are not supported.
func decodeProto(b []byte) (pbMessage, error) {
	var m pbMessage
	for len(b) > 0 {
		key, n := binary.Uvarint(b)
		if n <= 0 {
			return nil, errTruncated
		}
		b = b[n:]
		f := pbField{num: int(key >> 3), wire: int(key & 7)}
		switch f.wire {
		case wireVarint:
			f.value, n = binary.Uvarint(b)
			if n <= 0 {
				return nil, errTruncated
			}
			b = b[n:]
		case wireFixed64:
			if len(b) < 8 {
				return nil, errTruncated
			}
			f.value, b = binary.LittleEndian.Uint64(b), b[8:]
		case wireFixed32:
			if len(b) < 4 {
				return nil, errTruncated
			}
			f.value, b = uint64(binary.LittleEndian.Uint32(b)), b[4:]
		case wireBytes:
			size, n := binary.Uvarint(b)
			if n <= 0 || uint64(len(b)-n) < size {
				return nil, errTruncated
			}
			f.data, b = b[n:n+int(size)], b[n+int(size):]
		default:
			return nil, errors.New("protobuf: unsupported wire type")
		}
		if f.num == 0 {
			return nil, errors.New("protobuf: field number 0")
		}
		m = append(m, f)
	}
	return m, nil
}

func (m pbMessage) last(num int) (pbField, bool) {
	for i := len(m) - 1; i >= 0; i-- {
		if m[i].num == num {
			return m[i], true
		}
	}
	return pbField{}, false
}

func (m pbMessage) has(num int) bool {
	_, ok := m.last(num)
	return ok
}

func (m pbMessage) uint(num int) uint64 {
	f, _ := m.last(num)
	return f.value
}

func (m pbMessage) int(num int) int64 { return int64(m.uint(num)) }

func (m pbMessage) string(num int) string {
	f, _ := m.last(num)
	return string(f.data)
}

func (m pbMessage) strings(num int) []string {
	var out []string
	for _, f := range m {
		if f.num == num && f.wire == wireBytes {
			out = append(out, string(f.data))
		}
	}
	return out
}

// message is an embedded message; a malformed one reads as empty, like a missing one.
func (m pbMessage) message(num int) pbMessage {
	f, ok := m.last(num)
	if !ok || f.wire != wireBytes {
		return nil
	}
	sub, _ := decodeProto(f.data)
	return sub
}

func (m pbMessage) messages(num int) []pbMessage {
	var out []pbMessage
	for _, f := range m {
		if f.num == num && f.wire == wireBytes {
			sub, _ := decodeProto(f.data)
			out = append(out, sub)
		}
	}
	return out
}
