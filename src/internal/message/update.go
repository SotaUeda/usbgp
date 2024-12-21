package message

import (
	"fmt"

	"github.com/SotaUeda/usbgp/internal/ip"
	"github.com/SotaUeda/usbgp/internal/message/pathattribute"
)

type UpdateMessage struct {
	header           *Header
	wrBytesLen       uint16 // ルート数ではなく、bytesにしたときのオクテット数
	withdrawnRoutes  []*ip.IPv4Net
	pathAttrBytesLen uint16 // bytesにしたときのオクテット数
	pathAttributes   []pathattribute.PathAttribute
	nlri             []*ip.IPv4Net
	// NLRIのオクテット数はBGP UpdateMessageに含めず、
	// Headerのサイズを計算することにしか使用しないため、
	// メンバに含めていない。
}

func (*UpdateMessage) Type() Type {
	return Update
}

func NewUpdateMsg(
	pas []pathattribute.PathAttribute,
	nlri []*ip.IPv4Net,
	wr []*ip.IPv4Net,
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
		nlriLen += uint16(l)
	}
	wrLen := uint16(0)
	for _, w := range wr {
		l := w.Len()
		if l == 0 {
			return nil, fmt.Errorf("NLRIの長さが0です: %v", w)
		}
		wrLen += uint16(l)
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
	i := 21
	for _, w := range u.withdrawnRoutes {
		wb, err := w.MarshalBytes()
		if err != nil {
			return nil, err
		}
		i += copy(b[i:], wb)
	}
	// Path Attributes Length
	paLen := u.pathAttrBytesLen
	b[i] = uint8(paLen >> 8)
	b[i+1] = uint8(paLen)
	// Path Attributes
	i += 2
	for _, pa := range u.pathAttributes {
		pab, err := pa.MarshalBytes()
		if err != nil {
			return nil, err
		}
		i += copy(b[i:], pab)
	}
	// NLRI
	for _, n := range u.nlri {
		nb, err := n.MarshalBytes()
		if err != nil {
			return nil, err
		}
		i += copy(b[i:], nb)
	}
	return b, nil
}

func (u *UpdateMessage) unMarshalBytes(b []byte) error {
	// Header
	// message.goから利用する場合、Headerは作成済
	if u.header == nil {
		hLen := 19
		h := &Header{}
		err := h.unMarshalBytes(b[:hLen])
		if err != nil {
			return err
		}
		u.header = h
		b = b[hLen:]
	}

	// Withdrawn Routes Length
	if len(b) < 2 {
		return NewConvMsgErr(fmt.Sprintf("UpdateMessageのByte列が短すぎます length: %v", len(b)))
	}
	u.wrBytesLen = uint16(b[0])<<8 | uint16(b[1])

	// Withdrawn Routes
	i := 2
	j := i + int(u.wrBytesLen)
	if len(b) < j {
		return NewConvMsgErr(fmt.Sprintf("UpdateMessageのByte列が短すぎます length: %v", len(b)))
	}
	wr, err := ip.NewIPv4NetsFromBytes(b[i:j])
	if err != nil {
		err := ConvMsgErr{Err: err}
		return err
	}
	u.withdrawnRoutes = wr

	// Path Attributes Length
	i = j
	j = i + 2
	if len(b) < j {
		return NewConvMsgErr(fmt.Sprintf("UpdateMessageのByte列が短すぎます length: %v", len(b)))
	}
	u.pathAttrBytesLen = uint16(b[i])<<8 | uint16(b[i+1])

	// Path Attributes
	i = j
	j = i + int(u.pathAttrBytesLen)
	if len(b) < j {
		return NewConvMsgErr(fmt.Sprintf("UpdateMessageのByte列が短すぎます length: %v", len(b)))
	}
	pas, err := pathattribute.NewPathAttributesFromBytes(b[i:j])
	if err != nil {
		err := ConvMsgErr{Err: err}
		return err
	}
	u.pathAttributes = pas

	// NLRI
	i = j
	nlri, err := ip.NewIPv4NetsFromBytes(b[i:])
	if err != nil {
		err := ConvMsgErr{Err: err}
		return err
	}
	u.nlri = nlri

	return nil
}

func (u *UpdateMessage) String() string {
	return fmt.Sprintf(
		"UpdateMessage{header: %v, wrBytesLen: %v, withdrawnRoutes: %v, "+
			"pathAttrBytesLen: %v, pathAttributes: %v, nlri: %v",
		u.header, u.wrBytesLen, u.withdrawnRoutes,
		u.pathAttrBytesLen, u.pathAttributes, u.nlri,
	)
}
