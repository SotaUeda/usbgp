package event

// / BGPのRFC内 8.1
// / (https://datatracker.ietf.org/doc/html/rfc4271#section-8.1)で
// / 定義されているEventを表す列挙型です。
type Event int

//go:generate stringer -type=Event event.go
const (
	ManualStart Event = iota
)
