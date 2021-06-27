#!/bin/bash

function build() {
  cd "${GITHUB_WORKSPACE}" || exit
  source ./do

  cd "${GITHUB_WORKSPACE}" || exit
  bash -x ./contrib/deb/build.sh

  local VERSION
  VERSION=$(echo "${RELEASE_NAME}" | sed -E 's/.+v//')

  mv -v "${GITHUB_WORKSPACE}"'/pktd-linux_'"${VERSION}"'_amd64.deb' \
    "${GITHUB_WORKSPACE}"'/'"${RELEASE_NAME}"'-linux-amd64.deb'

  cd "${GITHUB_WORKSPACE}" || exit
  bash -x ./contrib/rpm/build.sh

  mv -v "${GITHUB_WORKSPACE}"'/pktd-linux-'"${VERSION}"'-1.x86_64.rpm' \
    "${GITHUB_WORKSPACE}"'/'"${RELEASE_NAME}"'-linux-x86_64.rpm'
}
build