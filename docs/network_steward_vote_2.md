# Network Steward Voting 2.0
Before getting in to the details of how the voting system works, we need to describe some basic policy
rules and lay down definitions for terms which are used throughout this document. Votes are cast by
creating a special type of transaction output, they can be cast at any time and will apply in the next
vote count. Because addresses are free to create, votes are weighted based on the balance of the address
at the time of voting, this is why we must track both votes and balances.

* **epoch**: An epoch is a time period from election to the next. An election happens after the end of
each epoch and all votes and addresss balances are counted. The epoch is defined as `10,080`, a
consensus rule which makes the votes get counted every 7 days.
* **vote table hash**: A hash which is computed as a checksum during each vote counting operation.

## New Flags
* `--dropvotes` -> This flag causes the `votebalance` and `votewinnerdb` tables to be deleted and the
node to exit immediately when that is done. These tables will be re-created by a re-index job on next
start.

## New RPCs:
* `getaddressinfo` -> Get the balance of an address and who they are voting for
* `listaddresses` -> List balance and voting info for all addresses (paginated)
* `getwinners` -> List the winners of the election plus the vote table hash, in reverse order.
The hash of the balance info is important for validating that there is no discrepancy between nodes.

## Vote structure
In order to cast a vote, you must make a transaction paying **zero** PKT to a special script:

`OP_RETURN <vote_data>`

The `<vote data>` is made up of two components:

* is_candidate -> 1 byte, `0x00` or `0x01` depending on whether you want to be a candidate.
* vote_for -> n bytes, the pkScript that you want to vote for.

The pkScript is exactly what would go in the TxOut when making a payment to that address, so for a
standard segwit address such as `pkt1qtu5y64ln9n9mw3e88ydt28jgmu327ygme8y54t`, that takes the form
`OP_0 OP_PUSH20(5f284d57f32ccbb74727391ab51e48df22af111b)` or `00145f284d57f32ccbb74727391ab51e48df22af111b`.
While this might seem strange, any wallet that is capable of making a transaction is already generating
these scripts, so for a wallet to make a vote, they need only create the script as though they were
going to pay the address, and then instead of paying it, they prepend 1 byte to indicate whether they
also wish to be a candidate and then place take the result and push it after an `OP_RETURN`.

The final result looks like this: `6a170100145f284d57f32ccbb74727391ab51e48df22af111b`.

```
                     6a 17 01 00145f284d57f32ccbb74727391ab51e48df22af111b
OP_RETURN -----------^^ ^^ ^^ ^^^^^^^^^^^^
OP_PUSH23 --------------^^ ^^ ^^^^^^^^^^
This voter is candidating -^^ ^^^^^^^^
Script they are voting for ---^^^^^^^
```

### Important notes about voting
1. Votes must pay ZERO pkt to this output, if the output is paid something then the vote is disregarded
2. Votes must use inputs from ONE address only, votes if there are multiple addresses paying the
transaction which votes, only one of them will be considered.
3. Votes must indicate `0x00` (not a candidate) or `0x01` (candidate). Any other prefix is invalid and
will not be counted.

## Vote db tables

### votebalance
This db contains both votes and balance entries. For each address there is one balance entry, and there
are votes going back far enough to sustain a rollback as far as one week.

All entries are prefixed with the relevant address in pkScript form, following this is 4 bytes: For balance
entries, all 4 bytes are zero, for vote entries they are the big endian representation of the number of the
block wherein the vote took place.

The value in the votebalance table depends on whether the entry is a vote or a balance:

* Vote Value: `[is_candidate: byte][vote_for_address: bytes]`
  * `is_candidate` is a 1 byte flag set to either zero or one depending on whether the address has signaled
  it's intention to candidate for the role of Network Steward.
  * `vote_for_address` is a variable number of bytes which represents who the address has voted for in
  pkScript format. If this is zero bytes then the address is not voting.
* Balance Value: `Array<[blockn: int32][balance: int64]>` - The balance value is a concatnated array of
fixed length balance entries. Each entry is 12 bytes long, starting with the block number where the balance
is effective and followed by the balance itself in atomic units of PKT. Both of these numbers are
represented in little endian form because they are not used for ordering.

The balance computation logic maintains one or two balance entries in the balance value field in accordance
with the logic described in the Address balance computation section.

