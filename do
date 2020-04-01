#!/bin/sh
die() { echo $1; exit 1; }
export GO111MODULE=on
PKTD_GIT_ID=$(git rev-list -1 HEAD | cut -c 1-8)
if ! git diff --quiet; then
    PKTD_GIT_ID="${PKTD_GIT_ID}-dirty"
fi
PKTD_LDFLAGS="-X github.com/pkt-cash/pktd/pktconfig/version.appBuild=${PKTD_GIT_ID}"

echo "Building pktd"
go build -ldflags="${PKTD_LDFLAGS}" || die "failed to build pktd"
echo "Building wallet"
go build -ldflags="${PKTD_LDFLAGS}" -o wallet ./pktwallet || die "failed to build wallet"
echo "Building btcctl"
go build -ldflags="${PKTD_LDFLAGS}" ./cmd/btcctl || die "failed to build btcctl"
echo "Running tests"
go test ./... || die "tests failed"
echo "Everything looks good - use ./pktd to launch"