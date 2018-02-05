package main

import (
	"log"
	"net"
	"net/url"
	"strings"
	"time"
)

const (
	probeData  = "DAIKIN_UDP/common/basic_info"
	bufferSize = 8192
)

func main() {
	// TODO: shared ipv4/v6 broadcast addr?
	discoveryAddr, err := net.ResolveUDPAddr("udp", "255.255.255.255:30050")
	if err != nil {
		log.Fatalf("couldn't resolve UDP: %v", err)
	}

	listenAddr, err := net.ResolveUDPAddr("udp", ":30000")
	if err != nil {
		log.Fatalf("couldn't resolve local UDP: %v", err)
	}

	listener, err := net.ListenUDP("udp", listenAddr)
	if err != nil {
		log.Fatalf("couldn't listen on UDP: %v", err)
	}
	listener.SetReadBuffer(bufferSize)
	defer listener.Close()

	go func() {
		buf := make([]byte, bufferSize)
		for {
			// TODO: runs forever
			n, addr, err := listener.ReadFromUDP(buf)
			if err != nil {
				log.Fatalf("couldn't read from UDP: %v", err)
			}
			values := parseValues(string(buf[:n]))
			log.Printf("got from=%v values=%+v", addr, values)
		}
	}()

	for {
		log.Printf("sending probe")
		_, err := listener.WriteTo([]byte(probeData), discoveryAddr)
		if err != nil {
			log.Printf("got err sending to probe: %v", err)
		}
		time.Sleep(time.Second * 5)
	}
}

// parseValues parses a string like `abc=123,foo=bar,name=%42%65%64%72%6f%6f%6d` to url.Values.
func parseValues(s string) url.Values {
	out := url.Values{}
	pairs := strings.Split(s, ",")
	for _, pair := range pairs {
		parts := strings.SplitN(pair, "=", 2)
		var value string
		if len(parts) == 2 {
			value, _ = url.QueryUnescape(parts[1])
		}
		out.Add(parts[0], value)
	}
	return out
}
