package message

import "fmt"

type Type uint8

//go:generate stringer -type=Type message.go
const (
	Open      Type = 1
	Update    Type = 2
	Keepalive Type = 4
)

func newType(t uint8) (Type, error) {
	if t <= 0 || t > 4 {
		return 0, NewConvMsgErr(fmt.Sprintf("BGPのTypeは1-4が期待されています: %d", t))
	}
	return Type(t), nil
}

type Message interface {
	Type() Type
	marshalBytes() ([]byte, error)
	unMarshalBytes([]byte) error
}

// BGP Message version
type version uint8

func newVersion(v uint8) (version, error) {
	if v > 4 {
		return defaultVersion, NewConvMsgErr(
			fmt.Sprintf("BGPのVersionは1-4が期待されています: %d", v))
	}
	return version(v), nil
}

var defaultVersion = version(4)

// BGP Message HoldTime
type holdtime uint16

func newHoldtime(ht uint16) (holdtime, error) {
	return holdtime(ht), nil
}

var defaultHoldtime = holdtime(0)

// BGP Message

func Marshal(m Message) ([]byte, error) {
	return m.marshalBytes()
}

func UnMarshal(b []byte) (Message, error) {
	hLen := 19
	if len(b) < hLen {
		return nil, NewConvMsgErr(fmt.Sprintf("Byte列が短すぎます: %d", len(b)))
	}
	h := &Header{}
	err := h.unMarshalBytes(b[:hLen])
	if err != nil {
		return nil, err
	}
	switch h.type_ {
	case Open:
		o := &OpenMessage{header: h}
		err := o.unMarshalBytes(b[hLen:])
		if err != nil {
			return nil, err
		}
		return o, nil
	case Update:
		// TODO
		u := &UpdateMessage{header: h}
		err := u.unMarshalBytes(b[hLen:])
		if err != nil {
			return nil, err
		}
		return u, nil
	case Keepalive:
		k := &KeepaliveMessage{header: h}
		err := k.unMarshalBytes(b[hLen:])
		if err != nil {
			return nil, err
		}
		return k, nil
	default:
		return nil, NewConvMsgErr(fmt.Sprintf("未知のMessage Typeです: %d", h.type_))
	}
}
