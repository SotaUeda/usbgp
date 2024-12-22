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
	lp     Peer
	rp     Peer
	wg     sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc
)

func TestMain(m *testing.M) {
	// テスト用にPort番号を変更
	// BGPPort = 1799
	ctx, cancel = context.WithCancel(context.Background())
	defer cancel()
	// テスト用のPeerを作成
	lcfg, err := config.New(
		64512,
		"127.0.0.1",
		65413,
		"127.0.0.2",
		config.Active,
		nil,
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
		nil,
	)
	if err != nil {
		log.Fatal(err)
	}
	rp = *New(rcfg)
	rp.Start()
	defer rp.Idle()
	m.Run()
}

func TestTransitionToEstablishedState(t *testing.T) {
	want := Established
	// 処理停止用のコンテキスト
	t_ctx, t_cancel := context.WithCancel(context.Background())
	t_wg := sync.WaitGroup{}
	// remote peer
	rErrCh := make(chan error)
	t_wg.Add(1)
	go func() {
		defer t_wg.Done()
		for {
			select {
			case <-t_ctx.Done():
				return
			default:
				wg.Add(1)
				err := rp.Next(ctx, &wg)
				if err != nil {
					rErrCh <- fmt.Errorf("failed to handle event: %v", err)
					return
				}
				if rp.State == want {
					log.Printf("remote peer is in %v state.", want)
					return
				}
				time.Sleep(100 * time.Millisecond)
			}
		}
	}()
	time.Sleep(1 * time.Second) // remote peerが遷移するよう、1秒待つ
	// local peer
	lErrCh := make(chan error)
	t_wg.Add(1)
	go func() {
		defer t_wg.Done()
		for {
			select {
			case <-t_ctx.Done():
				return
			default:
				wg.Add(1)
				err := lp.Next(ctx, &wg)
				if err != nil {
					lErrCh <- err
					return
				}
				if lp.State == want {
					log.Printf("local peer is in %v state.", want)
					return
				}
				time.Sleep(100 * time.Millisecond)
			}
		}
	}()

	done := make(chan struct{})
	go func() {
		t_wg.Wait()
		close(done)
	}()

	select {
	case re := <-rErrCh:
		t_cancel()
		if re != nil {
			t.Fatal(re)
		}
	case le := <-lErrCh:
		t_cancel()
		if le != nil {
			t.Fatal(le)
		}
	case <-done:
		t_cancel()
		ls := lp.State
		rs := rp.State
		log.Printf("Done. Local Peer State: %v, Remote Peer State: %v", ls, rs)
	case <-time.After(30 * time.Second):
		ls := lp.State
		rs := rp.State
		t_cancel()
		cancel()
		wg.Wait()
		t.Fatalf("Timeout. Local Peer State: %v, Remote Peer State: %v", ls, rs)
	}
}
