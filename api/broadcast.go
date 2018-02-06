package api

import (
	"fmt"
	"net"
	"net/url"
)

const (
	probeData  = "DAIKIN_UDP/common/basic_info"
	daikinPort = 30050
	bufferSize = 8192
)

var (
	discoveryAddr *net.UDPAddr
)

func init() {
	// TODO: shared ipv4/v6 broadcast addr?
	var err error
	discoveryAddr, err = net.ResolveUDPAddr("udp", fmt.Sprintf("255.255.255.255:%d", daikinPort))
	if err != nil {
		// no network interfaces?
		panic(fmt.Sprintf("couldn't resolve UDP broadcast address: %v", err))
	}
}

// Discover helps discover Daikin devices on the network.
type Discover struct {
	conn *net.UDPConn
}

// Next returns the next Daikin discovered.
func (d *Discover) Next() (string, url.Values, error) {
	buf := make([]byte, bufferSize)
	n, addr, err := d.conn.ReadFromUDP(buf)
	if err != nil {
		return "", nil, err
	}
	values := ParseValues(string(buf[:n]))
	return addr.IP.String(), values, nil
}

// Announce requests devices on the network to reply.
func (d *Discover) Announce() error {
	_, err := d.conn.WriteTo([]byte(probeData), discoveryAddr)
	return err
}

// Close closes the underlying listener.
func (d *Discover) Close() error {
	return d.conn.Close()
}

// NewDiscover builds a new Discover that is listening for device announcements.
func NewDiscover() (*Discover, error) {
	// nb. the Daikin API suggests that this requires us to _listen_ on port 30000, but /shrug
	conn, err := net.ListenUDP("udp", nil)
	if err != nil {
		return nil, err
	}
	conn.SetReadBuffer(bufferSize)

	return &Discover{conn}, nil
}
