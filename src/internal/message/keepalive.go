package message

import "fmt"

type KeepaliveMessage struct {
	header *Header
}

func (*KeepaliveMessage) Type() Type {
	return Keepalive
}

func NewKeepaliveMsg() (*KeepaliveMessage, error) {
	h, err := newHeader(19, Keepalive)
	if err != nil {
		return nil, err
	}
	return &KeepaliveMessage{
		header: h,
	}, nil
}

func (k *KeepaliveMessage) marshalBytes() ([]byte, error) {
	return k.header.marshalBytes()
}

func (k *KeepaliveMessage) unMarshalBytes(b []byte) error {
	// Keepalive MessageはHeaderのみ
	// Header
	// message.goから利用する場合、Headerは作成済
	if k.header == nil {
		h := &Header{}
		err := h.unMarshalBytes(b)
		if err != nil {
			return err
		}
		k.header = h
		return nil
	}
	if len(b) != 0 {
		return NewConvMsgErr(fmt.Sprintf("Keepalive MessageのByte列が不正です: %d", len(b)))
	}
	return nil
}
