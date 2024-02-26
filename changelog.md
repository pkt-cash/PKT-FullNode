# PKT-FullNode v1.1.0

NOT YET RELEASED

This is a major update which introduces the Network Steward Election 2.0 system.
This election system computes the balance and the votes of all addresses, a detailed
description of each part of this new feature is found in `docs/network_steward_vote_2.md`.

# PKT-FullNode v1.0.1

Jan 9, 2024

This is a minor update to the PKT-FullNode which addresses a number of bugs.

1. More checkpoints
  * Switch from a checkpoint every 2**13 to every 2**15 to reduce the number of checkpoints
  * Checkpoint up to block 2260992 (December 2nd, 2023)
2. Improvements in the peer selection logic - this effectively fixes a bug where a node can become
starved for peers and end up with no connections at all, though there are many possible peers in the
peers.json file. The improvements to fix this are detailed as follows:
  * The previous version v1.0 started off only selecting peers which had been validated
  (either had successfully connected before, or were received from the DNS seed).
  * That version also retained an older logic which called GetAddress() in a loop until a suitable
  address was returned. This version has rewritten GetAddress() to accept a filter function which
  tells GetAddress whether the address is valid before that address is returned.
  * In this version, GetAddress() cycles through ALL known addresses, beginning at a random point
  in a random bucket, and iterating through that bucket until it has been exhausted, then proceeding
  to a random point in the next bucket, and so on until all buckets have been exhausted.
  * In this version, GetAddress() will first try addresses which are "trusted", i.e. have been
  connected in the past, or have been received from a DNS seed, but before failing to return an
  address, it will move on to "untrusted" addresses, i.e. addresses which have been received by gossip
  from any node.
  * In the new peering logic, there is a mode switch between "startup mode" and "relaxed mode". In
  relaxed mode, GetAddress() does not prefer "trusted" addresses, it returns any address it knows
  about. In the previous version, the criteria for switching to relaxed mode was that there is 2 less
  than the target number of peers and a sync peer has been identified. In this version, we switch to
  relaxed mode after we have 1/2 of the target number of peers.
3. Bug fixes
  * Improper use of locks in BanScore() makes "getpeerinfo" hang indefinitely
  * Json-iterator cannot serialize a map, causing "getblockchaininfo" to crash
  * "debuglevel" RPC was not implemented.
4. Logging
  * Minor logging improvements and decreasing of noisy logs


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