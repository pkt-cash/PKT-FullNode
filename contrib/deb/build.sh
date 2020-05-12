#!/bin/bash

#
# This script should be run from the project root
# e.g. ./contrib/deb/build.sh
#


BINARY_NAME=pkt
DEB_PACKAGE_NAME=pkt
DEB_PACKAGE_DESCRIPTION="PKT"


./do
echo "Binary built. Building DEB now."


if which fpm; then
	fpm -n $DEB_PACKAGE_NAME -s dir -t deb $BINARY_FOLDER
	echo "DEB image built."
else
	echo "fpm not installed or not reachable"
	exit 1
fi

