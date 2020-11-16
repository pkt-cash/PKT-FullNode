pktd
====

[![ISC License](http://img.shields.io/badge/license-ISC-blue.svg)](http://Copyfree.org)

`pktd` is the full node *PKT Cash* implementation, written in Go (golang).

This project is currently under active development and considered 
to be beta quality software.

In particular, the development branch of `pktd` is highly experimental, 
and only be very carefully, if at all, operated in a production
environment or pn the PKT Cash mainnet.

`pktd` is the primary mainnet node - It is known to correctly download,
validate, and serve the PKT Cash blockchain, using the rule for block
acceptance based on Bitcoin Core, with the addition of PacketCrypt Proofs. 

It relays newly mined blocks, and individual transactions that have not yet
made it into a block, as well as maintaining a transaction pool. All
individual transactions admitted to the pool follow the rules defined by the
PKT Cash blockchain, which includes strict checks which filter transactions
based on miner requirements ("standard" vs "non-standard" transactions).

Unlike other similar software, `pktd` does *NOT* directly include wallet
functionality - this was an intentional design decision.  You will not be
able to make or receive payments directly with `pktd` directly.

Example wallet functionality will be provided by the included, separate,
[pktwallet](https://github.com/pkt-cash/pktd/pktwallet) package.

## Requirements

* [Go](http://golang.org) 1.14 or later.
* A somewhat recent release of Git.

## Issue Tracker

* The GitHub [integrated GitHub issue tracker](https://github.com/pkt-cash/pktd/issues)
is used for this project.  

## Building

Using `git`, clone the project from the repository:

`git clone https://github.com/pkt-cash/pktd`

Use the `./do` shell script to build `pktd`, `pktwallet`, and `pktctl`.

NOTE: It is highly recommended to use only the toolchain Google distributes
at the [official Golang homepage](https://golang.org/dl). Go provided by a 
Linux distribution very often uses different defaults and applies non-standard
patches against the official sources, often to meet specific distributions
requirements (for example, Red Hat backports security fixes, as well as
providing a different default linker configuration vs. the upstream Google
Golang.)

Support can only be provided for binaries compiled from unmodified released
compilers, using the upstream Golang toolchain, and the official PKT Cash 
source code. We simply cannot test and support every distribution specific
toolchain combination. The official Golang Linux installer always available 
for download [here](https://storage.googleapis.com/golang/getgo/installer_linux).

## Documentation

The documentation for `pktd` is work-in-progress, and is located in the [docs](https://github.com/pkt-cash/pktd/tree/master/docs) folder.

## License

`pktd` is licensed under the [Copyfree](http://Copyfree.org) ISC License.
