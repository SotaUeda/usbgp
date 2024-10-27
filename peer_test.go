package peer

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/SotaUeda/usbgp/config"
)

func TestTransitionToConnectState(t *testing.T) {
	cfg, err := config.NewConfig(
		64512,
		"127.0.0.1",
		65413,
		"127.0.0.2",
		config.Active,
	)
	if err != nil {
		t.Fatalf("failed to create config: %v", err)
	}
	peer := NewPeer(cfg)
	peer.Start()
	var wg sync.WaitGroup
	// remote peerを作成
	errCh := make(chan error)
	go func() {
		defer close(errCh)
		config, err := config.NewConfig(
			65413,
			"127.0.0.2",
			64512,
			"127.0.0.1",
			config.Passive,
		)
		if err != nil {
			errCh <- fmt.Errorf("failed to create config: %v", err)
			return
		}
		peer := NewPeer(config)
		peer.Start()
		wg.Add(1)
		go func() {
			err = peer.Next(context.Background(), &wg)
			if err != nil {
				errCh <- fmt.Errorf("failed to handle event: %v", err)
				return
			}
		}()
		errCh <- nil
	}()
	wg.Add(1)
	go func() { err = peer.Next(context.Background(), &wg) }()
	wg.Wait()
	if err != nil {
		t.Fatalf("failed to handle event: %v", err)
	}
	if err := <-errCh; err != nil {
		t.Fatal(err)
	}
	got := peer.State
	want := Connect
	if got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
}
