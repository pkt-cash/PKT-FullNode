#!/bin/bash

function build() {
  cd "${GITHUB_WORKSPACE}" || exit
  /bin/sh -c 'source ./do'

  cd "${GITHUB_WORKSPACE}" || exit
  bash -x ./contrib/macos/build.sh
}
build