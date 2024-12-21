package msg

import (
	"bytes"
	"net"
	"testing"

	"github.com/SotaUeda/usbgp/internal/bgp"
	"github.com/SotaUeda/usbgp/internal/ip"
	"github.com/SotaUeda/usbgp/internal/msg/pab"
)

func TestHeaderMarshalAndUnmarshal(t *testing.T) {
	h, err := newHeader(19, Open)
	if err != nil {
		t.Error(err)
	}
	b, err := h.marshalBytes()
	if err != nil {
		t.Error(err)
	}
	h2 := &Header{}
	err = h2.unMarshalBytes(b)
	if err != nil {
		t.Error(err)
	}
	if !headerEqual(h, h2) {
		t.Errorf("header not equal: %v, %v", h, h2)
	}
}

func TestOpenMessageMarshalAndUnmarshal(t *testing.T) {
	o, err := NewOpenMsg(
		bgp.ASNumber(64512),
		net.ParseIP("127.0.0.1"),
	)
	if err != nil {
		t.Error(err)
	}
	b, err := Marshal(o)
	if err != nil {
		t.Error(err)
	}
	o2, err := UnMarshal(b)
	if err != nil {
		t.Error(err)
	}
	if !openMsgeEqual(o, o2.(*OpenMessage)) {
		t.Errorf("open message not equal: %v, %v", o, o2)
	}
}

func TestUpdateMessageMarshalAndUnmarshal(t *testing.T) {
	someAS := bgp.ASNumber(64513)

	localAS := bgp.ASNumber(64514)
	localIP := net.ParseIP("10.200.100.3").To4()

	// AttributeはStructにしたほうがよさそう
	asp, err := pab.NewASPath(pab.ASSegTypeSequence, []bgp.ASNumber{someAS, localAS})
	if err != nil {
		t.Error(err)
	}
	pas := []pab.PathAttribute{
		pab.Igp,
		asp,
		pab.NextHop(localIP),
	}

	_, nw, _ := net.ParseCIDR("10.100.220.0/24")
	ipv4nw, err := ip.NewIPv4Net(nw)
	if err != nil {
		t.Error(err)
	}
	u, _ := NewUpdateMsg(
		pas,
		[]*ip.IPv4Net{ipv4nw},
		[]*ip.IPv4Net{},
	)

	b, err := Marshal(u)
	if err != nil {
		t.Error(err)
	}
	u2, err := UnMarshal(b)
	if err != nil {
		t.Error(err)
	}
	if !updateMsgeEqual(u, u2.(*UpdateMessage)) {
		t.Errorf("update message not equal:\n%v\n%v", u, u2)
	}
}

func headerEqual(h1, h2 *Header) bool {
	if h1.len != h2.len {
		return false
	}
	if h1.msgType != h2.msgType {
		return false
	}
	return true
}

func openMsgeEqual(o1, o2 *OpenMessage) bool {
	if !headerEqual(o1.header, o2.header) {
		return false
	}
	if o1.version != o2.version {
		return false
	}
	if o1.myAS != o2.myAS {
		return false
	}
	if o1.holdtime != o2.holdtime {
		return false
	}
	if !net.IP.Equal(o1.bgpID, o2.bgpID) {
		return false
	}
	if o1.optsLen != o2.optsLen {
		return false
	}
	if o1.opts != nil && !bytes.Equal(o1.opts, o2.opts) {
		return false
	}
	return true
}

func updateMsgeEqual(u1, u2 *UpdateMessage) bool {
	if !headerEqual(u1.header, u2.header) {
		return false
	}
	if u1.wrBytesLen != u2.wrBytesLen {
		return false
	}
	if !routeEqual(u1.withdrawnRoutes, u2.withdrawnRoutes) {
		return false
	}
	if u1.pathAttrBytesLen != u2.pathAttrBytesLen {
		return false
	}
	if !pathAttributesEqual(u1.pathAttributes, u2.pathAttributes) {
		return false
	}
	if !routeEqual(u1.nlri, u2.nlri) {
		return false
	}

	return true
}

func routeEqual(r1, r2 []*ip.IPv4Net) bool {
	if len(r1) != len(r2) {
		return false
	}
	for i, n1 := range r1 {
		if !n1.IP.Equal(r2[i].IP) {
			return false
		}
		if !bytes.Equal(n1.Mask, r2[i].Mask) {
			return false
		}
	}
	return true
}

func pathAttributesEqual(pas1, pas2 []pab.PathAttribute) bool {
	if len(pas1) != len(pas2) {
		return false
	}
	for i, pa1 := range pas1 {
		if !pathAttrEqual(pa1, pas2[i]) {
			return false
		}
	}
	return true
}

func pathAttrEqual(pa1, pa2 pab.PathAttribute) bool {
	if pa1.BytesLen() != pa2.BytesLen() {
		return false
	}

	switch pa1.(type) {
	case pab.Origin:
		if pa1 != pa2 {
			return false
		}
		return true
	case pab.ASPath:
		if !asPathEqual(pa1.(pab.ASPath), pa2.(pab.ASPath)) {
			return false
		}
		return true
	case pab.NextHop:
		nh1 := pa1.(pab.NextHop)
		nh2 := pa2.(pab.NextHop)
		if !net.IP.Equal(nh1.Val(), nh2.Val()) {
			return false
		}
		return true
	case pab.DontKnow:
		_, ok := pa2.(pab.DontKnow)
		return ok
	}
	return false
}

func asPathEqual(ap1, ap2 pab.ASPath) bool {
	if ap1.SegType() != ap2.SegType() {
		return false
	}
	if ap1.SegLen() != ap2.SegLen() {
		return false
	}
	switch ap1.(type) {
	case pab.ASSequence:
		seq1 := ap1.(pab.ASSequence)
		seq2 := ap2.(pab.ASSequence)
		for i, as1 := range seq1 {
			if as1 != seq2[i] {
				return false
			}
		}
		return true
	case pab.ASSet:
		set1 := ap1.(pab.ASSet)
		set2 := ap2.(pab.ASSet)
		for as1 := range set1 {
			_, ok := set2[as1]
			return ok
		}
	}
	return false
}
