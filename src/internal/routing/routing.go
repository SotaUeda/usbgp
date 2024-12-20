package routing

import (
	"fmt"
	"net"

	"github.com/vishvananda/netlink"
)

type IPv4Net struct {
	*net.IPNet
	len uint8
}

func NewIPv4Net(nw *net.IPNet) (*IPv4Net, error) {
	if nw.IP.To4() == nil {
		return nil, fmt.Errorf("IPv4アドレスにのみ対応しています: %v", nw)
	}
	ones, _ := nw.Mask.Size()
	switch {
	case ones == 0:
		return &IPv4Net{
			IPNet: nw,
			len:   1,
		}, nil
	case ones <= 8:
		return &IPv4Net{
			IPNet: nw,
			len:   2,
		}, nil
	case ones <= 16:
		return &IPv4Net{
			IPNet: nw,
			len:   3,
		}, nil
	case ones <= 24:
		return &IPv4Net{
			IPNet: nw,
			len:   4,
		}, nil
	case ones <= 32:
		return &IPv4Net{
			IPNet: nw,
			len:   5,
		}, nil
	default:
		return nil, fmt.Errorf("prefixが不正です: %v", nw)
	}
}

func NewIPv4NetsFromBytes(b []byte) ([]*IPv4Net, error) {
	// TODO
	var nws []*IPv4Net
	for len(b) > 0 {
		ones := int(b[0])
		var len uint8
		switch {
		case ones == 0:
			len = 1
		case ones <= 8:
			len = 2
		case ones <= 16:
			len = 3
		case ones <= 24:
			len = 4
		case ones <= 32:
			len = 5
		default:
			return nil, fmt.Errorf("prefixが不正です: %v", ones)
		}
		n := make([]byte, 4)
		copy(n, b[1:len])
		nw := net.IPNet{
			IP:   net.IPv4(n[0], n[1], n[2], n[3]),
			Mask: net.CIDRMask(ones, 32),
		}
		b = b[len:]
		nnw := &IPv4Net{
			IPNet: &nw,
			len:   len,
		}
		nws = append(nws, nnw)
	}
	return nws, nil
}

func (i *IPv4Net) Len() uint8 {
	return i.len
}

func (nw *IPv4Net) MarshalBytes() ([]byte, error) {
	b := make([]byte, nw.len)
	ones, _ := nw.Mask.Size()
	b[0] = byte(ones)
	copy(b[1:], nw.IP.To4())
	return b, nil
}

type locRib struct{}

func newLocRib() *locRib {
	// TODO
	return &locRib{}
}

func (*locRib) lookupRT(nw *net.IPNet) []*net.IPNet {
	routes, err := netlink.RouteList(nil, netlink.FAMILY_V4)
	if err != nil {
		return nil
	}
	var r []*net.IPNet
	p, _ := nw.Mask.Size()
	for _, route := range routes {
		dp, _ := route.Dst.Mask.Size()
		if nw.IP.Equal(route.Dst.IP) && p == dp {
			r = append(r, route.Dst)
		}
	}
	return r
}
