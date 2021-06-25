#!/bin/bash

function build() {
  /bin/sh -c 'source ./do'
  bash -x ./contrib/macos/build.sh
}
build