### votewinnerdb
This table contains the winners for each epoch where votes have been counted. The keys in this table are
the big endian representations of the block heights where the voting closed (this is not where the winner
becomes *effective*, nor even where the counting begins, it is where the voting is nolonger considered).

Each value in this table are the hash of the votes as they were scanned and computed, followed by the
address of the winner represented in pkScript form. If there is no winner (i.e. nobody has voted) then
the value contains only the hash.

## Database pruning
An address balance which is zero is semantically no different from one which is missing from the table.
However, the vote counting takes into consideration not the current value of an address but the value
at the end of the previous epoch, so address balances are used during the vote computation for the
previous epoch.

An address balance which has been zero since the end of the last epoch can safely be discarded once the
vote counting is complete. This is because even in the event of a rollback which crosses the epoch
boundary, the correct balance at the end of the the last epoch will be re-constituted by the rollback
of blocks and related unmint/unspend of transactions.

In the interest of reducing any possible risk of bugs, we only prune balances which have continuously
been zero for more than *two* epochs.

Pruning is performed by the vote counting thread and takes place after the completion of the vote count.
Like vote counting, pruning is performed in stages which may take no more than 50 milliseconds each.
This ensures the database will remain accessible for the handling of new blocks even while these
"background" tasks are performed.

### Pruning of votes
Votes lose their validity in one of two different ways:
1. The vote is more than 52 epochs (1 year) old
2. The same address has cast another vote

Currently votes are **not** pruned ever because there is no way to recover them after a rollback so
pruning them without introducing a prune journal table would impose a limit on number of blocks that
can safely be rolled back.

## Address balance computation
The vote balance system relies on knowing the balance of each address at the vote cutoff point (the end
of the last epoch). In order to do this, one or two address balance entries are maintained for each
non-zero-balance address.

### applyBalanceChange()
These entries are updated by a function called `applyBalanceChange()` which takes the existing entry,
if any, and the sum of address balance changes from the applied or reverted block, and creates a new
balance entry. This function does the following:

1. Save the most recent address balance and block height (if no previous balance entry then this is
considered to be zero and zero)
2. Remove any existing balance entry whose block height which is *greater* than the current block (rollback)
3. Remove any existing balance entry whose block height falls within the same epoch as the current block
4. Remove any existing balance entry whose block height falls within an epoch *older* than the previous
epoch, before the one in which the current block falls.
5. Compute the sum of the saved balance from 1 and balance change implied by this block, and append as a
new entry with the new block height.

This means as long as the balance changes of every new block and of every rollback are faithfully passed
to `applyBalanceChange()`, the balances table will keep the current balance of every address as well as
the balance as of end of the previous epoch.

### getBlockChanges()
In order to determine the address balance changes implied by a new block or a rollback, we must sum the
balance differences implied by all of the transaction outputs spent in that block and the transaction
outputs created, for each impacted address.

In the event that the block is a rollback, we simply multiply all balance changes by `-1` and deincrement
the block number by `1`, such that the result of `applyBalanceChange()` will correctly reflect the
previous block.

## Vote counting
Vote counting is performed in a background thread, in such a way as to not incubmer the normal operation
of the node any more than is absolutely necessary. Clearly there is a point at which the result of the
vote must be known, but this point is placed *far* away from the deadline for voting. This way the vote
counter may take hours to complete without affecting the normal operation of the blockchain.

Votes are counted once per *epoch*, an epoch is `10,080` blocks long (with a 1 minute block time, this
corrisponds to 7 days). The voting *deadline* for the first vote is block `10,080`, for the second vote
`20,160` and so on.

The block height at which the result of the vote becomes *effective*, the "inauguration height", is `360`
blocks past the deadline, thus allowing for about 6 hours of blocks to pass while the vote counting
takes place - before the lack of a result could cause a chain stall. Though there is no consensus rule
related to the vote counting *yet*, it is a requirement of the vote counting logic that it must have
a valid result that is available on demand for any given block height such that a consensus rule could
be established.

In case there is a rollback which crosses the vote deadline, while vote counting is in progress, the vote
counting thread is aborted. To minimize the chance of an aborted vote count, the vote counting does not
begin until `60` blocks after the vote deadline. This is *not* a consensus rule, it is an implementation
detail with the objective of reducing the likelihood that a vote count will need to be re-run more than
once.

