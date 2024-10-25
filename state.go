package peer

type State int

//go:generate stringer -type=State state.go
const (
	Idle State = iota
	Connect
)
