package rib

import (
	"fmt"
	"net"

	"github.com/SotaUeda/usbgp/config"
	"github.com/SotaUeda/usbgp/internal/bgp"
	"github.com/SotaUeda/usbgp/internal/ip"
	"github.com/SotaUeda/usbgp/internal/message"
	"github.com/SotaUeda/usbgp/internal/message/pathattribute"
	"github.com/vishvananda/netlink"
)

// 各種RIBの処理の際、以前に処理したエントリは再度処理する必要がない。
// その判別のためのステータス
type Status int

//go:generate stringer -type=Status rib.go
const (
	New Status = iota
	UnChanged
)

type RIBEntry struct {
	nw    *ip.IPv4Net
	attrs []pathattribute.PathAttribute
}

func NewRIBEntry(nw *ip.IPv4Net, attrs []pathattribute.PathAttribute) *RIBEntry {
	return &RIBEntry{
		nw:    nw,
		attrs: attrs,
	}
}

func (re *RIBEntry) String() string {
	return fmt.Sprintf("RIBEntry{nw: %s, attrs: %v}", re.nw, re.attrs)
}

func (re *RIBEntry) containAS(as bgp.ASNumber) bool {
	for _, attr := range re.attrs {
		switch a := attr.(type) {
		case pathattribute.ASPath:
			return a.Contains(as)
		}
	}
	return false
}

// AdjRIBIn / LocRIB / AdjRIBOutで同じようなデータ構造・処理をもつため、
// 共通の処理はribオブジェクトに実装し、これらの3つの構造体のメンバにribを埋め込む。
//
// RIBEntryは、3つのribを渡りながら処理される。
type rib map[*RIBEntry]Status

// RIB内にentryが存在しなければInsert
func (r rib) Insert(ent *RIBEntry) {
	if _, ok := r[ent]; !ok {
		r[ent] = New
	}
}

func (r rib) Routes() []*RIBEntry {
	rts := make([]*RIBEntry, 0, len(r))
	for rt := range r {
		rts = append(rts, rt)
	}
	return rts
}

type LocRIB struct {
	rib
	localAS bgp.ASNumber
}

func NewLocRib(c *config.Config) *LocRIB {
	// AS Pathは、ほかのピアから受信したルートと統一的に扱うために、
	// LocRib -> AdjRIBOutにルートを送るときに、自分のAS番号を
	// 追加するので、ここでは空にしておく。
	ap, err := pathattribute.NewASPath(pathattribute.ASSegTypeSequence, []bgp.ASNumber{})
	if err != nil {
		return nil
	}
	pas := []pathattribute.PathAttribute{
		pathattribute.Igp,
		ap,
		pathattribute.NextHop(c.LocalIP().To4()),
	}
	rib := rib{}

	l := &LocRIB{
		rib:     rib,
		localAS: c.LocalAS(),
	}

	for _, nw := range c.Networks() {
		rts := l.LookupRT(nw)
		for _, rt := range rts {
			l.Insert(NewRIBEntry(rt, pas))
		}
	}

	return l
}

func (l *LocRIB) LookupRT(nw *ip.IPv4Net) []*ip.IPv4Net {
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

type AdjRIBOut struct {
	rib
}

func NewAdjRIBOut() *AdjRIBOut {
	return &AdjRIBOut{
		rib: rib{},
	}
}

// LocRIBから必要なルートをインストールする
// この時、Remote AS番号が含まれているルートはインストールしない。
func (ro *AdjRIBOut) Update(lr *LocRIB, c *config.Config) {
	for _, rt := range lr.Routes() {
		if rt.containAS(c.RemoteAS()) {
			continue
		}
		ro.Insert(rt)
	}
}

func (ro *AdjRIBOut) ToUpdateMessage(locIP net.IP, locAS bgp.ASNumber) (*message.UpdateMessage, error) {
	// TODO
	return nil, nil
}
