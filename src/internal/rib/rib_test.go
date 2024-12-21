package rib

import (
	"net"
	"testing"

	"github.com/SotaUeda/usbgp/internal/ip"
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
	lRib := NewLocRib()
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
