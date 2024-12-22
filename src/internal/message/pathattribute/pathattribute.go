package pathattribute

import (
	"fmt"
	"net"

	"github.com/SotaUeda/usbgp/internal/bgp"
)

type PathAttribute interface {
	BytesLen() uint16
	MarshalBytes() ([]byte, error)
}

func NewPathAttributesFromBytes(b []byte) ([]PathAttribute, error) {
	pas := make([]PathAttribute, 0)
	for len(b) > 0 {
		if len(b) < 3 {
			return nil, fmt.Errorf("invalid path attribute length: %d", len(b))
		}
		// Attribute Flags
		af := b[0]
		// Attribute Type Code
		atc := b[1]
		// Attribute Length
		// Attribute FlagsのExtended Length bitが立っているかを判定
		// bitが立っている場合は、Attribute Lengthを表すoctetが2byteで表現される
		al := uint16(b[2])
		if af&0b00010000 != 0 {
			al = uint16(b[2])<<8 + uint16(b[3])
		}

		i := 3
		j := i + int(al)
		if len(b) < j {
			return nil, fmt.Errorf("PathAttributeのByte列が短すぎます length: %v", len(b))
		}
		// Attribute Value
		av := b[3 : 3+al]
		switch AttrType(atc) {
		case ORG:
			if len(av) != 1 {
				return nil, fmt.Errorf("invalid origin length: %d", len(av))
			}
			o, err := NewOrigin(av[0])
			if err != nil {
				return nil, err
			}
			pas = append(pas, o)
		case ASP:
			if len(av) < 3 {
				return nil, fmt.Errorf("invalid AS path length: %d", len(av))
			}
			st := ASPathSegmentType(av[0])
			sl := int(av[1])
			idx := 2
			sv := make([]bgp.ASNumber, sl)
			for i := 0; i < sl; i++ {
				sv[i] = bgp.ASNumber(av[idx])<<8 + bgp.ASNumber(av[idx+1])
				idx += 2
			}
			p, err := NewASPath(st, sv)
			if err != nil {
				return nil, err
			}
			pas = append(pas, p)
		case NHP:
			if len(av) != 4 {
				return nil, fmt.Errorf("invalid next hop length: %d", len(av))
			}
			nh, err := NewNextHop(av)
			if err != nil {
				return nil, err
			}
			pas = append(pas, nh)
		default:
			pas = append(pas, DontKnow(b))
		}
		b = b[3+al:]
	}
	return pas, nil
}

func bytesLen(i uint16) uint16 {
	// flagを表す1byteと、typeを表す1byteを含める
	len := i + 2
	if i > 255 {
		// PathAttributeValueLengthが255以上(1byteで表現できない)の場合は、
		// AttributeLengthを表すoctetが1byteで表せず、2byteで表現する必要がある
		len += 2
	} else {
		len++
	}
	return len
}

type AttrType uint8

//go:generate stringer -type=AttrType pathattribute.go
const (
	ORG AttrType = 1
	ASP AttrType = 2
	NHP AttrType = 3
)

type Origin uint8

//go:generate stringer -type=Origin pathattribute.go
const (
	Igp        Origin = 0
	Egp        Origin = 1
	Incomplete Origin = 2
)

func NewOrigin(o uint8) (Origin, error) {
	switch o {
	case 0:
		return Igp, nil
	case 1:
		return Egp, nil
	case 2:
		return Incomplete, nil
	}
	return 0, fmt.Errorf("invalid Origin: %d", o)
}

func (o Origin) BytesLen() uint16 {
	return bytesLen(1)
}

func (o Origin) MarshalBytes() ([]byte, error) {
	aFlg := 0b01000000 // Attribute Flags
	aTC := ORG         // Attribute Type Code
	al := 1            // Attribute Length
	av := o
	return []byte{byte(aFlg), byte(aTC), byte(al), byte(av)}, nil
}

type ASPath interface {
	PathAttribute
	asMarshalBytes() ([]byte, error)
	SegType() ASPathSegmentType
	SegLen() uint8 // ASの数を返す
	Contains(bgp.ASNumber) bool
}

type ASPathSegmentType uint8

//go:generate stringer -type=ASPathSegmentType pathattribute.go
const (
	ASSegTypeSet      ASPathSegmentType = 1
	ASSegTypeSequence ASPathSegmentType = 2
)

// Type, Length, Valueの合計Octet数を返す
func asByteLen(a ASPath) uint16 {
	// ASSetかASSequenceかを表すoctet + ASの数を表すoctet + ASのbytesの値
	l := uint16(2 * a.SegLen())
	return l + 1 + 1
}

type ASSequence []bgp.ASNumber

func (seq ASSequence) BytesLen() uint16 {
	bl := asByteLen(seq)

	return bytesLen(bl)
}

