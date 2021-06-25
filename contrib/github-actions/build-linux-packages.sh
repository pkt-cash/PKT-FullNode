#!/bin/bash

function build() {
  cd "${GITHUB_WORKSPACE}" || exit
  source ./do

  cd "${GITHUB_WORKSPACE}" || exit
  bash -x ./contrib/deb/build.sh

  mv "${GITHUB_WORKSPACE}"'/'"${RELEASE_NAME}"'.amd64.deb' \
    "${GITHUB_WORKSPACE}"'/'"${RELEASE_NAME}"'-amd64.deb'

  cd "${GITHUB_WORKSPACE}" || exit
  bash -x ./contrib/rpm/build.sh

  mv "${GITHUB_WORKSPACE}"'/'"${RELEASE_NAME}"'-1.x86_64.rpm' \
    "${GITHUB_WORKSPACE}"'/'"${RELEASE_NAME}"'-x86_64.rpm'
}
build