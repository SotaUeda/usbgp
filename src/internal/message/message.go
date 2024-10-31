package message

type Type uint8

//go:generate stringer -type=Type message.go
const (
	Open Type = 1
)

func (t Type) Uint8() uint8 {
	return uint8(t)
}

type Message interface {
	Type() Type
	bytes() ([]byte, error)
	String() string
}

func New(b []byte) (Message, error) {
	// TODO
	return nil, nil
}

func Bytes(Message) ([]byte, error) {
	// TODO
	return nil, nil
}
