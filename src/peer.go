package peer

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/SotaUeda/usbgp/config"
	"github.com/SotaUeda/usbgp/internal/event"
)

type State int

//go:generate stringer -type=State peer.go
const (
	Idle State = iota
	Connect
	OpenSent
)

// BGPのRFCで示されている実装方針
// (https://datatracker.ietf.org/doc/html/rfc4271#section-8)では、
// 1つのPeerを1つのイベント駆動ステートマシンとして実装しています。
// Peer構造体はRFC内で示されている実装方針に従ったイベント駆動ステートマシンです。
type Peer struct {
	State      State
	eventQueue chan event.Event
	conn       *conn
	config     *config.Config
}

func New(c *config.Config) *Peer {
	return &Peer{
		// Stateはnil
		eventQueue: make(chan event.Event),
		conn:       nil,
		config:     c,
	}
}

func (p *Peer) Start() {
	log.Println("peer is started.")
	p.State = Idle
	go func() { p.eventQueue <- event.ManualStart }()
}

func (p *Peer) Next(ctx context.Context, wg *sync.WaitGroup) error {
	defer wg.Done()
	select {
	case <-ctx.Done():
		log.Println("Peer Next is done.")
		return p.Idle()
	case ev := <-p.eventQueue:
		log.Printf("event is occured, event=%v.\n", ev)
		if err := p.handleEvent(ev); err != nil {
			return err
		}
		return nil
	}
}

func (p *Peer) Idle() error {
	if p.conn != nil {
		if err := p.conn.Close(); err != nil {
			return err
		}
	}
	p.State = Idle
	return nil
}

func (p *Peer) handleEvent(ev event.Event) error {
	switch p.State {
	case Idle:
		if ev == event.ManualStart {
			var err error
			p.conn, err = newConnect(p.config)
			if err != nil {
				return fmt.Errorf("connection error: %v", err)
			}
			go func() { p.eventQueue <- event.TCPConnectionConfirmed }()
			p.State = Connect
		}
	case Connect:
		if ev == event.TCPConnectionConfirmed {
			p.State = OpenSent
		}
	}
	return nil
}
