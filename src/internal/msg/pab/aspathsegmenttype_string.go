// Code generated by "stringer -type=ASPathSegmentType pathattribute.go"; DO NOT EDIT.

package pab

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[ASSegTypeSet-1]
	_ = x[ASSegTypeSequence-2]
}

const _ASPathSegmentType_name = "ASSegTypeSetASSegTypeSequence"

var _ASPathSegmentType_index = [...]uint8{0, 12, 29}

func (i ASPathSegmentType) String() string {
	i -= 1
	if i >= ASPathSegmentType(len(_ASPathSegmentType_index)-1) {
		return "ASPathSegmentType(" + strconv.FormatInt(int64(i+1), 10) + ")"
	}
	return _ASPathSegmentType_name[_ASPathSegmentType_index[i]:_ASPathSegmentType_index[i+1]]
}