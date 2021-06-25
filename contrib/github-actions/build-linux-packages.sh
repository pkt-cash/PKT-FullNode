#!/bin/bash

function build() {
  cd "${GITHUB_WORKSPACE}" || exit
  /bin/sh -c 'source ./do'

  cd "${GITHUB_WORKSPACE}" || exit
  bash -x ./contrib/deb/build.sh

  cd "${GITHUB_WORKSPACE}" || exit
  bash -x ./contrib/rpm/build.sh
}
build