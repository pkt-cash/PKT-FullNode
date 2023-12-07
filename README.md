PKT-FullNode
====

[![ISC License](http://img.shields.io/badge/license-ISC-blue.svg)](http://Copyfree.org)

PKT-FullNode is the core blockchain engine which supports the PKT blockchain.
It is based on the [pktd](https://github.com/pkt-cash/pktd) codebase which was
a single repository combining blockchain engine, wallet, and lightning daemon.

In order to allow the wallet and lightning code to evolve more quickly, the full
node was branched off and lives here where it will evolve on it's own path.

## Why should I run a FullNode?

A FullNode is *not* needed to own PKT, mine PKT, or have a wallet. You only
need this if you want to:
1. Set up your own mining pool
2. Run your own block explorer
3. Run an ElectrumX instance
4. Query your FullNode to learn things about the PKT blockchain
5. Be a good community member and contribute resources

When you run a PKT-FullNode, you are making the PKT Network more decentralized and
providing a service to wallets which need to talk to your node in order to know if
they have been paid on the blockchain.

## How should I run a FullNode?

FullNodes should be run on *servers*, they need not be housed in big datacenters
but they should have a public IP address or ability to forward a port, and should
be something which will be turned on most of the time.

As far as resources, a FullNode can run on anything as small as a Raspberry Pi,
as long as it is attached to a large (500 GB) SSD and has at least 4GB of
available memory.

### Note about hard drives
The PKT FullNode runs best on an SSD, it *can* run on a spinning disk, but the
latency makes its database operate much more slowly.

## How to install

1. [Install golang](https://go.dev/doc/install) if you have not already.
2. Clone this repository with git
3. Type `./do`
4. You should find `./bin/pktd` now exists

## How to run

Typically you can type `./bin/pktd` and your PKT-FullNode will begin syncing
the PKT blockchain.

## Now what?

1. Make sure you have forwarded port 64764 from the public internet
2. If your node is on a private IP address (e.g. 192.168.X.X) then you can
tell pktd how it is reachable from the internet by using the flag
`--externalip=133.33.33.7` (replacing 133.33.33.7 with whatever your public IP
address is).
3. Check if how close you are to being in sync `./bin/pktctl getinfo` and compare
the "blocks" field with a block explorer like explorer.pkt.cash
4. When you're synced, check on the nodes that are connecting to you with
`./bin/pktctl getpeerinfo | jq -r '.[] | select(.inbound == true) | .addr + "    " + .subver '`
make sure you install `jq` first. Note that nodes with `neutrino` in the name are
wallets.

## Issues and support
For help, check:
* https://pkt.chat
* [Discord](https://discord.gg/bjJutHm9CN) or
* [Telegram](https://t.me/pkt_cash)

## More Info

Check https://docs.pkt.cash for more information.

## License

`pktd` is licensed under the [Copyfree](http://Copyfree.org) ISC License.
