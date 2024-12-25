package ip

import (
	"fmt"
	"net"
)

type IPv4Net struct {
	*net.IPNet
	len uint8
}

func NewIPv4Net(nw *net.IPNet) (*IPv4Net, error) {
	nw.IP = nw.IP.To4()
	if nw.IP == nil {
		return nil, fmt.Errorf("IPv4アドレスにのみ対応しています: %v", nw)
	}
	ipv4nw := &IPv4Net{
		IPNet: nw,
	}
	ones, _ := nw.Mask.Size()
	switch {
	case ones == 0:
		ipv4nw.len = 1
	case ones <= 8:
		ipv4nw.len = 2
	case ones <= 16:
		ipv4nw.len = 3
	case ones <= 24:
		ipv4nw.len = 4
	case ones <= 32:
		ipv4nw.len = 5
	default:
		return nil, fmt.Errorf("prefixが不正です: %v", nw)
	}
	return ipv4nw, nil
}

func NewIPv4NetsFromBytes(b []byte) ([]*IPv4Net, error) {
	var nws []*IPv4Net
	for len(b) > 0 {
		ones := int(b[0])
		var len uint8
		switch {
		case ones == 0:
			len = 1
		case ones <= 8:
			len = 2
		case ones <= 16:
			len = 3
		case ones <= 24:
			len = 4
		case ones <= 32:
			len = 5
		default:
			return nil, fmt.Errorf("prefixが不正です: %v", ones)
		}
		n := make([]byte, 4)
		copy(n, b[1:len])
		nw := net.IPNet{
			IP:   net.IPv4(n[0], n[1], n[2], n[3]),
			Mask: net.CIDRMask(ones, 32),
		}
		b = b[len:]
		nnw := &IPv4Net{
			IPNet: &nw,
			len:   len,
		}
		nws = append(nws, nnw)
	}
	return nws, nil
}

func (i *IPv4Net) Len() uint8 {
	return i.len
}

func (nw *IPv4Net) MarshalBytes() ([]byte, error) {
	b := make([]byte, nw.len)
	ones, _ := nw.Mask.Size()
	b[0] = byte(ones)
	copy(b[1:], nw.IP.To4())
	return b, nil
}

func (nw *IPv4Net) String() string {
	return fmt.Sprintf("%s/%d len:%d", nw.IPNet.IP, nw.IPNet.Mask, nw.len)
}
