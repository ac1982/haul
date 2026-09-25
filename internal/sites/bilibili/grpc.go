package bilibili

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"io"

	"github.com/ac1982/haul/internal/errs"
)

// gRPC length-prefixed framing: a compressed flag byte, a 4-byte big-endian length, then the message. The APP API
// is plain HTTPS POSTs carrying one such frame each way.

// packFrame gzips message into a compressed frame.
func packFrame(message []byte) []byte {
	var body bytes.Buffer
	zw := gzip.NewWriter(&body)
	_, _ = zw.Write(message) // writes to a bytes.Buffer cannot fail
	_ = zw.Close()
	out := make([]byte, 5, 5+body.Len())
	out[0] = 1
	binary.BigEndian.PutUint32(out[1:], uint32(body.Len()))
	return append(out, body.Bytes()...)
}

// unpackFrame is the message of the first frame; bytes past its declared length (trailers) are ignored.
func unpackFrame(data []byte) ([]byte, error) {
	if len(data) < 5 {
		return nil, errs.New("Incomplete gRPC response (%d bytes)", len(data))
	}
	body := data[5:]
	if n := binary.BigEndian.Uint32(data[1:5]); uint64(n) < uint64(len(body)) {
		body = body[:n]
	}
	if data[0] != 1 {
		return body, nil
	}
	zr, err := gzip.NewReader(bytes.NewReader(body))
	if err != nil {
		return nil, errs.New("Invalid gRPC response: %v", err)
	}
	out, err := io.ReadAll(zr)
	if err != nil {
		return nil, errs.New("Invalid gRPC response: %v", err)
	}
	return out, nil
}
