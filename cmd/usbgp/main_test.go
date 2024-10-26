package main

import (
	"log"
	"testing"

	"github.com/SotaUeda/usbgp/config"
)

func TestPaseConfig(t *testing.T) {
	actConf, _ := config.NewConfig(64512, "198.51.100.10", 65413, "198.51.100.20", config.Active)
	psvConf, _ := config.NewConfig(64513, "198.51.100.20", 65412, "198.51.100.10", config.Passive)
	tests := []struct {
		name string
		args string
		want *config.Config
	}{
		{name: "none args", args: "", want: nil},
		{name: "invalid las", args: "65536 198.51.100.10 65413 198.51.100.20 active", want: nil},
		{name: "invalid ras", args: "64512 198.51.100.10 65536 198.51.100.20 active", want: nil},
		{name: "invalid lip", args: "64512 198.51.100.300 65413 198.51.100.20 active", want: nil},
		{name: "invalid rip", args: "64512 198.51.100.10 65413 198.51.100.300 active", want: nil},
		{name: "invalid mode", args: "64512 198.51.100.10 65413 198.51.100.20 foobar", want: nil},
		{name: "Active", args: "64512 198.51.100.10 65413 198.51.100.20 active", want: actConf},
		{name: "Passive", args: "64513 198.51.100.20 65412 198.51.100.10 passive", want: psvConf},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := paseConfig(tc.args)
			if tc.want == nil {
				if got != nil || err == nil {
					t.Errorf("%s: want nil, got %v, err %v", tc.name, got, err)
				} else {
					log.Printf("%s: err %v", tc.name, err)
					return
				}
			}
			if tc.want.String() != got.String() {
				t.Errorf("%s:  want %v, got %v", tc.name, tc.want, got)
			}
		})
	}
}
