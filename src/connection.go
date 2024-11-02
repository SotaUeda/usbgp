package peer

import (
	"fmt"
	"net"

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

func newConnect(cfg *config.Config) (*conn, error) {
	c := &conn{}
	err := c.connect(cfg)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (c *conn) connect(cfg *config.Config) error {
	switch cfg.Mode() {
	case config.Active:
		err := c.dial(cfg)
		if err != nil {
			return fmt.Errorf("connection dial error: %v", err)
		}
	case config.Passive:
		err := c.accept(cfg)
		if err != nil {
			return fmt.Errorf("connection accept error: %v", err)
		}
	}
	c.buf = make([]byte, 0, 1500)
	return nil
}

func (c *conn) dial(cfg *config.Config) error {
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
		return err
	}
	c.TCPConn = conn
	return nil
}

func (c *conn) accept(cfg *config.Config) error {
	laddr := &net.TCPAddr{
		IP:   cfg.LocalIP(),
		Port: BGPPort,
	}
	listener, err := net.ListenTCP("tcp", laddr)
	if err != nil {
		return err
	}
	conn, err := listener.AcceptTCP()
	if err != nil {
		return err
	}
	c.TCPConn = conn
	return nil
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
