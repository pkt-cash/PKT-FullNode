// Copyright (c) 2013-2015 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package btcutil

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	_ "crypto/sha512" // Needed for RegisterHash in init
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"os"
	"time"

	"github.com/pkt-cash/PKT-FullNode/btcutil/er"
)

// NewTLSCertPair returns a new PEM-encoded x.509 certificate pair
// based on a 521-bit ECDSA private key.  The machine's local interface
// addresses and all variants of IPv4 and IPv6 localhost are included as
// valid IP addresses.
func NewTLSCertPair(organization string, validUntil time.Time, extraHosts []string) (cert, key []byte, err er.R) {
	now := time.Now()
	if validUntil.Before(now) {
		return nil, nil, er.New("validUntil would create an already-expired certificate")
	}

	priv, errr := ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	if errr != nil {
		return nil, nil, er.E(errr)
	}

	// end of ASN.1 time
	endOfTime := time.Date(2049, 12, 31, 23, 59, 59, 0, time.UTC)
	if validUntil.After(endOfTime) {
		validUntil = endOfTime
	}

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, errr := rand.Int(rand.Reader, serialNumberLimit)
	if errr != nil {
		return nil, nil, er.Errorf("failed to generate serial number: %s", errr)
	}

	host, errr := os.Hostname()
	if errr != nil {
		return nil, nil, er.E(errr)
	}

	ipAddresses := []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("::1")}
	dnsNames := []string{host}
	if host != "localhost" {
		dnsNames = append(dnsNames, "localhost")
	}

	addIP := func(ipAddr net.IP) {
		for _, ip := range ipAddresses {
			if net.IP.Equal(ip, ipAddr) {
				return
			}
		}
		ipAddresses = append(ipAddresses, ipAddr)
	}
	addHost := func(host string) {
		for _, dnsName := range dnsNames {
			if host == dnsName {
				return
			}
		}
		dnsNames = append(dnsNames, host)
	}

	addrs, err := interfaceAddrs()
	if err != nil {
		return nil, nil, err
	}
	for _, a := range addrs {
		ipAddr, _, err := net.ParseCIDR(a.String())
		if err == nil {
			addIP(ipAddr)
		}
	}

	for _, hostStr := range extraHosts {
		host, _, err := net.SplitHostPort(hostStr)
		if err != nil {
			host = hostStr
		}
		if ip := net.ParseIP(host); ip != nil {
			addIP(ip)
		} else {
			addHost(host)
		}
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{organization},
			CommonName:   host,
		},
		NotBefore: now.Add(-time.Hour * 24),
		NotAfter:  validUntil,

		KeyUsage: x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature |
			x509.KeyUsageCertSign,
		IsCA:                  true, // so can sign self.
		BasicConstraintsValid: true,

		DNSNames:    dnsNames,
		IPAddresses: ipAddresses,
	}

	derBytes, errr := x509.CreateCertificate(rand.Reader, &template,
		&template, &priv.PublicKey, priv)
	if errr != nil {
		return nil, nil, er.Errorf("failed to create certificate: %v", errr)
	}

	certBuf := &bytes.Buffer{}
	errr = pem.Encode(certBuf, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	if errr != nil {
		return nil, nil, er.Errorf("failed to encode certificate: %v", errr)
	}

	keybytes, errr := x509.MarshalECPrivateKey(priv)
	if errr != nil {
		return nil, nil, er.Errorf("failed to marshal private key: %v", errr)
	}

	keyBuf := &bytes.Buffer{}
	errr = pem.Encode(keyBuf, &pem.Block{Type: "EC PRIVATE KEY", Bytes: keybytes})
	if errr != nil {
		return nil, nil, er.Errorf("failed to encode private key: %v", errr)
	}

	return certBuf.Bytes(), keyBuf.Bytes(), nil
}
