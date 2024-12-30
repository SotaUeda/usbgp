package rib

import (
	"fmt"
	"log"
	"net"
	"sync"

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
	mu    sync.RWMutex
	nw    *ip.IPv4Net
	attrs []pathattribute.PathAttribute
}

func NewRIBEntry(nw *ip.IPv4Net, attrs []pathattribute.PathAttribute) *RIBEntry {
	return &RIBEntry{
		mu:    sync.RWMutex{},
		nw:    nw,
		attrs: attrs,
	}
}

func (re *RIBEntry) String() string {
	re.mu.RLock()
	defer re.mu.RUnlock()
	return fmt.Sprintf("RIBEntry{nw: %s, attrs: %v}", re.nw, re.attrs)
}

func (re *RIBEntry) containAS(as bgp.ASNumber) bool {
	re.mu.RLock()
	defer re.mu.RUnlock()
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
	for e := range r {
		rts = append(rts, e)
	}
	return rts
}

func (r rib) AllUnchanged() {
	for e := range r {
		r[e] = UnChanged
	}
}

func (r rib) ContainNew() bool {
	for _, s := range r {
		if s == New {
			return true
		}
	}
	return false
}

type LocRIB struct {
	rib
	localAS bgp.ASNumber
	mu      sync.RWMutex
}

func NewLocRIB(c *config.Config) (*LocRIB, error) {
	// AS Pathは、ほかのピアから受信したルートと統一的に扱うために、
	// LocRib -> AdjRIBOutにルートを送るときに、自分のAS番号を
	// 追加するので、ここでは空にしておく。
	ap, err := pathattribute.NewASPath(pathattribute.ASSegTypeSequence, []bgp.ASNumber{})
	if err != nil {
		return nil, err
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

	return l, nil
}

func (l *LocRIB) LookupRT(nw *ip.IPv4Net) []*ip.IPv4Net {
	l.mu.RLock()
	defer l.mu.RUnlock()
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

func (l *LocRIB) WriteRT() {
	l.mu.RLock()
	defer l.mu.RUnlock()
	for _, e := range l.Routes() {
		l.writeRT(e)
	}
}
func (l *LocRIB) writeRT(e *RIBEntry) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	for _, a := range e.attrs {
		switch a := a.(type) {
		case pathattribute.NextHop:
			nw := &net.IPNet{
				IP:   e.nw.IP,
				Mask: e.nw.Mask,
			}
			gw := net.IP(a).To4()
			if err := netlink.RouteAdd(&netlink.Route{
				Dst: nw,
				Gw:  gw,
			}); err != nil {
				log.Fatal(err)
			}
		}
	}
}

// AdjRIBInから必要なルートをLocRIBにインストールする
func (l *LocRIB) Update(ri *AdjRIBIn) {
	l.mu.Lock()
	defer l.mu.Unlock()
	la := l.localAS
	for _, rt := range ri.Routes() {
		// 自ASが含まれているルートはインストールしない
		if rt.containAS(la) {
			continue
		}
		l.Insert(rt)
	}
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
	lr.mu.RLock()
	defer lr.mu.RUnlock()
	for _, rt := range lr.Routes() {
		if rt.containAS(c.RemoteAS()) {
			continue
		}
		ro.Insert(rt)
	}
}

// AdjRIBOutからUpadateMessageを生成する
// PathAttributeごとにUpdateMessageが分かれるため、
// []*message.UpdateMessageを戻り値にしている。
func (ro *AdjRIBOut) ToUpdateMessage(locIP net.IP, locAS bgp.ASNumber) ([]*message.UpdateMessage, error) {
	// IPv4のみ対応
	locIP = locIP.To4()
	if locIP == nil {
		return nil, fmt.Errorf("support IPv4 only")
	}
	// 書籍のRustによる実装では、
	// PathAttributeをKeyに、Vec<IPv4Network>をValueのHashMapを使って、
	// 同じPathAttributeのNLRIは同じVec<IPv4Network>にまとめている。
	// ここで同じPathAttributeとされた経路は1つのUpdateMessageにまとめられる。
	hashMap := map[*[]pathattribute.PathAttribute][]*ip.IPv4Net{}
	for _, e := range ro.Routes() {
		e.mu.Lock()
		defer e.mu.Unlock()
		// Hashとしてポインタを使っているが、
		// "同じPathAttribute"とは、PathAttributeの中身が同じであることを指す?
		h := &e.attrs
		if _, ok := hashMap[h]; !ok {
			hashMap[h] = append(hashMap[h], e.nw)
		} else {
			hashMap[h] = []*ip.IPv4Net{e.nw}
		}
	}

	// UpdateMessageを生成する
	var ums []*message.UpdateMessage
	for pas, nws := range hashMap {
		// PathAttributeのうちNexthopまたはASPathを変更する
		// Nexthopはlocal ipに変更、
		// ASPathはlocal asを追加
		for i, p := range *pas {
			switch p := p.(type) {
			case pathattribute.NextHop:
				n, err := pathattribute.NewNextHop(locIP)
				if err != nil {
					return nil, err
				}
				copy(p, n)
			case pathattribute.ASPath:
				a, err := pathattribute.AppendASPath(p, locAS)
				if err != nil {
					return nil, err
				}
				(*pas)[i] = a
			}
		}
		um, err := message.NewUpdateMsg(*pas, nws, nil)
		if err != nil {
			return nil, err
		}
		ums = append(ums, um)
	}
	return ums, nil
}

type AdjRIBIn struct {
	rib
}

func NewAdjRIBIn() *AdjRIBIn {
	return &AdjRIBIn{
		rib: rib{},
	}
}

// UpdateMessageを受信したときに、AdjRIBInを更新する
func (ri *AdjRIBIn) Update(um *message.UpdateMessage) {
	// TODO: withdrawnに対応
	for _, nw := range um.NLRI() {
		// TODO: Pathattributeが同じであれば、同じRIBEntryにまとめなければならない
		// 実装を見直す必要がある？
		ri.Insert(NewRIBEntry(nw, um.PathAttributes()))
	}
}
