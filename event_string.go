// Code generated by "stringer -type=Event event.go"; DO NOT EDIT.

package peer

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[ManualStartEvent-0]
}

const _Event_name = "ManualStartEvent"

var _Event_index = [...]uint8{0, 16}

func (i Event) String() string {
	if i < 0 || i >= Event(len(_Event_index)-1) {
		return "Event(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _Event_name[_Event_index[i]:_Event_index[i+1]]
}
