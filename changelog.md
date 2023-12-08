# PKT-FullNode v1.0

Dec 8, 2023

This is a fork from [pktd](https://github.com/pkt-cash/pktd/blob/pktd-v1.6.1/CHANGELOG.md), a
single repository with the full node, wallet, and lightning daemon all in one. The purpose of this
branch-off is to isolate changes to the full node as these must go through a higher level of
verification as any change which affects the consensus logic may cause an accidental hard-fork of
the chain.

This project was started from pktd-v1.5.1 and received the following specific updates that were
back-ported from pktd-v1.6.1 and other branches.

## Security
* Applied security update from v1.6.1 which prevents mutated PacketCrypt Proof DoS attack
See: https://github.com/pkt-cash/pktd/blob/pktd-v1.6.1/CHANGELOG.md for details.

## General updates
* Removed all files which are not necessary in order to build the pktd full node
* Update to golang v1.18

## Improvements to ban logic
* Reimplement ban manager to ban IP addresses rather than banning nodes. Banning nodes is ineffective because the node can simply disconnect and reconnect again.
* Ban nodes if they repidly connect and disconnect too quickly.

## Improvements to peer connection logic
* Added new seed nodes (seed.pkt.pink and seed.pkt.ai)
* Collect local addresses and determine whether different remote addresses are reachable, for example, do not
attempt to connect to a cjdns address unless you have cjdns, or Yggdrasil address, or IPv6 address, etc.
* On startup, do not trust IP addresses received from other peers, only trust those which come from DNS seeds
or those which we have already actually connected to before. Once we have at least 12 outgoing connections, we
switch to "relaxed mode" and connect to any IP address we find, so as to test the ones that we learn about from
other nodes.
* Do not tell other nodes about an IP we learned about until we have actually connected to it at least once.
* Set default maximum peers to 2048 which is commonly used in practice.