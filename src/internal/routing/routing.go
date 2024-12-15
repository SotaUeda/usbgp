package routing

import (
	"net"

	"github.com/vishvananda/netlink"
)

type IPv4NetWork struct {
	*net.IPNet
}

func NewIPv4NetWork(nw *net.IPNet) *IPv4NetWork {
	return &IPv4NetWork{nw}
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
