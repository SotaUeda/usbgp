package peer

import (
	"fmt"
	"log"
)

// BGPのRFCで示されている実装方針
// (https://datatracker.ietf.org/doc/html/rfc4271#section-8)では、
// 1つのPeerを1つのイベント駆動ステートマシンとして実装しています。
// Peer構造体はRFC内で示されている実装方針に従ったイベント駆動ステートマシンです。
type Peer struct {
	State      State
	eventQueue chan Event
	config     *Config
}

func NewPeer(c *Config) *Peer {
	return &Peer{
		// Stateはnil
		eventQueue: make(chan Event),
		config:     c,
	}
}

func (p *Peer) Start() {
	log.Println("peer is started.")
	p.State = Idle
	go func() { p.eventQueue <- ManualStartEvent }()
}

func (p *Peer) Next() error {
	select {
	case ev := <-p.eventQueue:
		log.Printf("event is occured, event=%v.\n", ev)
		p.handleEvent(ev)
	}
	return nil
}

func (p *Peer) handleEvent(ev Event) error {
	switch p.State {
	case Idle:
		if ev == ManualStartEvent {
			p.State = Connect
		}
	default:
		return fmt.Errorf("Peer is not started: %v", p.State)
	}
	return nil
}
