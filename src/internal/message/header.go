package message

import "fmt"

type Header struct {
	len     uint16
	msgType Type
}

func newHeader(len uint16, t Type) (*Header, error) {
	// No need to check len > 0xffff as len is of type uint16
	return &Header{
		len:     len,
		msgType: t,
	}, nil
}

func (h *Header) marshalBytes() ([]byte, error) {
	b := make([]byte, 19)
	// Marker
	for i := 0; i < 16; i++ {
		b[i] = 0xff
	}
	b[16] = uint8(h.len >> 8)
	b[17] = uint8(h.len)
	// Type
	b[18] = uint8(h.msgType)
	return b, nil
}

func (h *Header) unMarshalBytes(b []byte) error {
	// Headerの長さは19byte
	if len(b) != 19 {
		return NewConvMsgErr(fmt.Sprintf("HeaderのByte列が短すぎます: %d", len(b)))
	}
	// Marker
	for i := 0; i < 16; i++ {
		if b[i] != 0xff {
			return NewConvMsgErr(fmt.Sprintf("HeaderのMarkerが不正です:%v", b[:16]))
		}
	}
	// Length
	h.len = uint16(b[16])<<8 | uint16(b[17])
	// Type
	t, err := newType(b[18])
	if err != nil {
		return err
	}
	h.msgType = t
	return nil
}

func (h *Header) String() string {
	return fmt.Sprintf("Header{len: %d, type: %v}", h.len, h.msgType)
}
