package peer

import (
	"net"

	"github.com/SotaUeda/usbgp/config"
	"github.com/vishvananda/netlink"
)

type locRib struct{}

func newLocRib(c config.Config) *locRib {
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
