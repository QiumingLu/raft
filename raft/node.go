package raft

import (
	"errors"

	pb "./raftpb"
	"golang.org/x/next/context"
)

type SnapshotStatus int

const (
	SnapshotFinish  SnapshotStatus = 1
	SnapshotFailure SnapshotStatus = 2
)

var (
	emptyState = pb.HardState{}

	ErrStopped = errors.New("raft:stopped")
)

type SoftState struct {
	Lead      uint64
	RaftState StateType
}

func (a *SoftState) equal(b *SoftState) bool {
	return a.Lead == b.Lead && a.RaftState == b.RaftState
}

type Ready struct {
	*SoftState
	pb.HardState
	ReadStates       []ReadState
	Entries          []pb.Entry
	CommittedEntries []pb.Entry
	Messages         []pb.Message
	MustSync         bool
}

func isHardStateEqual(a, b pb.HardState) bool {
}

func IsEmptyHardState(st pb.HardState) bool {
}

func IsEmptySnap(sp pb.Snapshot) bool {
	return sp.Metadata.Index == 0
}

func (rd Ready) containsUpdates() bool {
}

type Node interface {
}

type Peer struct {
}

func StartNode(c *Config, peers []Peer) Node {
}
