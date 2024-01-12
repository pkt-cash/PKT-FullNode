package candidatetree

import (
	electorium "github.com/cjdelisle/Electorium_go"
	"github.com/emirpasic/gods/trees/redblacktree"
	"github.com/emirpasic/gods/utils"
)

func PreferRichest(a, b interface{}) int {
	s1 := a.(*electorium.Vote)
	if s1 == nil {
		panic("PreferRichest: s1 == nil")
	}
	s2 := b.(*electorium.Vote)
	if s2 == nil {
		panic("PreferRichest: s2 == nil")
	}

	if s1.NumberOfVotes < s2.NumberOfVotes {
		return 1
	} else if s1.NumberOfVotes > s2.NumberOfVotes {
		return -1
	} else {
		// Tie breaker
		return utils.StringComparator(s1.VoterId, s2.VoterId)
	}
}

type CandidateTree struct {
	tree          *redblacktree.Tree
	votesInternal []electorium.Vote
	byId          map[string]*electorium.Vote
	sizeLimit     uint32
	overLimit     bool
}

func NewCandidateTree(sizeLimit uint32) *CandidateTree {
	return &CandidateTree{
		tree:      redblacktree.NewWith(PreferRichest),
		sizeLimit: sizeLimit,
		byId:      make(map[string]*electorium.Vote),
	}
}

func (ct *CandidateTree) AddCandidate(v *electorium.Vote) {
	var realVote *electorium.Vote
	if ct.tree.Size() >= int(ct.sizeLimit) {
		realVote = ct.GetWorst()
		ct.tree.Remove(realVote)
		delete(ct.byId, realVote.VoterId)
		ct.overLimit = true
		*realVote = *v
	} else {
		ct.votesInternal = append(ct.votesInternal, *v)
		realVote = &ct.votesInternal[len(ct.votesInternal)-1]
	}
	ct.tree.Put(realVote, struct{}{})
	ct.byId[realVote.VoterId] = realVote
}

func (ct *CandidateTree) GetWorst() *electorium.Vote {
	n := ct.tree.Right()
	if n == nil {
		return nil
	}
	return n.Key.(*electorium.Vote)
}

func (ct *CandidateTree) OverLimit() bool {
	return ct.overLimit
}

func (ct *CandidateTree) NodesById() map[string]*electorium.Vote {
	return ct.byId
}

func (ct *CandidateTree) Votes() *[]electorium.Vote {
	return &ct.votesInternal
}
