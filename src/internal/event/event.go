package event

// / BGPのRFC内 8.1
// / (https://datatracker.ietf.org/doc/html/rfc4271#section-8.1)で
// / 定義されているEventを表す列挙型です。
type Event int

//go:generate stringer -type=Event event.go
const (
	ManualStart Event = iota
	// 正常系しか実装しない本実装では別のEventとして扱う意味がないため、
	// TCPConnectionConfirmedはTcpCrAckedも兼ねている。
	TCPConnectionConfirmed
	BGPOpen
	// BGPのRFC内での表記
	KeepAliveMsg
	// BGPのRFC内での表記
	UpdateMsg
	// StateがEstablishedに遷移したことを表す。
	// 存在するほうが実装が楽なので追加したオリジナルのイベント
	Established

	// LocRIB / AdjRIBOut / AdjRIBIn が変わったときのイベント
	// 存在するほうが実装が楽なので追加
	LocRIBChanged
	AdjRIBOutChanged
	AdjRIBInChanged
)
