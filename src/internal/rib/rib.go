package rib

import (
	"github.com/SotaUeda/usbgp/internal/ip"
	"github.com/vishvananda/netlink"
)

type locRIB struct{}

func NewLocRib() *locRIB {
	// TODO
	return &locRIB{}
}

func (*locRIB) LookupRT(nw *ip.IPv4Net) []*ip.IPv4Net {
	routes, err := netlink.RouteList(nil, netlink.FAMILY_V4)
	if err != nil {
		return nil
	}
	var r []*ip.IPv4Net
	p, _ := nw.Mask.Size()
	for _, route := range routes {
		dp, _ := route.Dst.Mask.Size()
		if nw.IP.Equal(route.Dst.IP) && p == dp {
			dst, err := ip.NewIPv4Net(route.Dst)
			if err != nil {
				continue
			}
			r = append(r, dst)
		}
	}
	return r
}
