package message

import (
	"github.com/SotaUeda/usbgp/internal/pathattribute"
	"github.com/SotaUeda/usbgp/internal/routing"
)

type UpdateMessage struct {
	header           *Header
	wrBytesLen       uint16 // ルート数ではなく、bytesにしたときのオクテット数
	withdrawnRoutes  []*routing.IPv4NetWork
	pathAttrBytesLen uint16 // bytesにしたときのオクテット数
	pathAttributes   []pathattribute.PathAttribute
	nlri             []*routing.IPv4NetWork
	// NLRIのオクテット数はBGP UpdateMessageに含めず、
	// Headerのサイズを計算することにしか使用しないため、
	// メンバに含めていない。
}

func (*UpdateMessage) Type() Type {
	return Update
}

func NewUpdateMsg(
	pas []pathattribute.PathAttribute,
	nlri []*routing.IPv4NetWork,
	wr []*routing.IPv4NetWork,
) (*UpdateMessage, error) {
	// TODO
	h, err := newHeader(20, Update)
	if err != nil {
		return nil, err
	}
	u := &UpdateMessage{
		header: h,
	}
	return u, nil
}

func (u *UpdateMessage) marshalBytes() ([]byte, error) {
	// TODO
	return u.header.marshalBytes()
}

func (u *UpdateMessage) unMarshalBytes(b []byte) error {
	// TODO
	return nil
}
