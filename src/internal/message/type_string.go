// Code generated by "stringer -type=Type message.go"; DO NOT EDIT.

package message

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[Open-1]
	_ = x[Keepalive-4]
}

const (
	_Type_name_0 = "Open"
	_Type_name_1 = "Keepalive"
)

func (i Type) String() string {
	switch {
	case i == 1:
		return _Type_name_0
	case i == 4:
		return _Type_name_1
	default:
		return "Type(" + strconv.FormatInt(int64(i), 10) + ")"
	}
}
