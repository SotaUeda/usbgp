package peer

import (
	"fmt"
	"net"
)

type Config struct {
	localAS  ASNumber
	localIP  net.IP
	remoteAS ASNumber
	remoteIP net.IP
	mode     Mode
}

type Mode int

//go:generate stringer -type=Mode config.go
const (
	Passive Mode = iota
	Active
)

func NewConfig(localAS ASNumber, localIP string, remoteAS ASNumber, remoteIP string, mode Mode) (*Config, error) {
	lIP := net.ParseIP(localIP)
	if lIP == nil {
		return nil, fmt.Errorf("invalid local IP address: %s", localIP)
	}
	rIP := net.ParseIP(remoteIP)
	if rIP == nil {
		return nil, fmt.Errorf("invalid remote IP address: %s", remoteIP)
	}
	return &Config{
		localAS:  localAS,
		localIP:  lIP,
		remoteAS: remoteAS,
		remoteIP: rIP,
		mode:     mode,
	}, nil
}
