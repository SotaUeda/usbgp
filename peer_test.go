package peer

import (
	"testing"

	"github.com/SotaUeda/usbgp/config"
)

func TestTransitionToConnectState(t *testing.T) {
	config, err := config.NewConfig(
		64512,
		"127.0.0.1",
		65413,
		"127.0.0.2",
		config.Active,
	)
	if err != nil {
		t.Fatalf("failed to create config: %v", err)
	}
	peer := NewPeer(config)
	peer.Start()
	err = peer.Next()
	if err != nil {
		t.Fatalf("failed to handle event: %v", err)
	}
	got := peer.State
	want := Connect
	if got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
}
