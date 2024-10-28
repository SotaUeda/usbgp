package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	peer "github.com/SotaUeda/usbgp"
	"github.com/SotaUeda/usbgp/bgp"
	"github.com/SotaUeda/usbgp/config"
)

func main() {
	cStrs := []string{
		"64512 127.0.0.1 65413 127.0.0.2 active",
	}
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	// Ctrl+Cでキャンセル
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("Received an interrupt, stopping...")
		cancel()
	}()

	for _, cStr := range cStrs {
		c, err := paseConfig(cStr)
		if err != nil {
			log.Fatal(err)
		}
		p := peer.NewPeer(c)
		p.Start()
		go func() {
			for {
				wg.Add(1)
				if err := p.Next(ctx, &wg); err != nil {
					log.Fatal(err)
				}
				time.Sleep(100 * time.Millisecond)
			}
		}()
	}
	<-ctx.Done()
	wg.Wait()
	log.Println("usbgp is done.")
	os.Exit(0)
}

func paseConfig(s string) (*config.Config, error) {
	cStrs := strings.Split(s, " ")
	cLen := 5
	if len(cStrs) != cLen {
		return nil, fmt.Errorf("invalid config string: %s, length: %v", s, len(cStrs))
	}

	lAS, err := stringToASNumber(cStrs[0])
	if err != nil {
		return nil, fmt.Errorf("cannot parse 1st part of config, %v, as as-number and config is %v",
			cStrs[0], s)
	}
	lIP := net.ParseIP(cStrs[1])
	if lIP == nil {
		return nil, fmt.Errorf("cannot parse 2nd part of config, %v, as IP address and config is %v",
			cStrs[1], s)
	}
	rAS, err := stringToASNumber(cStrs[2])
	if err != nil {
		return nil, fmt.Errorf("cannot parse 3rd part of config, %v, as as-number and config is %v",
			cStrs[2], s)
	}
	rIP := net.ParseIP(cStrs[3])
	if rIP == nil {
		return nil, fmt.Errorf("cannot parse 4th part of config, %v, as IP address and config is %v",
			cStrs[3], s)
	}
	mode, err := stringToMode(cStrs[4])
	if err != nil {
		return nil, fmt.Errorf("cannot parse 5th part of config, %v, as mode and config is %v",
			cStrs[4], s)
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
