package message

import (
	"net"
	"testing"

	"github.com/SotaUeda/usbgp/internal/bgp"
	"github.com/SotaUeda/usbgp/internal/ip"
	"github.com/SotaUeda/usbgp/internal/message/pathattribute"
	"github.com/SotaUeda/usbgp/internal/test"
)

func TestUpdateMessageMarshalAndUnmarshal(t *testing.T) {
	someAS := bgp.ASNumber(64513)

	localAS := bgp.ASNumber(64514)
	localIP := net.ParseIP("10.200.100.3").To4()

	ap, err := pathattribute.NewASPath(pathattribute.ASSegTypeSequence, []bgp.ASNumber{someAS, localAS})
	if err != nil {
		t.Error(err)
	}
	pas := []pathattribute.PathAttribute{
		pathattribute.Igp,
		ap,
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

func updateMsgeEqual(u1, u2 *UpdateMessage, t *testing.T) bool {
	if !headerEqual(u1.header, u2.header, t) {
		return false
	}
	if u1.wrBytesLen != u2.wrBytesLen {
		t.Errorf("update message withdrawn routes length not equal: %v, %v", u1.wrBytesLen, u2.wrBytesLen)
		return false
	}
	if !test.RouteEqual(u1.withdrawnRoutes, u2.withdrawnRoutes, t) {
		return false
	}
	if u1.pathAttrBytesLen != u2.pathAttrBytesLen {
		t.Errorf("update message path attributes length not equal: %v, %v", u1.pathAttrBytesLen, u2.pathAttrBytesLen)
		return false
	}
	if !test.PathAttributesEqual(u1.pathAttributes, u2.pathAttributes, t) {
		return false
	}
	if !test.RouteEqual(u1.nlri, u2.nlri, t) {
		return false
	}

	return true
}
