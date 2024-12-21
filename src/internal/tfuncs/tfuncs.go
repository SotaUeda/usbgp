package tfuncs

import (
	"bytes"
	"net"
	"testing"

	"github.com/SotaUeda/usbgp/internal/ip"
	"github.com/SotaUeda/usbgp/internal/message/pathattribute"
)

func RouteEqual(r1, r2 []*ip.IPv4Net, t *testing.T) bool {
	if len(r1) != len(r2) {
		t.Errorf("len(r1) = %d, len(r2) = %d", len(r1), len(r2))
		return false
	}
	for i, n1 := range r1 {
		if !n1.IP.Equal(r2[i].IP) {
			t.Errorf("r1[%d].IP = %v, r2[%d].IP = %v", i, n1.IP, i, r2[i].IP)
			return false
		}
		if !bytes.Equal(n1.Mask, r2[i].Mask) {
			t.Errorf("r1[%d].Mask = %v, r2[%d].Mask = %v", i, n1.Mask, i, r2[i].Mask)
			return false
		}
	}
	return true
}

func PathAttributesEqual(pas1, pas2 []pathattribute.PathAttribute, t *testing.T) bool {
	if len(pas1) != len(pas2) {
		t.Errorf("len(pas1) = %d, len(pas2) = %d", len(pas1), len(pas2))
		return false
	}
	for i, pa1 := range pas1 {
		if !PathAttrEqual(pa1, pas2[i], t) {
			t.Errorf("pas1[%d] = %v, pas2[%d] = %v", i, pa1, i, pas2[i])
			return false
		}
	}
	return true
}

func PathAttrEqual(pa1, pa2 pathattribute.PathAttribute, t *testing.T) bool {
	if pa1.BytesLen() != pa2.BytesLen() {
		t.Errorf("pa1.BytesLen() = %d, pa2.BytesLen() = %d", pa1.BytesLen(), pa2.BytesLen())
		return false
	}

	switch pa1.(type) {
	case pathattribute.Origin:
		if pa1 != pa2 {
			t.Errorf("pa1 = %v, pa2 = %v", pa1, pa2)
			return false
		}
		return true
	case pathattribute.ASPath:
		if !ASPathEqual(pa1.(pathattribute.ASPath), pa2.(pathattribute.ASPath), t) {
			return false
		}
		return true
	case pathattribute.NextHop:
		nh1 := pa1.(pathattribute.NextHop)
		nh2 := pa2.(pathattribute.NextHop)
		if !net.IP.Equal(nh1.Val(), nh2.Val()) {
			t.Errorf("nh1 = %v, nh2 = %v", nh1, nh2)
			return false
		}
		return true
	case pathattribute.DontKnow:
		_, ok := pa2.(pathattribute.DontKnow)
		if !ok {
			t.Errorf("pa1 = %v, pa2 = %v", pa1, pa2)
		}
		return ok
	}
	t.Errorf("invalid PathAttribute type: %v", pa1)
	return false
}

func ASPathEqual(ap1, ap2 pathattribute.ASPath, t *testing.T) bool {
	if ap1.SegType() != ap2.SegType() {
		t.Errorf("ap1.SegType() = %d, ap2.SegType() = %d", ap1.SegType(), ap2.SegType())
		return false
	}
	if ap1.SegLen() != ap2.SegLen() {
		t.Errorf("ap1.SegLen() = %d, ap2.SegLen() = %d", ap1.SegLen(), ap2.SegLen())
		return false
	}
	switch ap1.(type) {
	case pathattribute.ASSequence:
		seq1 := ap1.(pathattribute.ASSequence)
		seq2 := ap2.(pathattribute.ASSequence)
		for i, as1 := range seq1 {
			if as1 != seq2[i] {
				t.Errorf("seq1[%d] = %v, seq2[%d] = %v", i, as1, i, seq2[i])
				return false
			}
		}
		return true
	case pathattribute.ASSet:
		set1 := ap1.(pathattribute.ASSet)
		set2 := ap2.(pathattribute.ASSet)
		for as1 := range set1 {
			_, ok := set2[as1]
			if !ok {
				t.Errorf("set1 = %v, set2 = %v", set1, set2)
			}
			return ok
		}
	}
	t.Errorf("invalid ASPath type: %v", ap1)
	return false
}
