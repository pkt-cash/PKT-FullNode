package tmap

import (
	"github.com/emirpasic/gods/trees/redblacktree"
	"github.com/pkt-cash/PKT-FullNode/btcutil/er"
)

// Map is a tree-backed map, it is placed in order and uses a custom comparitor to
// compare map keys. This allows you to:
// 1. Decide what constitutes equality
// 2. Use any key type that you like
// 3. Place the entries in a logical order and iterate over them in that order
type Map[K, V any] struct {
	tm   *redblacktree.Tree
	comp func(a, b *K) int
}

// New creates a new tmap with the given comparitor function, the comparitor is used to
// identify duplicate keys and to order the map for the purpose of the ForEach function.
func New[K, V any](comp func(a, b *K) int) *Map[K, V] {
	return &Map[K, V]{
		tm: redblacktree.NewWith(func(a interface{}, b interface{}) int {
			return comp((a).(*K), (b).(*K))
		}),
		comp: comp,
	}
}

// ForEach iterates over the entries in the tmap in the order defined by the comparitor.
// It and calls the function f with the key/value pairs. If the function returns an error
// then it stops before completing and returns that error. If the error is er.LoopBreak
// then it stops returning nil.
func ForEach[K, V any](s *Map[K, V], f func(k *K, v *V) er.R) er.R {
	it := s.tm.Iterator()
	for it.Next() {
		if err := f(it.Key().(*K), it.Value().(*V)); err != nil {
			if er.IsLoopBreak(err) {
				return nil
			} else {
				return err
			}
		}
	}
	return nil
}

// Insert adds a new key/value to the tmap. If it happens that there is an old
// entry which has the a matching key, the old entry key and value are returned.
func Insert[K, V any](s *Map[K, V], k *K, v *V) (*K, *V) {
	if n, ok := s.tm.Ceiling(k); ok {
		if ok && s.comp(k, n.Key.(*K)) == 0 {
			oldK := n.Key.(*K)
			oldV := n.Value.(*V)
			s.tm.Put(k, v)
			return oldK, oldV
		}
	}
	s.tm.Put(k, v)
	return nil, nil
}

// GetEntry provids a key and value of an entry, based on an example of the key
// the returned key is a pointer to the actual key, while the input key k is
// something which is considered by the comparitor to match the key.
func GetEntry[K, V any](s *Map[K, V], k *K) (*K, *V) {
	if n, ok := s.tm.Ceiling(k); ok && s.comp(k, n.Key.(*K)) == 0 {
		return n.Key.(*K), n.Value.(*V)
	} else {
		return nil, nil
	}
}

// Len gives the size of the tmap, number of entries
func Len[K, V any](s *Map[K, V]) int {
	return s.tm.Size()
}

// Clear empties the tmap
func Clear[K, V any](s *Map[K, V]) {
	s.tm.Clear()
}
