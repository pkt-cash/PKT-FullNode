#!/bin/bash

function publish() {
  local binary
  binary="${1}"

  local extension
  extension="${2}"

  local checksum
  checksum=$(sha256sum ${binary} | cut -d ' ' -f 1)

  local base_url
  base_url='https://github.com/thierrymarianne/contrib-pktd'

  local upload_url
  upload_url="$(curl \
    -H 'Content-Type: application/octet-stream' \
    -H "Authorization: Bearer ${GITHUB_TOKEN}" \
    "${base_url}"/releases 2>> /dev/null | \
    jq -r '.[] | .upload_url' | \
    head -n1)"
  upload_url=${upload_url/\{?name,label\}/}

  local release_name
  release_name="$(curl \
    -H 'Content-Type: application/octet-stream' \
    -H "Authorization: Bearer ${GITHUB_TOKEN}" \
    "${base_url}"/releases 2>> /dev/null | \
    jq -r '.[] | .tag_name' | \
    head -n1)"

  curl \
    -X POST \
    --data-binary @${binary} \
    -H 'Content-Type: application/octet-stream' \
    -H "Authorization: Bearer ${GITHUB_TOKEN}" \
    "${upload_url}?name=${release_name}${extension}"

  curl \
    -X POST \
    --data "$checksum" \
    -H 'Content-Type: text/plain' \
    -H "Authorization: Bearer ${GITHUB_TOKEN}" \
    "${upload_url}?name=${release_name}${extension}.sha256sum"
}

publish '/home/runner/work/thierrymarianne/contrib-pktd/pktd-linux-0.0.0_custom-1.x86_64.rpm' '.rpm'
publish '/home/runner/work/thierrymarianne/contrib-pktd/pktd-linux_0.0.0-custom_amd64.deb' '.deb'
