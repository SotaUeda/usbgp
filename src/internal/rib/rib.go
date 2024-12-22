package rib

import (
	"fmt"

	"github.com/SotaUeda/usbgp/config"
	"github.com/SotaUeda/usbgp/internal/bgp"
	"github.com/SotaUeda/usbgp/internal/ip"
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

// AdjRIBIn / LocRIB / AdjRIBOutで同じようなデータ構造・処理をもつため、
// 共通の処理はribオブジェクトに実装し、これらの3つの構造体のメンバにribを埋め込む。
//
// RIBEntryは、3つのribを渡りながら処理される。
type rib map[*RIBEntry]Status

// RIB内にentryが存在しなければInsert
func (rib) Insert(ent *RIBEntry) {
	// TODO
}

type LocRIB struct {
	rib
	localAS bgp.ASNumber
}

func NewLocRib(c *config.Config) *LocRIB {
	// TODO
	return &LocRIB{
		rib:     rib{},
		localAS: c.LocalAS(),
	}
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
	// TODO
	return &AdjRIBOut{
		rib: rib{},
	}
}

// LocRIBから必要なルートをインストールする
// この時、Remote AS番号が含まれているルートはインストールしない。
func (ro *AdjRIBOut) Update(lr *LocRIB, c *config.Config) {
	// TODO
}
