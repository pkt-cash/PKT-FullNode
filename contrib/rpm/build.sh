#!/bin/bash

#
# This script should be run from the project root
# e.g. ./contrib/rpm/build.sh
#


BINARY_FOLDER=.
RPM_PACKAGE_NAME=pktd

./do
echo "Binary built. Building RPM now."

if which fpm; then
	if which rpmbuild; then
		fpm -n $RPM_PACKAGE_NAME -s dir -t rpm $BINARY_FOLDER
		echo "RPM file built."
	else
		echo "rpmbuild not installed or not reachable"
		exit 1
	fi
else
	echo "fpm not installed or not reachable"
	exit 1
fi



