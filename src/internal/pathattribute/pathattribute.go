package pathattribute

import (
	"net"

	"github.com/SotaUeda/usbgp/internal/bgp"
)

type PathAttribute interface {
	ExLenBit() bool
	AttrType() AttrType
	AttrLen() int16
}

type AttrType int8

//go:generate stringer -type=AttrType pathattribute.go
const (
	ORG AttrType = 1
	ASP AttrType = 2
	NHP AttrType = 3
	UNK AttrType = 4 // 対応していないPathAttribute用
)

type Origin int8

//go:generate stringer -type=Origin pathattribute.go
const (
	Igp       Origin = 0
	Egp       Origin = 1
	Incomlete Origin = 2
)

func (o Origin) ExLenBit() bool {
	return false
}

func (o Origin) AttrType() AttrType {
	return ORG
}

func (o Origin) AttrLen() int16 {
	return 1
}

type ASPathSegmentType int8

//go:generate stringer -type=AsPathSegmentType pathattribute.go
const (
	ASSet ASPathSegmentType = 1
	ASSeq ASPathSegmentType = 2
)

type ASPath interface {
	PathAttribute
	SegType() ASPathSegmentType
	SegLen() int8
}

type ASPathSeq []bgp.ASNumber

func (a ASPathSeq) ExLenBit() bool {
	// TODO
	return false
}

func (a ASPathSeq) AttrType() AttrType {
	return ASP
}

func (a ASPathSeq) AttrLen() int16 {
	// TODO
	return 0
}

func (a ASPathSeq) SegType() ASPathSegmentType {
	return ASSeq
}

func (a ASPathSeq) SegLen() int8 {
	return int8(len(a))
}

type ASPathSet map[bgp.ASNumber]struct{}

func (a ASPathSet) ExLenBit() bool {
	// TODO
	return false
}

func (a ASPathSet) AttrType() AttrType {
	return ASP
}

func (a ASPathSet) AttrLen() int16 {
	// TODO
	return 0
}

func (a ASPathSet) SegType() ASPathSegmentType {
	return ASSet
}

func (a ASPathSet) SegLen() int8 {
	return int8(len(a))
}

func NewASPath(t ASPathSegmentType, as []bgp.ASNumber) ASPath {
	switch t {
	case ASSeq:
		return ASPathSeq(as)
	case ASSet:
		set := make(map[bgp.ASNumber]struct{})
		for _, a := range as {
			set[a] = struct{}{}
		}
		return ASPathSet(set)
	}
	return nil
}

type NextHop []byte

func (n NextHop) ExLenBit() bool {
	return false
}

func (n NextHop) AttrType() AttrType {
	return NHP
}

func (n NextHop) AttrLen() int16 {
	return int16(len(n))
}

func (n NextHop) Val() net.IP {
	if len(n) == 4 {
		return net.IP(n)
	}
	return nil
}

type DontKnow []byte

func (d DontKnow) ExLenBit() bool {
	// TODO
	return false
}

func (d DontKnow) AttrType() AttrType {
	return UNK
}

func (d DontKnow) AttrLen() int16 {
	return int16(len(d))
}
