package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	peer "github.com/SotaUeda/usbgp"
	"github.com/SotaUeda/usbgp/config"
	"github.com/SotaUeda/usbgp/internal/bgp"
)

func main() {
	cStrs := os.Args[1:]
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
		c, err := parseConfig(cStr)
		if err != nil {
			log.Fatal(err)
		}
		p := peer.New(c)
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

func parseConfig(s string) (*config.Config, error) {
	cStrs := strings.Split(s, " ")
	cMinLen := 5
	if len(cStrs) < cMinLen {
		return nil, fmt.Errorf("invalid config string: %s, length: %v", s, len(cStrs))
	}

	lAS, err := bgp.ParseASNumber(cStrs[0])
	if err != nil {
		return nil, fmt.Errorf("cannot parse 1st part of config, %v, as as-number and config is %v",
			cStrs[0], s)
	}
	lIP := net.ParseIP(cStrs[1])
	if lIP == nil {
		return nil, fmt.Errorf("cannot parse 2nd part of config, %v, as IP address and config is %v",
			cStrs[1], s)
	}
	rAS, err := bgp.ParseASNumber(cStrs[2])
	if err != nil {
		return nil, fmt.Errorf("cannot parse 3rd part of config, %v, as as-number and config is %v",
			cStrs[2], s)
	}
	rIP := net.ParseIP(cStrs[3])
	if rIP == nil {
		return nil, fmt.Errorf("cannot parse 4th part of config, %v, as IP address and config is %v",
			cStrs[3], s)
	}
	mode, err := config.ParseMode(cStrs[4])
	if err != nil {
		return nil, fmt.Errorf("cannot parse 5th part of config, %v, as mode and config is %v",
			cStrs[4], s)
	}

	if len(cStrs) == cMinLen {
		return config.New(lAS, lIP.String(), rAS, rIP.String(), mode, nil)
	}

	nws := []*net.IPNet{}
	for _, nw := range cStrs[5:] {
		_, nw, err := net.ParseCIDR(nw)
		if err != nil {
			return nil, fmt.Errorf("cannot parse %v as network and config is %v", nw, s)
		}
		nws = append(nws, nw)
	}
	return config.New(lAS, lIP.String(), rAS, rIP.String(), mode, nws)
}
