package main

import (
	"APKAgent/xds"
	"fmt"
	"os"
	"os/signal"
)

const (
	maxRandomInt             int = 999999999
	grpcMaxConcurrentStreams     = 1000000
	address                      = "localhost:18000"
)

func main() {
	fmt.Println("Hello, world from Agent.")
	sig := make(chan os.Signal)
	signal.Notify(sig, os.Interrupt)
	go xds.InitApkMgtClient(address)

OUTER:
	for {
		select {
		case s := <-sig:
			switch s {
			case os.Interrupt:
				break OUTER
			}
		}
	}
}
