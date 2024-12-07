package message

import (
	"bytes"
	"net"
	"testing"

	"github.com/SotaUeda/usbgp/internal/bgp"
)

func headerEqual(h1, h2 *Header) bool {
	if h1.len != h2.len {
		return false
	}
	if h1.type_ != h2.type_ {
		return false
	}
	return true
}

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


func TestUpdateMessageMarshalAndUnmarshal(t *testing.T) {
	u, err := NewUpdateMsg(
		[]*bgp.PathAttribute{
			bgp.NewPathAttributeOrigin(bgp.IGP),
			bgp.NewPathAttributeAsPath([]bgp.ASPathSegment{
				bgp.NewASPathSegment(bgp.AS_SEQUENCE, []uint16{64512, 64513}),
			}),
			bgp.NewPathAttributeNextHop(net.ParseIP("