package sm_snapshots

import (
	"io"

	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/trie"
)

type SnapshotManager interface {
	UpdateAsync()
	BlockCommittedAsync(SnapshotInfo)
	SnapshotExists(uint32, *state.L1Commitment) bool
	LoadSnapshotAsync(SnapshotInfo) <-chan error
}

type SnapshotInfo interface {
	GetStateIndex() uint32
	GetCommitment() *state.L1Commitment
	GetTrieRoot() trie.Hash
	GetBlockHash() state.BlockHash
	String() string
	Equals(SnapshotInfo) bool
}

type snapshotter interface {
	storeSnapshot(SnapshotInfo, io.Writer) error
	loadSnapshot(SnapshotInfo, io.Reader) error
}

type SliceStruct[E any] interface {
	Add(E)
	Get(int) E
	Set(int, E)
	Length() int
	ForEach(func(int, E) bool) bool
	ContainsBy(func(E) bool) bool
	Find(func(E) bool) (E, bool)
}
