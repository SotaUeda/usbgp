package peer

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/SotaUeda/usbgp/config"
)

func TestTransitionToConnectState(t *testing.T) {
	// テスト用にPort番号を変更
	// BGPPort = 1790
	cfg, err := config.New(
		64512,
		"127.0.0.1",
		65413,
		"127.0.0.2",
		config.Active,
	)
	if err != nil {
		t.Fatalf("failed to create config: %v", err)
	}
	peer := New(cfg)
	peer.Start()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var wg sync.WaitGroup
	// remote peerを作成
	errCh := make(chan error)
	go func() {
		defer close(errCh)
		config, err := config.New(
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
		peer := New(config)
		peer.Start()
		wg.Add(1)
		go func() {
			err = peer.Next(ctx, &wg)
			if err != nil {
				errCh <- fmt.Errorf("failed to handle event: %v", err)
				return
			}
		}()
		errCh <- nil
	}()
	if err := <-errCh; err != nil {
		t.Fatal(err)
	}
	time.Sleep(1 * time.Second) // remote_peerが起動するのを待つため
	wg.Add(1)
	go func() { err = peer.Next(ctx, &wg) }()
	wg.Wait()
	if err != nil {
		t.Fatalf("failed to handle event: %v", err)
	}
	got := peer.State
	want := Connect
	if got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
}
