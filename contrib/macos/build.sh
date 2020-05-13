#!/bin/bash

#
# This script should be run from the project root
# e.g. ./contrib/gem/build.sh
#
if which fpm; then
	if which pkgbuild; then
		fpm -n pktd-mac-amd64 -s dir -t osxpkg -v "$(./bin/pktctl --version | sed 's/.* version //' | tr -d '\n')" ./bin
		echo "GEM file built."
	else
		echo "pkgbuild not installed or not reachable"
		exit 1
	fi
else
	echo "fpm not installed or not reachable"
	exit 1
fi



