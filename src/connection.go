package peer

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"time"

	"github.com/SotaUeda/usbgp/config"
	"github.com/SotaUeda/usbgp/internal/msg"
)

var BGPPort = 179

// 通信に関する処理を担当する構造体です。
// TCPConnectionを張ったり、
// BGPMessageのデータを送受信します。
type conn struct {
	*net.TCPConn
	buf []byte
}

func newConnection(ctx context.Context, cfg *config.Config) (*conn, error) {
	c := &conn{}
	c.buf = make([]byte, 0, 1500)
	err := c.connect(ctx, cfg)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (c *conn) connect(
	ctx context.Context,
	cfg *config.Config,
) error {
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
		log.Printf("dial connected to %v", conn.RemoteAddr())
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
		log.Printf("accept connected from %v", conn.RemoteAddr())
		c.TCPConn = conn
	}()
	return ech
}

func (c *conn) sendMsg(
	ctx context.Context,
	mch chan msg.Message,
	ech chan error,
) {
	go c.send(ctx, mch, ech)
}

func (c *conn) send(
	ctx context.Context,
	mch chan msg.Message,
	ech chan error,
) {
	for {
		select {
		case <-ctx.Done():
			return
		case m := <-mch:
			err := c.writeMsg(m)
			if err != nil {
				ech <- err
			}
		}
	}
}

// メッセージの送信
func (c *conn) writeMsg(m msg.Message) error {
	b, err := msg.Marshal(m)
	if err != nil {
		return err
	}
	n, err := c.Write(b)
	if err != nil {
		return err
	}
	log.Printf("sent message: %v, byte: %v. to %v", m.Type(), n, c.RemoteAddr())
	return nil
}

func (c *conn) recvMsg(
	ctx context.Context,
	mch chan msg.Message,
	ech chan error,
) {
	go c.recv(ctx, mch, ech)
}

func (c *conn) recv(
	ctx context.Context,
	mch chan msg.Message,
	ech chan error,
) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			m, err := c.readMsg()
			if err != nil {
				ech <- err
			}
			if m != nil {
				mch <- m
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
}

// メッセージの受信
// BGP Messageを1つ以上受信している場合は
// 最も古く受信したMessageを返す。
// BGP Messageを受信中（途中）あるいは
// 何も受信していない場合はnilを返す。
func (c *conn) readMsg() (msg.Message, error) {
	t := make([]byte, 1500)
	n, err := c.Read(t)
	switch {
	case n == 0 || err == io.EOF:
		return nil, nil
	case err != nil:
		return nil, err
	}
	c.buf = append(c.buf, t[:n]...)
	b, err := c.splitMsg()
	if err != nil {
		return nil, err
	}
	// 1つのBGP Messageを表すbyteが揃っていない場合
	if b == nil {
		return nil, nil
	}
	m, err := msg.UnMarshal(b)
	if err != nil {
		return nil, err
	}
	log.Printf("received message: %v. from %v", m.Type(), c.RemoteAddr())
	return m, nil
}

// bufから1つ分のBGP Messageを表す[]byteを切り出す
// 1つのBGP Messageを表すbyteが揃っていない場合はnilを返す
func (c *conn) splitMsg() ([]byte, error) {
	// 1つのBGP Messageを表すbyteが揃っている場合
	ml := msgLen(c.buf)
	if ml == 0 || len(c.buf) < ml {
		return nil, nil
	}
	b := c.buf[:ml]
	c.buf = c.buf[ml:]
	return b, nil
}

// []byteのうちどこまでが1つのBGP Messageを表すbyteであるか、整数を返す
// HeaderのLengthフィールドを参照して、BGP Messageの長さを取得する
func msgLen(b []byte) int {
	if len(b) < 19 {
		// 19byte未満の場合は1つのBGP Messageを表すbyteが揃っていない
		log.Printf("MessageのSeparateorを表すデータまでbufferに入っていません。"+
			"データの受信が半端であることが想定されます。 buffer: %v", len(b))
		return 0
	}
	return int(b[16])<<8 | int(b[17])
}
