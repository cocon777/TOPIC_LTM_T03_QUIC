// đóng khung message cho cả TCP và QUIC ở benchmark.
package bench

import (
	"encoding/binary"
	"io"
)

// ghi 4 byte độ dài rồi ghi dữ liệu
func WriteMessage(w io.Writer, payload []byte) error {
	var hdr [4]byte
	binary.BigEndian.PutUint32(hdr[:], uint32(len(payload)))
	if _, err := w.Write(hdr[:]); err != nil {
		return err
	}
	_, err := w.Write(payload)
	return err
}
// đọc 4 byte độ dài, rồi đọc đủ số byte đó
func ReadMessage(r io.Reader) ([]byte, error) {
	var hdr [4]byte
	if _, err := io.ReadFull(r, hdr[:]); err != nil {
		return nil, err
	}
	size := binary.BigEndian.Uint32(hdr[:])
	buf := make([]byte, size)   // theo độ dài do bên kia gửi
	_, err := io.ReadFull(r, buf)
	return buf, err
}
