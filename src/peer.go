package peer

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/SotaUeda/usbgp/config"
	"github.com/SotaUeda/usbgp/internal/event"
	"github.com/SotaUeda/usbgp/internal/message"
)

type State int

//go:generate stringer -type=State peer.go
const (
	Idle State = iota
	Connect
	OpenSent
	OpenConfirm
	Established
)

// BGPのRFCで示されている実装方針
// (https://datatracker.ietf.org/doc/html/rfc4271#section-8)では、
// 1つのPeerを1つのイベント駆動ステートマシンとして実装しています。
// Peer構造体はRFC内で示されている実装方針に従ったイベント駆動ステートマシンです。
type Peer struct {
	State      State
	eventQueue chan event.Event
	*conn
	config *config.Config
}

// メッセージの送受信を行うためのChannel
var (
	send chan message.Message
	recv chan message.Message
	ech  chan error
)

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
	p.evEnqueue(event.ManualStart)
}

func (p *Peer) Next(ctx context.Context, wg *sync.WaitGroup) error {
	defer wg.Done()
	select {
	case <-ctx.Done():
		log.Println("Peer Next is done.")
		return p.Idle()
	case ev := <-p.eventQueue:
		log.Printf("event is occured, event=%v.\n", ev)
		if err := p.handleEvent(ctx, ev); err != nil {
			return err
		}
		return nil
	case r := <-recv:
		if err := p.handleMessage(r); err != nil {
			return err
		}
		return nil
	case e := <-ech:
		log.Printf("send/recv error occured: %v\n", e)
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

func (p *Peer) evEnqueue(ev event.Event) {
	go func() { p.eventQueue <- ev }()
}

func (p *Peer) handleEvent(ctx context.Context, ev event.Event) error {
	switch p.State {
	case Idle:
		if ev == event.ManualStart {
			var err error
			p.conn, err = newConnection(ctx, p.config)
			if err != nil {
				return fmt.Errorf("connection error: %v", err)
			}
			if p.conn == nil {
				return fmt.Errorf("TCP Connectionが確立されていません")
			}
			ech = make(chan error)
			send = make(chan message.Message)
			recv = make(chan message.Message)
			// // 送受信のgoroutineを起動
			// // ここが原因？
			// p.sendMsg(ctx, send, ech)
			// p.recvMsg(ctx, recv, ech)
			p.State = Connect
			p.evEnqueue(event.TCPConnectionConfirmed)
		}
	case Connect:
		if ev == event.TCPConnectionConfirmed {
			if p.conn == nil {
				return fmt.Errorf("TCP Conectionが確立されていません")
			}
			om, err := message.NewOpenMsg(
				p.config.LocalAS(),
				p.config.LocalIP(),
			)
			if err != nil {
				return err
			}
			// 送受信のgoroutineを起動
			// ここに移動すると成功する。
			p.sendMsg(ctx, send, ech)
			p.recvMsg(ctx, recv, ech)
			send <- om
			p.State = OpenSent
		}
	case OpenSent:
		if ev == event.BGPOpen {
			if p.conn == nil {
				return fmt.Errorf("TCP Connectionが確立されていません")
			}
			km, err := message.NewKeepaliveMsg()
			if err != nil {
				return err
			}
			send <- km
			p.State = OpenConfirm
		}
	case OpenConfirm:
		if ev == event.KeepAliveMsg {
			p.evEnqueue(event.Established)
			p.State = Established
		}
	}
	return nil
}

func (p *Peer) handleMessage(m message.Message) error {
	switch m.(type) {
	case *message.OpenMessage:
		p.evEnqueue(event.BGPOpen)
	case *message.KeepaliveMessage:
		p.evEnqueue(event.KeepAliveMsg)
	case *message.UpdateMessage:
		p.evEnqueue(event.UpdateMsg)
	}
	return nil
}
