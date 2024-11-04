package peer

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/SotaUeda/usbgp/config"
	"github.com/SotaUeda/usbgp/internal/message"
)

var BGPPort = 179

// 通信に関する処理を担当する構造体です。
// TCPConnectionを張ったり、
// BGPMessageのデータを送受信します。
type conn struct {
	*net.TCPConn
	buf []byte
}

func newConnect(ctx context.Context, cfg *config.Config) (*conn, error) {
	c := &conn{}
	err := c.connect(ctx, cfg)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (c *conn) connect(ctx context.Context, cfg *config.Config) error {
	switch cfg.Mode() {
	case config.Active:
		// エラーが発生した場合、接続を繰り返す
		// TODO: Timeout、エラーハンドリング
		for {
			select {
			case <-ctx.Done():
				return fmt.Errorf("connection dial canceled")
			case err, ok := <-c.dial(cfg):
				if !ok {
					c.buf = make([]byte, 0, 1500)
					return nil
				}
				log.Printf("connection dial error: %v", err)
				time.Sleep(1 * time.Second)
			}
		}
	case config.Passive:
		select {
		case <-ctx.Done():
			return fmt.Errorf("connection dial canceled")
		case err, ok := <-c.accept(cfg):
			if !ok {
				c.buf = make([]byte, 0, 1500)
				return nil
			}
			return fmt.Errorf("connection accept error: %v", err)
		}
	default:
		return fmt.Errorf("invalid mode: %v", cfg.Mode())
	}
}

func (c *conn) dial(cfg *config.Config) <-chan error {
	ech := make(chan error)
	go func() {
		defer close(ech)
		laddr := &net.TCPAddr{
			IP:   cfg.LocalIP(),
			Port: BGPPort,
		}
		raddr := &net.TCPAddr{
			IP:   cfg.RemoteIP(),
			Port: BGPPort,
		}
		conn, err := net.DialTCP("tcp", laddr, raddr)
		if err != nil {
			ech <- err
			return
		}
		c.TCPConn = conn
	}()
	return ech
}

func (c *conn) accept(cfg *config.Config) <-chan error {
	ech := make(chan error)
	go func() {
		defer close(ech)
		laddr := &net.TCPAddr{
			IP:   cfg.LocalIP(),
			Port: BGPPort,
		}
		listener, err := net.ListenTCP("tcp", laddr)
		if err != nil {
			ech <- err
			return
		}
		conn, err := listener.AcceptTCP()
		if err != nil {
			ech <- err
			return
		}
		c.TCPConn = conn
	}()
	return ech
}

func (c *conn) writeMsg(m message.Message) error {
	b, err := message.Marshal(m)
	if err != nil {
		return err
	}
	_, err = c.Write(b)
	if err != nil {
		return err
	}
	return nil
}
