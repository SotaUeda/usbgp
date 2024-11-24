package main

import (
	"fmt"
	"log"
	"net"
	"testing"

	"github.com/SotaUeda/usbgp/config"
)

func TestPaseConfig(t *testing.T) {
	_, nw1, _ := net.ParseCIDR("192.0.2.0/24")
	_, nw2, _ := net.ParseCIDR("203.0.113.0/24")
	actConf, _ := config.New(64512, "198.51.100.10", 65413, "198.51.100.20", config.Active, []*net.IPNet{nw1})
	psvConf, _ := config.New(64513, "198.51.100.20", 65412, "198.51.100.10", config.Passive, []*net.IPNet{nw2})
	actConf2, _ := config.New(64512, "198.51.100.10", 65413, "198.51.100.20", config.Active, []*net.IPNet{nw1, nw2})
	actConfNillNW, _ := config.New(64512, "198.51.100.10", 65413, "198.51.100.20", config.Active, nil)
	tests := []struct {
		name  string
		args  string
		want  *config.Config
		isErr bool
	}{
		{name: "none args", args: "", want: nil, isErr: true},
		{name: "invalid las", args: "65536 198.51.100.10 65413 198.51.100.20 active 192.0.2.0/24", want: nil, isErr: true},
		{name: "invalid ras", args: "64512 198.51.100.10 65536 198.51.100.20 active 192.0.2.0/24", want: nil, isErr: true},
		{name: "invalid lip", args: "64512 198.51.100.300 65413 198.51.100.20 active 192.0.2.0/24", want: nil, isErr: true},
		{name: "invalid rip", args: "64512 198.51.100.10 65413 198.51.100.300 active 192.0.2.0/24", want: nil, isErr: true},
		{name: "invalid mode", args: "64512 198.51.100.10 65413 198.51.100.20 foobar 192.0.2.0/24", want: nil, isErr: true},
		{name: "invalid network", args: "64512 198.51.100.10 65413 198.51.100.20 active 2001:db8::", want: nil, isErr: true},
		{name: "nil network", args: "64512 198.51.100.10 65413 198.51.100.20 active", want: actConfNillNW, isErr: false},
		{name: "Active 1 network", args: "64512 198.51.100.10 65413 198.51.100.20 active 192.0.2.0/24", want: actConf, isErr: false},
		{name: "Passive 1 network", args: "64513 198.51.100.20 65412 198.51.100.10 passive 203.0.113.0/24", want: psvConf, isErr: false},
		{name: "Active 2 network", args: "64512 198.51.100.10 65413 198.51.100.20 active 192.0.2.0/24 203.0.113.0/24", want: actConf2, isErr: false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := parseConfig(tc.args)
			if tc.isErr {
				if got != nil || err == nil {
					t.Errorf("%s: want nil, got %v, err %v", tc.name, got, err)
				} else {
					log.Printf("%s: err %v", tc.name, err)
					return
				}
			}
			if !equalConfig(tc.want, got) {
				t.Errorf("%s:  want %v, got %v", tc.name, fmt.Sprintf("%+v", tc.want), fmt.Sprintf("%+v", got))
			}
		})
	}
}

func equalConfig(c1, c2 *config.Config) bool {
	if c1 == nil || c2 == nil {
		return false
	}
	if c1.LocalAS() != c2.LocalAS() {
		return false
	}
	if c1.LocalIP().String() != c2.LocalIP().String() {
		return false
	}
	if c1.RemoteAS() != c2.RemoteAS() {
		return false
	}
	if c1.RemoteIP().String() != c2.RemoteIP().String() {
		return false
	}
	if c1.Mode() != c2.Mode() {
		return false
	}
	if len(c1.Networks()) != len(c2.Networks()) {
		return false
	}
	for i, nw := range c1.Networks() {
		if nw.String() != c2.Networks()[i].String() {
			return false
		}
	}
	return true
}
