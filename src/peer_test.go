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
	BGPPort = 1790
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
		err := rp.Next(ctx, &wg)
		if err != nil {
			rErrCh <- fmt.Errorf("failed to handle event: %v", err)
			return
		}
		rErrCh <- nil
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
	want := OpenSent
	// 処理停止用のコンテキスト
	t_ctx, t_cancel := context.WithCancel(context.Background())
	// remote peer
	rErrCh := make(chan error)
	go func() {
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
					rErrCh <- nil
					return
				}
				time.Sleep(1 * time.Second)
			}
		}
	}()
	time.Sleep(1 * time.Second) // remote peerが遷移するよう、1秒待つ
	// local peer
	lErrCh := make(chan error)
	go func() {
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
					lErrCh <- nil
					return
				}
				time.Sleep(1 * time.Second)
			}
		}
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
	case <-time.After(10 * time.Second):
		t_cancel()
		cancel()
		wg.Wait()
		t.Fatal("timeout")
	}
}

func TestTransitionToOpenConfirmState(t *testing.T) {
	want := OpenConfirm
	// 処理停止用のコンテキスト
	t_ctx, t_cancel := context.WithCancel(context.Background())
	// remote peer
	rErrCh := make(chan error)
	go func() {
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
					rErrCh <- nil
					return
				}
				time.Sleep(1 * time.Second)
			}
		}
	}()
	time.Sleep(1 * time.Second) // remote peerが遷移するよう、1秒待つ
	// local peer
	lErrCh := make(chan error)
	go func() {
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
					lErrCh <- nil
					return
				}
				time.Sleep(1 * time.Second)
			}
		}
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
	case <-time.After(10 * time.Second):
		t_cancel()
		cancel()
		wg.Wait()
		t.Fatal("timeout")
	}
}
