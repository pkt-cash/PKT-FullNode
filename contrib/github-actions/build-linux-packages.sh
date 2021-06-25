#!/bin/bash

function build() {
  /bin/sh -c 'source ./do'
  bash -x ./contrib/deb/build.sh
  bash -x ./contrib/rpm/build.sh
}
build