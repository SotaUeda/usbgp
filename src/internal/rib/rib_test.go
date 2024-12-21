package rib

import (
	"net"
	"testing"

	"github.com/SotaUeda/usbgp/config"
	"github.com/SotaUeda/usbgp/internal/bgp"
	"github.com/SotaUeda/usbgp/internal/ip"
	"github.com/SotaUeda/usbgp/internal/message/pathattribute"
)

func TestLocRIBCanLookupRoutingTable(t *testing.T) {
	// 本テストの値は環境によって異なる。
	// 本実装では開発機、テスト実施機に
	// 10.200.100.0/24 に属するIPが付与されていることを仮定している。
	nw, err := ip.NewIPv4Net(&net.IPNet{
		IP:   net.ParseIP("10.200.100.0"),
		Mask: net.CIDRMask(24, 32),
	})
	if err != nil {
		t.Fatal(err)
	}
	lRib := &locRIB{}
	get := lRib.LookupRT(nw)
	if len(get) == 0 {
		t.Fatal("lookupRT() = 0, want 1")
	}
	want := []*ip.IPv4Net{nw}
	for i, got := range get {
		if (got.String() != want[i].String()) && (got.Len() != want[i].Len()) {
			t.Errorf("lookupRT() = %v, want %v", got, want[i])
		}
	}
}

func TestLocRibToAdjRIBOut(t *testing.T) {
	// 本テストの値は環境によって異なる。
	// 本実装では開発機、テスト実施機に
	// 10.200.100.0/24 に属するIPが付与されていることを仮定している。
	// docker-compose環境のhost2で実行することを仮定している。
	nw := &net.IPNet{
		IP:   net.ParseIP("10.100.220.0"),
		Mask: net.CIDRMask(24, 32),
	}
	c, err := config.New(
		64513,
		"10.200.100.3",
		64512,
		"10.200.100.2",
		config.Passive,
		[]*net.IPNet{nw},
	)
	if err != nil {
		t.Fatal(err)
	}
	lr := NewLocRib(c)
	get := NewAdjRIBOut()
	get.Update(lr, c)

	ipn, err := ip.NewIPv4Net(nw)
	if err != nil {
		t.Error(err)
	}
	asp, err := pathattribute.NewASPath(pathattribute.ASSegTypeSequence, []bgp.ASNumber{someAS, localAS})
	if err != nil {
		t.Error(err)
	}
	pas := []pathattribute.PathAttribute{
		pathattribute.Igp,
		asp,
		pathattribute.NextHop(c.LocalIP().To4()),
	}
	want := NewAdjRIBOut()
	want.insert(NewRibEntry(
		ipn,
		pas,
	))

	if !adjRIBOutEqual(get, want, t) {
		t.Errorf("adjRIBOut not equal: %v, %v", get, want)
	}
}

func adjRIBOutEqual(get, want *AdjRIBOut, t *testing.T) bool {
	return ribEqual(get.rib, want.rib)
}

func ribEqual(get, want *RIB, t *testing.T) bool {
	// get, wantそれぞれのRIBEntry(*IPv4Net, []PathAttribute)とREStatusを比較する必要がある
	// RIBはmap[*RIBEntry]REStatusであるため、mapの順番は保証されないかつポインタのため比較ができない
	// それぞれのRIBEntryを文字列として取得し、それをkeyとして新しいmapを作成し、それを比較する
	getMap := make(map[string]REStatus)
	for k, v := range get {
		getMap[k.String()] = v
	}
	wantMap := make(map[string]REStatus)
	for k, v := range want {
		wantMap[k.String()] = v
	}
	if len(getMap) != len(wantMap) {
		t.Errorf("len(get) = %d, len(want) = %d", len(getMap), len(wantMap))
		return false
	}
	for k, v := range getMap {
		if wantMap[k] != v {
			t.Errorf("get[%s] = %v, want[%s] = %v", k, v, k, wantMap[k])
			return false
		}
	}

	return true
}
