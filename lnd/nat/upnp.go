package nat

import (
	"context"
	"net"
	"sync"

	upnp "github.com/NebulousLabs/go-upnp"
	"github.com/pkt-cash/pktd/btcutil/er"
)

// Compile-time check to ensure UPnP implements the Traversal interface.
var _ Traversal = (*UPnP)(nil)

// UPnP is a concrete implementation of the Traversal interface that uses the
// UPnP technique.
type UPnP struct {
	device *upnp.IGD

	forwardedPortsMtx sync.Mutex
	forwardedPorts    map[uint16]struct{}
}

// DiscoverUPnP scans the local network for a UPnP enabled device.
func DiscoverUPnP(ctx context.Context) (*UPnP, er.R) {
	// Scan the local network for a UPnP-enabled device.
	device, err := upnp.DiscoverCtx(ctx)
	if err != nil {
		return nil, er.E(err)
	}

	u := &UPnP{
		device:         device,
		forwardedPorts: make(map[uint16]struct{}),
	}

	// We'll then attempt to retrieve the external IP address of this
	// device to ensure it is not behind multiple NATs.
	if _, err := u.ExternalIP(); err != nil {
		return nil, err
	}

	return u, nil
}

// ExternalIP returns the external IP address of the UPnP enabled device.
func (u *UPnP) ExternalIP() (net.IP, er.R) {
	ip, err := u.device.ExternalIP()
	if err != nil {
		return nil, er.E(err)
	}

	if isPrivateIP(net.ParseIP(ip)) {
		return nil, ErrMultipleNAT.Default()
	}

	return net.ParseIP(ip), nil
}

// AddPortMapping enables port forwarding for the given port.
func (u *UPnP) AddPortMapping(port uint16) er.R {
	u.forwardedPortsMtx.Lock()
	defer u.forwardedPortsMtx.Unlock()

	if err := u.device.Forward(port, ""); err != nil {
		return er.E(err)
	}

	u.forwardedPorts[port] = struct{}{}

	return nil
}

// DeletePortMapping disables port forwarding for the given port.
func (u *UPnP) DeletePortMapping(port uint16) er.R {
	u.forwardedPortsMtx.Lock()
	defer u.forwardedPortsMtx.Unlock()

	if _, exists := u.forwardedPorts[port]; !exists {
		return er.Errorf("port %d is not being forwarded", port)
	}

	if err := u.device.Clear(port); err != nil {
		return er.E(err)
	}

	delete(u.forwardedPorts, port)

	return nil
}

// ForwardedPorts returns a list of ports currently being forwarded.
func (u *UPnP) ForwardedPorts() []uint16 {
	u.forwardedPortsMtx.Lock()
	defer u.forwardedPortsMtx.Unlock()

	ports := make([]uint16, 0, len(u.forwardedPorts))
	for port := range u.forwardedPorts {
		ports = append(ports, port)
	}

	return ports
}

// Name returns the name of the specific NAT traversal technique used.
func (u *UPnP) Name() string {
	return "UPnP"
}
