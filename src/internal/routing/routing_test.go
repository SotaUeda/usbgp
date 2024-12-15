package routing

import (
	"net"
	"testing"
)

func TestLocRibCanLookupRoutingTable(t *testing.T) {
	// 本テストの値は環境によって異なる。
	// 本実装では開発機、テスト実施機に
	// 10.200.100.0/24 に属するIPが付与されていることを仮定している。
	nw := net.IPNet{
		IP:   net.ParseIP("10.200.100.0"),
		Mask: net.CIDRMask(24, 32),
	}
	lRib := newLocRib()
	get := lRib.lookupRT(&nw)
	if len(get) == 0 {
		t.Fatal("lookupRT() = 0, want 1")
	}
	want := []*net.IPNet{&nw}
	for i, got := range get {
		if got.String() != want[i].String() {
			t.Errorf("lookupRT() = %v, want %v", got, want[i])
		}
	}
}
