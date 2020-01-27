#!/bin/sh
die() { echo $1; exit 1; }
export GO111MODULE=on
echo "Building pktd"
go build || die "failed to build pktd"
echo "Building wallet"
go build -o wallet ./pktwallet || die "failed to build wallet"
echo "Building btcctl"
go build ./cmd/btcctl || die "failed to build btcctl"
echo "Running tests"
go test ./... || die "tests failed"
echo "Everything looks good - use ./pktd to launch"