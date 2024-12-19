package message

import (
	"fmt"

	"github.com/SotaUeda/usbgp/internal/message/pathattribute"
	"github.com/SotaUeda/usbgp/internal/routing"
)

type UpdateMessage struct {
	header           *Header
	wrBytesLen       uint16 // ルート数ではなく、bytesにしたときのオクテット数
	withdrawnRoutes  []*routing.IPv4Net
	pathAttrBytesLen uint16 // bytesにしたときのオクテット数
	pathAttributes   []pathattribute.PathAttribute
	nlri             []*routing.IPv4Net
	// NLRIのオクテット数はBGP UpdateMessageに含めず、
	// Headerのサイズを計算することにしか使用しないため、
	// メンバに含めていない。
}

func (*UpdateMessage) Type() Type {
	return Update
}

func NewUpdateMsg(
	pas []pathattribute.PathAttribute,
	nlri []*routing.IPv4Net,
	wr []*routing.IPv4Net,
) (*UpdateMessage, error) {
	// PathAttributeの長さを計算
	paLen := uint16(0)
	for _, pa := range pas {
		paLen += pa.BytesLen()
	}
	nlriLen := uint16(0)
	for _, n := range nlri {
		l := n.Len()
		if l == 0 {
			return nil, fmt.Errorf("NLRIの長さが0です: %v", n)
		}
		nlriLen += l
	}
	wrLen := uint16(0)
	for _, w := range wr {
		l := w.Len()
		if l == 0 {
			return nil, fmt.Errorf("NLRIの長さが0です: %v", w)
		}
		wrLen += l
	}
	hMinLen := uint16(19)
	h, err := newHeader(
		hMinLen+
			paLen+
			nlriLen+
			wrLen+
			// +4はPahtAttributeLength(u16)と
			// WithdrawnRoutesLength(u16)のbytes表現分
			4,
		Update)
	if err != nil {
		return nil, err
	}
	u := &UpdateMessage{
		header:           h,
		wrBytesLen:       wrLen,
		withdrawnRoutes:  wr,
		pathAttrBytesLen: paLen,
		pathAttributes:   pas,
		nlri:             nlri,
	}
	return u, nil
}

func (u *UpdateMessage) marshalBytes() ([]byte, error) {
	b := make([]byte, u.header.len)
	// Header
	h, err := u.header.marshalBytes()
	if err != nil {
		return nil, err
	}
	copy(b, h)
	// Withdrawn Routes Length
	wrLen := u.wrBytesLen
	b[19] = uint8(wrLen >> 8)
	b[20] = uint8(wrLen)
	// Withdrawn Routes
	Idx := 21
	for _, w := range u.withdrawnRoutes {
		wb, err := w.MarshalBytes()
		if err != nil {
			return nil, err
		}
		Idx += copy(b[Idx:], wb)
	}
	// Path Attributes Length
	paLen := u.pathAttrBytesLen
	b[Idx] = uint8(paLen >> 8)
	b[Idx+1] = uint8(paLen)
	// Path Attributes
	Idx += 2
	for _, pa := range u.pathAttributes {
		pab, err := pa.MarshalBytes()
		if err != nil {
			return nil, err
		}
		Idx += copy(b[Idx:], pab)
	}
	// NLRI
	for _, n := range u.nlri {
		nb, err := n.MarshalBytes()
		if err != nil {
			return nil, err
		}
		Idx += copy(b[Idx:], nb)
	}
	return b, nil
}

func (u *UpdateMessage) unMarshalBytes(b []byte) error {
	// TODO
	return nil
}
