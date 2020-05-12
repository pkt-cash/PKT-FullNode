#!/bin/bash

#
# This script should be run from the project root
# e.g. ./contrib/deb/build.sh
#
set -e

./do
echo "Binary built. Building DEB now."

mkdir ./bins
mv ./pktd ./bins
mv ./wallet ./bins/pktwallet
mv ./btcctl ./bins/pktctl
fpm -n pkt -s dir -t deb -v "$(./bins/pktctl --version | sed 's/.* version //' | tr -d '\n')" ./bins
echo "DEB image built."