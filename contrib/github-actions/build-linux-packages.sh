#!/bin/bash

function build() {
  cd "${GITHUB_WORKSPACE}" || exit
  source ./do

  cd "${GITHUB_WORKSPACE}" || exit
  bash -x ./contrib/deb/build.sh

  local RELEASE_VERSION
  RELEASE_VERSION="${RELEASE_NAME}"

  local RELEASE_NAME
  RELEASE_NAME=$(echo "${RELEASE_NAME}" | sed -e 's/v//')

  mv "${GITHUB_WORKSPACE}"'/'"${RELEASE_NAME}"'_amd64.deb' \
    "${GITHUB_WORKSPACE}"'/'"${RELEASE_VERSION}"'-amd64.deb'

  cd "${GITHUB_WORKSPACE}" || exit
  bash -x ./contrib/rpm/build.sh

  mv "${GITHUB_WORKSPACE}"'/'"${RELEASE_NAME}"'-1.x86_64.rpm' \
    "${GITHUB_WORKSPACE}"'/'"${RELEASE_VERSION}"'-x86_64.rpm'
}
build