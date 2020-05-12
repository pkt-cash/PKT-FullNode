#!/bin/bash

#
# This script should be run from the project root
# e.g. ./contrib/rpm/build.sh
#

./do
echo "Binary built. Building RPM now."

mkdir ./bins
mv ./pktd ./bins
mv ./wallet ./bins/pktwallet
mv ./btcctl ./bins/pktctl

if which fpm; then
	if which rpmbuild; then
		fpm -n pktd -s dir -t rpm -v "$(./bins/pktctl --version | sed 's/.* version //' | tr -d '\n')" ./bins
		echo "RPM file built."
	else
		echo "rpmbuild not installed or not reachable"
		exit 1
	fi
else
	echo "fpm not installed or not reachable"
	exit 1
fi



