package peer

import (
	"context"
	"fmt"
	"log"
	"sync"
	"testing"
	"time"

	"github.com/SotaUeda/usbgp/config"
)

// テスト用にPort番号を変更
// BGPPort = 1790
var (
	lp  Peer
	rp  Peer
	wg  sync.WaitGroup
	ctx context.Context
)

func TestMain(m *testing.M) {
	// テスト用のPeerを作成
	lcfg, err := config.New(
		64512,
		"127.0.0.1",
		65413,
		"127.0.0.2",
		config.Active,
	)
	if err != nil {
		log.Fatal(err)
	}
	lp = *New(lcfg)
	rcfg, err := config.New(
		65413,
		"127.0.0.2",
		64512,
		"127.0.0.1",
		config.Passive,
	)
	if err != nil {
		log.Fatal(err)
	}
	rp = *New(rcfg)
	var cancel context.CancelFunc
	ctx, cancel = context.WithCancel(context.Background())
	defer cancel()
	m.Run()
	lp.Idle()
	rp.Idle()
}

func TestTransitionToConnectState(t *testing.T) {
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
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	p := New(cfg)
	p.Start()
	defer p.Idle()
	var wg sync.WaitGroup

	// remote peerを作成
	errCh := make(chan error)
	rCfg, err := config.New(
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
	rP := New(rCfg)
	defer rP.Idle()
	go func() {
		rP.Start()
		wg.Add(1)
		go func() {
			err = rP.Next(ctx, &wg)
			if err != nil {
				errCh <- fmt.Errorf("failed to handle event: %v", err)
				return
			}
			errCh <- nil
		}()
	}()
	time.Sleep(1 * time.Second) // remote peerが遷移するよう、1秒待つ
	wg.Add(1)
	go func() { err = p.Next(ctx, &wg) }()
	wg.Wait()
	if err != nil {
		t.Fatalf("failed to handle event: %v", err)
	}
	if err := <-errCh; err != nil {
		t.Fatal(err)
	}

	got := p.State
	want := Connect
	if got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestTransitionToOpenSentState(t *testing.T) {
	cfg, err := config.New(
		64512,
		"127.0.0.3",
		65413,
		"127.0.0.4",
		config.Active,
	)
	if err != nil {
		t.Fatalf("failed to create config: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	p := New(cfg)
	p.Start()
	defer p.Idle()
	var wg sync.WaitGroup
	// remote peerを作成
	rCfg, err := config.New(
		65413,
		"127.0.0.4",
		64512,
		"127.0.0.3",
		config.Passive,
	)
	if err != nil {
		t.Fatalf("failed to create config: %v", err)
		return
	}
	rP := New(rCfg)
	defer rP.Idle()
	errCh := make(chan error)
	go func() {
		rP.Start()
		go func() {
			for i := 0; i < 2; i++ {
				wg.Add(1)
				err = rP.Next(ctx, &wg)
				if err != nil {
					errCh <- fmt.Errorf("failed to handle event: %v", err)
					return
				}
			}
			errCh <- nil
		}()
	}()
	if err := <-errCh; err != nil {
		t.Fatal(err)
	}
	go func() {
		for i := 0; i < 2; i++ {
			wg.Add(1)
			err = p.Next(ctx, &wg)
			if err != nil {
				errCh <- fmt.Errorf("failed to handle event: %v", err)
			}
		}
		errCh <- nil
	}()
	if err := <-errCh; err != nil {
		t.Fatal(err)
	}
	wg.Wait()
	got := p.State
	want := Connect
	if got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
}
