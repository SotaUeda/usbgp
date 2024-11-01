package bgp

import (
	"fmt"
	"strconv"
)

type ASNumber uint16

func ParseASNumber(s string) (ASNumber, error) {
	as, err := strconv.ParseUint(s, 10, 16)
	if err != nil {
		return 0, fmt.Errorf("invalid AS number: %s", s)
	}
	return ASNumber(as), nil
}

func (a ASNumber) Uint16() uint16 {
	return uint16(a)
}
