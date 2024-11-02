package message

import (
	"fmt"
	"net"

	"github.com/SotaUeda/usbgp/internal/bgp"
)

type OpenMessage struct {
	header   *Header
	version  version
	myAS     bgp.ASNumber
	holdtime holdtime // 正常系のみ実装するので、実質的に使用しない
	bgpID    net.IP

	// 使用しないが、受信用に念のため用意
	optsLen uint8
	opts    []byte
}

func (*OpenMessage) Type() Type {
	return Open
}

func NewOpenMsg(as bgp.ASNumber, ip net.IP) (*OpenMessage, error) {
	h, err := newHeader(29, Open)
	if err != nil {
		return nil, err
	}
	ipv4 := ip.To4()
	if ipv4 == nil {
		return nil, NewConvBytesErr(
			fmt.Sprintf("IPv4アドレスにのみ対応しています: %v", ip),
		)
	}
	return &OpenMessage{
		header:   h,
		version:  defaultVersion,
		myAS:     as,
		holdtime: defaultHoldtime,
		bgpID:    ipv4,
		optsLen:  0,
		opts:     []byte{},
	}, nil
}

func (o *OpenMessage) marshalBytes() ([]byte, error) {
	b := make([]byte, 29)
	// Header
	h, err := o.header.marshalBytes()
	if err != nil {
		return nil, err
	}
	copy(b, h)
	// Version
	b[19] = uint8(o.version)
	// My Autonomous System
	as := o.myAS.Uint16()
	b[20] = uint8(as >> 8)
	b[21] = uint8(as)
	// Hold Time
	ht := uint16(o.holdtime)
	b[22] = uint8(ht >> 8)
	b[23] = uint8(ht)
	// BGP Identifier
	id := o.bgpID.To4()
	if id == nil {
		return nil, NewConvBytesErr("BGP Identifierの変換に失敗しました")
	}
	copy(b[24:28], o.bgpID.To4())
	// Optional Parameters Length
	b[28] = o.optsLen
	// Optional Parameters
	if o.optsLen > 0 {
		b = append(b, o.opts...)
	}

	return b, nil
}

func (o *OpenMessage) unMarshalBytes(b []byte) error {
	// OpenMessageの長さは29byte以上
	oLen := 29
	hLen := 19
	p := 0
	// Header
	// message.goから利用する場合、Headerは作成済
	if o.header == nil {
		h := &Header{}
		err := h.unMarshalBytes(b[:hLen])
		if err != nil {
			return err
		}
		o.header = h
		p += hLen
	} else {
		oLen -= hLen
	}
	if len(b) < oLen {
		return NewConvMsgErr("OpenMessageのByte列が短すぎます")
	}
	var err error
	// Version
	o.version, err = newVersion(b[p])
	if err != nil {
		return err
	}
	// My Autonomous System
	o.myAS = bgp.ASNumber(uint16(b[p+1])<<8 | uint16(b[p+2]))
	// Hold Time
	o.holdtime, err = newHoldtime(uint16(b[p+3])<<8 | uint16(b[p+4]))
	if err != nil {
		return err
	}
	// BGP Identifier
	o.bgpID = net.IP(b[p+5 : p+9]).To4()
	if o.bgpID == nil {
		return NewConvMsgErr("BGP Identifierの変換に失敗しました")
	}
	// Optional Parameters Length
	o.optsLen = b[p+9]
	// Optional Parameters
	if o.optsLen > 0 {
		o.opts = b[p+10:]
	} else {
		o.opts = []byte{}
	}

	return nil
}
