package message

import (
	"bytes"
	"net"
	"testing"

	"github.com/SotaUeda/usbgp/internal/bgp"
	"github.com/SotaUeda/usbgp/internal/ip"
	"github.com/SotaUeda/usbgp/internal/message/pathattribute"
	"github.com/SotaUeda/usbgp/internal/tfuncs"
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
	if !headerEqual(h, h2, t) {
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
	if !openMsgeEqual(o, o2.(*OpenMessage), t) {
		t.Errorf("open message not equal: %v, %v", o, o2)
	}
}

func TestUpdateMessageMarshalAndUnmarshal(t *testing.T) {
	someAS := bgp.ASNumber(64513)

	localAS := bgp.ASNumber(64514)
	localIP := net.ParseIP("10.200.100.3").To4()

	asp, err := pathattribute.NewASPath(pathattribute.ASSegTypeSequence, []bgp.ASNumber{someAS, localAS})
	if err != nil {
		t.Error(err)
	}
	pas := []pathattribute.PathAttribute{
		pathattribute.Igp,
		asp,
		pathattribute.NextHop(localIP),
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
	if !updateMsgeEqual(u, u2.(*UpdateMessage), t) {
		t.Errorf("update message not equal:\n%v\n%v", u, u2)
	}
}

func headerEqual(h1, h2 *Header, t *testing.T) bool {
	if h1.len != h2.len {
		t.Errorf("header length not equal: %v, %v", h1.len, h2.len)
		return false
	}
	if h1.msgType != h2.msgType {
		t.Errorf("header message type not equal: %v, %v", h1.msgType, h2.msgType)
		return false
	}
	return true
}

func openMsgeEqual(o1, o2 *OpenMessage, t *testing.T) bool {
	if !headerEqual(o1.header, o2.header, t) {
		return false
	}
	if o1.version != o2.version {
		t.Errorf("open message version not equal: %v, %v", o1.version, o2.version)
		return false
	}
	if o1.myAS != o2.myAS {
		t.Errorf("open message myAS not equal: %v, %v", o1.myAS, o2.myAS)
		return false
	}
	if o1.holdtime != o2.holdtime {
		t.Errorf("open message holdtime not equal: %v, %v", o1.holdtime, o2.holdtime)
		return false
	}
	if !net.IP.Equal(o1.bgpID, o2.bgpID) {
		t.Errorf("open message bgpID not equal: %v, %v", o1.bgpID, o2.bgpID)
		return false
	}
	if o1.optsLen != o2.optsLen {
		t.Errorf("open message optsLen not equal: %v, %v", o1.optsLen, o2.optsLen)
		return false
	}
	if o1.opts != nil && !bytes.Equal(o1.opts, o2.opts) {
		t.Errorf("open message opts not equal: %v, %v", o1.opts, o2.opts)
		return false
	}
	return true
}

func updateMsgeEqual(u1, u2 *UpdateMessage, t *testing.T) bool {
	if !headerEqual(u1.header, u2.header, t) {
		return false
	}
	if u1.wrBytesLen != u2.wrBytesLen {
		t.Errorf("update message withdrawn routes length not equal: %v, %v", u1.wrBytesLen, u2.wrBytesLen)
		return false
	}
	if !tfuncs.RouteEqual(u1.withdrawnRoutes, u2.withdrawnRoutes, t) {
		return false
	}
	if u1.pathAttrBytesLen != u2.pathAttrBytesLen {
		t.Errorf("update message path attributes length not equal: %v, %v", u1.pathAttrBytesLen, u2.pathAttrBytesLen)
		return false
	}
	if !tfuncs.PathAttributesEqual(u1.pathAttributes, u2.pathAttributes, t) {
		return false
	}
	if !tfuncs.RouteEqual(u1.nlri, u2.nlri, t) {
		return false
	}

	return true
}
