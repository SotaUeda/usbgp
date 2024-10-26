package main

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/SotaUeda/usbgp/bgp"
	"github.com/SotaUeda/usbgp/config"
)

func main() {
	// TODO
}

func paseConfig(s string) (*config.Config, error) {
	cstrs := strings.Split(s, " ")
	clen := 5
	if len(cstrs) != clen {
		return nil, fmt.Errorf("invalid config string: %s, length: %v", s, len(cstrs))
	}

	lAS, err := stringToASNumber(cstrs[0])
	if err != nil {
		return nil, fmt.Errorf("cannot parse 1st part of config, %v, as as-number and config is %v",
			cstrs[0], s)
	}
	lIP := net.ParseIP(cstrs[1])
	if lIP == nil {
		return nil, fmt.Errorf("cannot parse 2nd part of config, %v, as IP address and config is %v",
			cstrs[1], s)
	}
	rAS, err := stringToASNumber(cstrs[2])
	if err != nil {
		return nil, fmt.Errorf("cannot parse 3rd part of config, %v, as as-number and config is %v",
			cstrs[2], s)
	}
	rIP := net.ParseIP(cstrs[3])
	if rIP == nil {
		return nil, fmt.Errorf("cannot parse 4th part of config, %v, as IP address and config is %v",
			cstrs[3], s)
	}
	mode, err := stringToMode(cstrs[4])
	if err != nil {
		return nil, fmt.Errorf("cannot parse 5th part of config, %v, as mode and config is %v",
			cstrs[4], s)
	}

	return config.NewConfig(lAS, lIP.String(), rAS, rIP.String(), mode)
}

func stringToASNumber(s string) (bgp.ASNumber, error) {
	as, err := strconv.ParseUint(s, 10, 16)
	if err != nil {
		return 0, fmt.Errorf("invalid AS number: %s", s)
	}
	return bgp.ASNumber(as), nil
}

func stringToMode(s string) (config.Mode, error) {
	switch s {
	case "passive", "PASSIVE", "Passive":
		return config.Passive, nil
	case "active", "ACTIVE", "Active":
		return config.Active, nil
	default:
		return 0, fmt.Errorf("invalid mode: %s", s)
	}
}
