package config

import (
	"fmt"
	"net"

	"github.com/SotaUeda/usbgp/bgp"
)

type Config struct {
	localAS  bgp.ASNumber
	localIP  net.IP
	remoteAS bgp.ASNumber
	remoteIP net.IP
	mode     Mode
}

type Mode int

//go:generate stringer -type=Mode config.go
const (
	Passive Mode = iota
	Active
)

func ParseMode(s string) (Mode, error) {
	switch s {
	case "passive", "PASSIVE", "Passive":
		return Passive, nil
	case "active", "ACTIVE", "Active":
		return Active, nil
	default:
		return 0, fmt.Errorf("invalid mode: %s", s)
	}
}

func New(
	localAS bgp.ASNumber, localIP string,
	remoteAS bgp.ASNumber, remoteIP string,
	mode Mode,
) (*Config, error) {
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

func (c *Config) LocalAS() bgp.ASNumber {
	return c.localAS
}

func (c *Config) LocalIP() net.IP {
	return c.localIP
}

func (c *Config) RemoteAS() bgp.ASNumber {
	return c.remoteAS
}

func (c *Config) RemoteIP() net.IP {
	return c.remoteIP
}

func (c *Config) Mode() Mode {
	return c.mode
}