To minimize the impact the vote counting has on the normal operation of the blockchain, the counting
takes place in stages. Each stage can take no more than 50 milliseconds with the database lock held.
After 50 milliseconds the stage unlocks the database such that the normal database updates can take
place. Because the vote counting addresses a past state of the balance table, blocks can be imported
and *current* balances and votes can change while the vote counting thread is in progress, without
changing the outcome of the count. This is of course not a consensus rule but an implementation detail.

### Vote counting process
The vote counting process scans the `votebalance` table collecting address balances and votes into a
red-black tree ordered by address balance. Addresses are only included if:
1. They have non-zero balance
2. They have an unexpired vote object which is either voting for someone, candidating, or both.

This red-black tree has a size limit of `100,000`, if the tree goes over this limit, addresses are
evicted in the order of smallest balance first and an `overLimit` flag is set.

If after scanning the table, the `overLimit` flag has been set, the counting process re-scans the
`votebalance` table a second time, this time it searches for addresses which fit the following
criteria:
1. They have non-zero balance
2. They have an unexpired vote object which is voting for someone
3. They were not included in the red-black tree created previously
4. They are voting for a candidate who is included in the red-black tree

These address/votes are applied to the candidates they voted by updating the candidate's effective
balance as stored within the red-black tree.

The effect of this 2 pass system is that *any* address can vote, but only addresses within the top
`100,000` balances can be voted for, either as a candidate or as a "delegate" who will pass through
their vote to another candidate.

### Winner computation
Once the red-black tree has been populated, the address/vote/balances in it are passed along to the
[Electorium_go](https://github.com/cjdelisle/Electorium_go) implementation of the
[Electorium](https://github.com/cjdelisle/Electorium) winner computation algorithm. This is well
described elsewhere so it will not be described in detail here.

The result of the winner computation is placed in the `votewinnerdb` along with the vote table hash.

## Vote table hash
The vote table hash is a tool for helping verify the integrity of the `votebalance` table. When
there is a chain-fork, some nodes will receive first the block which eventually becomes the standard,
while others will receive first a block which is rolled back.

The address balance computation logic should in all cases reach the same consensus on the resulting
balance, no matter what order of new blocks and rolled back blocks they receive in order to get
there.

In the event of a terrible bug, it may become possible that different full nodes would come to
different conclusions about who has won the election - which would after this voting is applied to
consensus in a soft fork - cause an accidental hard fork.

To minimize the risk of this happening, we compute a hash of every entry in the votebalance table
as we are counting the votes, in order that any discrepancy, even small, will be immediately
obvious to the node operators.

The way the hash is computed is as follows:
1. Initialize a new Blake2b-256 hasher
2. Hash the number 1 - to indicate section 1
3. For each non-zero-balance address in the votebalance table:
  1. Hash the length of the address pkScript
  2. Hash the address pkScript bytes
  3. Hash the address balance as of the end of the epoch which is being computed
4. Hash the number 0 - to indicate end of section 1
5. Hash the number 2 - to indicate the beginning of section 2
6. For each of the votes which has been collected in the red-black tree:
  1. Hash the length of the voter's address as pkScript
  2. Hash the voter's address as pkScript
  3. Hash the length of the address that is being voted for as pkScript
  4. Hash the address that is being voted for as pkScript
  5. Hash the *effective* balance of the voting address, including the coins which were "assigned"
  to the addresses by voters who do not have enough balance to be included in the red-black tree.
7. Compute the hash sum

In all cases where a number is hashed, that number is represented as a little endian uint64.
This process ensures firstly that any significant deviation in the votebalance table or in what
is sent to the Electorium election computation library will result in a different hash.

The use of lengths, begin markers, and end markers ensures that there is no way that two different
`votebalance` table states could have the same binary representation.

Notably, zero-balance addresses are *not* hashed, this is because a zero-balance address is not
semantically different from an omitted entry, and at some point a zero balance address will
eventually be pruned, but the code makes no consensus guarantees about when that pruning will take
place.

When an address has a balance, casts a vote, then the balance becomes zero, the vote entry remains
in the table until it naturally expires after 1 year, but it is disconsidered for the purpose
of vote computation, and it is excluded from the hash.

If the address later acquires a balance again, the vote will begin to be counted again.