func (seq ASSequence) SegType() ASPathSegmentType {
	return ASSegTypeSequence
}

func (seq ASSequence) SegLen() uint8 {
	return uint8(len(seq))
}

func (seq ASSequence) asMarshalBytes() ([]byte, error) {
	if len(seq) == 0 {
		return nil, fmt.Errorf("ASPathSegment is empty")
	}

	b := make([]byte, asByteLen(seq))

	b[0] = byte(seq.SegType())
	b[1] = byte(seq.SegLen())
	idx := 2
	for _, as := range seq {
		b[idx] = byte(as >> 8)
		b[idx+1] = byte(as)
		idx += 2
	}
	return b, nil
}

func (seq ASSequence) MarshalBytes() ([]byte, error) {
	af := 0b01000000     // Attribute Flags
	atc := ASP           // Attribute Type Code
	al := asByteLen(seq) // Attribute Length
	av, err := seq.asMarshalBytes()
	if err != nil {
		return nil, err
	}
	b := make([]byte, seq.BytesLen())
	if al > 255 {
		af += 0b00010000
		b[0] = byte(af)
		b[1] = byte(atc)
		b[2] = byte(al >> 8)
		b[3] = byte(al)
		copy(b[4:], av)
	} else {
		b[0] = byte(af)
		b[1] = byte(atc)
		b[2] = byte(al)
		copy(b[3:], av)
	}
	return b, nil
}

func (seq ASSequence) Contains(as bgp.ASNumber) bool {
	for _, a := range seq {
		if a == as {
			return true
		}
	}
	return false
}

type ASSet map[bgp.ASNumber]struct{}

func (set ASSet) BytesLen() uint16 {
	bl := asByteLen(set)

	return bytesLen(bl)
}

func (set ASSet) SegType() ASPathSegmentType {
	return ASSegTypeSet
}

func (set ASSet) SegLen() uint8 {
	return uint8(len(set))
}

func (set ASSet) asMarshalBytes() ([]byte, error) {
	if len(set) == 0 {
		return nil, fmt.Errorf("ASPathSegment is empty")
	}

	b := make([]byte, asByteLen(set))

	b[0] = byte(set.SegType())
	b[1] = byte(set.SegLen())
	idx := 2
	for as := range set {
		b[idx] = byte(as >> 8)
		b[idx+1] = byte(as)
		idx += 2
	}
	return b, nil
}

func (set ASSet) MarshalBytes() ([]byte, error) {
	af := 0b01000000     // Attribute Flags
	atc := ASP           // Attribute Type Code
	al := asByteLen(set) // Attribute Length
	av, err := set.asMarshalBytes()
	if err != nil {
		return nil, err
	}
	b := make([]byte, set.BytesLen())
	if al > 255 {
		af += 0b00010000
		b[0] = byte(af)
		b[1] = byte(atc)
		b[2] = byte(al >> 8)
		b[3] = byte(al)
		copy(b[4:], av)
	} else {
		b[0] = byte(af)
		b[1] = byte(atc)
		b[2] = byte(al)
		copy(b[3:], av)
	}
	return b, nil
}

func (set ASSet) Contains(as bgp.ASNumber) bool {
	_, ok := set[as]
	return ok
}

func NewASPath(t ASPathSegmentType, as []bgp.ASNumber) (ASPath, error) {
	switch t {
	case ASSegTypeSequence:
		return ASSequence(as), nil
	case ASSegTypeSet:
		set := make(map[bgp.ASNumber]struct{})
		for _, a := range as {
			set[a] = struct{}{}
		}
		return ASSet(set), nil
	}
	return nil, fmt.Errorf("invalid ASPathSegmentType: %d", t)
}

type NextHop []byte

func NewNextHop(n []byte) (NextHop, error) {
	if len(n) != 4 {
		return nil, fmt.Errorf("invalid next hop length: %d", len(n))
	}
	return NextHop(n), nil
}

func (n NextHop) BytesLen() uint16 {
	return bytesLen(4)
}

func (n NextHop) Val() net.IP {
	if len(n) == 4 {
		return net.IP(n)
	}
	return nil
}

func (n NextHop) MarshalBytes() ([]byte, error) {
	if len(n) != 4 {
		return nil, fmt.Errorf("invalid next hop length: %d", len(n))
	}
	af := 0b01000000 // Attribute Flags
	atc := NHP       // Attribute Type Code
	al := 4          // Attribute Length
	return []byte{byte(af), byte(atc), byte(al), n[0], n[1], n[2], n[3]}, nil
}

type DontKnow []byte

func (d DontKnow) BytesLen() uint16 {
	return bytesLen(uint16(len(d)))
}

func (d DontKnow) MarshalBytes() ([]byte, error) {
	return d, nil
}
