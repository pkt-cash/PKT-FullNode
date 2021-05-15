package nat

import (
	"net"
	"sync"
	"time"

	"github.com/jackpal/gateway"
	natpmp "github.com/jackpal/go-nat-pmp"
	"github.com/pkt-cash/pktd/btcutil/er"
)

// Compile-time check to ensure PMP implements the Traversal interface.
var _ Traversal = (*PMP)(nil)

// PMP is a concrete implementation of the Traversal interface that uses the
// NAT-PMP technique.
type PMP struct {
	client *natpmp.Client

	forwardedPortsMtx sync.Mutex
	forwardedPorts    map[uint16]struct{}
}

// DiscoverPMP attempts to scan the local network for a NAT-PMP enabled device
// within the given timeout.
func DiscoverPMP(timeout time.Duration) (*PMP, er.R) {
	// Retrieve the gateway IP address of the local network.
	gatewayIP, err := gateway.DiscoverGateway()
	if err != nil {
		return nil, er.E(err)
	}

	pmp := &PMP{
		client:         natpmp.NewClientWithTimeout(gatewayIP, timeout),
		forwardedPorts: make(map[uint16]struct{}),
	}

	// We'll then attempt to retrieve the external IP address of this
	// device to ensure it is not behind multiple NATs.
	if _, err := pmp.ExternalIP(); err != nil {
		return nil, err
	}

	return pmp, nil
}

// ExternalIP returns the external IP address of the NAT-PMP enabled device.
func (p *PMP) ExternalIP() (net.IP, er.R) {
	res, err := p.client.GetExternalAddress()
	if err != nil {
		return nil, er.E(err)
	}

	ip := net.IP(res.ExternalIPAddress[:])
	if isPrivateIP(ip) {
		return nil, ErrMultipleNAT.Default()
	}

	return ip, nil
}

// AddPortMapping enables port forwarding for the given port.
func (p *PMP) AddPortMapping(port uint16) er.R {
	p.forwardedPortsMtx.Lock()
	defer p.forwardedPortsMtx.Unlock()

	_, err := p.client.AddPortMapping("tcp", int(port), int(port), 0)
	if err != nil {
		return er.E(err)
	}

	p.forwardedPorts[port] = struct{}{}

	return nil
}

// DeletePortMapping disables port forwarding for the given port.
func (p *PMP) DeletePortMapping(port uint16) er.R {
	p.forwardedPortsMtx.Lock()
	defer p.forwardedPortsMtx.Unlock()

	if _, exists := p.forwardedPorts[port]; !exists {
		return er.Errorf("port %d is not being forwarded", port)
	}

	_, err := p.client.AddPortMapping("tcp", int(port), 0, 0)
	if err != nil {
		return er.E(err)
	}

	delete(p.forwardedPorts, port)

	return nil
}

// ForwardedPorts returns a list of ports currently being forwarded.
func (p *PMP) ForwardedPorts() []uint16 {
	p.forwardedPortsMtx.Lock()
	defer p.forwardedPortsMtx.Unlock()

	ports := make([]uint16, 0, len(p.forwardedPorts))
	for port := range p.forwardedPorts {
		ports = append(ports, port)
	}

	return ports
}

// Name returns the name of the specific NAT traversal technique used.
func (p *PMP) Name() string {
	return "NAT-PMP"
}
