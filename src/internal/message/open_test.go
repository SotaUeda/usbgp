package message

import (
	"bytes"
	"net"
	"testing"

	"github.com/SotaUeda/usbgp/internal/bgp"
)

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
