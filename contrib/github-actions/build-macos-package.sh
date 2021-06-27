#!/bin/bash

function build() {
  cd "${GITHUB_WORKSPACE}" || exit
  source ./do

  cd "${GITHUB_WORKSPACE}" || exit
  bash -x ./contrib/macos/build.sh

  local VERSION
  VERSION=$(echo "${RELEASE_NAME}" | sed -E 's/.+-v//')

  mv -v "${GITHUB_WORKSPACE}"'/pktd-mac-'"${VERSION}"'.pkg' \
    "${GITHUB_WORKSPACE}"'/'"${RELEASE_NAME}"'-mac.pkg'
}
build