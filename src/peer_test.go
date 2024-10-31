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

var (
	lp  Peer
	rp  Peer
	wg  sync.WaitGroup
	ctx context.Context
)

func TestMain(m *testing.M) {
	// テスト用にPort番号を変更
	// BGPPort = 1790
	var cancel context.CancelFunc
	ctx, cancel = context.WithCancel(context.Background())
	defer cancel()
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
	lp.Start()
	defer lp.Idle()
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
	rp.Start()
	defer rp.Idle()
	m.Run()
}

func TestTransitionToConnectState(t *testing.T) {
	// remote peer
	rErrCh := make(chan error)
	wg.Add(1)
	go func() {
		go func() {
			err := rp.Next(ctx, &wg)
			if err != nil {
				rErrCh <- fmt.Errorf("failed to handle event: %v", err)
				return
			}
			rErrCh <- nil
		}()
	}()
	time.Sleep(1 * time.Second) // remote peerが遷移するよう、1秒待つ
	// local peer
	lErrCh := make(chan error)
	wg.Add(1)
	go func() {
		err := lp.Next(ctx, &wg)
		if err != nil {
			lErrCh <- fmt.Errorf("failed to handle event: %v", err)
			return
		}
		lErrCh <- nil
	}()
	wg.Wait()
	if err := <-rErrCh; err != nil {
		t.Fatal(err)
	}
	if err := <-lErrCh; err != nil {
		t.Fatalf("failed to handle event: %v", err)
	}

	got := lp.State
	want := Connect
	if got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestTransitionToOpenSentState(t *testing.T) {
	// remote peer
	rErrCh := make(chan error)
	wg.Add(1)
	go func() {
		go func() {
			err := rp.Next(ctx, &wg)
			if err != nil {
				rErrCh <- fmt.Errorf("failed to handle event: %v", err)
				return
			}
			rErrCh <- nil
		}()
	}()
	time.Sleep(1 * time.Second) // remote peerが遷移するよう、1秒待つ
	// local peer
	lErrCh := make(chan error)
	wg.Add(1)
	go func() {
		err := lp.Next(ctx, &wg)
		if err != nil {
			lErrCh <- fmt.Errorf("failed to handle event: %v", err)
			return
		}
		lErrCh <- nil
	}()
	wg.Wait()
	if err := <-rErrCh; err != nil {
		t.Fatal(err)
	}
	if err := <-lErrCh; err != nil {
		t.Fatalf("failed to handle event: %v", err)
	}

	got := lp.State
	want := OpenSent
	if got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
}
