#!/bin/bash

#
# This script should be run from the project root
# e.g. ./contrib/gem/build.sh
#

./do
echo "Binary built. Building GEM now."

mkdir ./bins
mv ./pktd ./bins
mv ./wallet ./bins/pktwallet
mv ./btcctl ./bins/pktctl

if which fpm; then
	if which pkgbuild; then
		fpm -n pktd -s dir -t osxpkg -v "$(./bins/pktctl --version | sed 's/.* version //' | tr -d '\n')" ./bins
		echo "GEM file built."
	else
		echo "pkgbuild not installed or not reachable"
		exit 1
	fi
else
	echo "fpm not installed or not reachable"
	exit 1
fi



