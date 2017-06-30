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
	return a.Term == b.Term && a.Vote == b.Vote && a.Commit == b.Commit
}

func IsEmptyHardState(st pb.HardState) bool {
	return isHardStateEqual(st, emptyState)
}

func IsEmptySnap(sp pb.Snapshot) bool {
	return sp.Metadata.Index == 0
}

func (rd Ready) containsUpdates() bool {
	return rd.SoftState != nil || !IsEmptyHardState(rd.HardState) ||
		!IsEmptySnap(rd.Snapshot) || len(rd.Entries) > 0 ||
		len(rd.CommittedEntries) > 0 || len(rd.Messages) > 0 ||
		len(rd.ReadStates) != 0
}

type Node interface {
	Tick()
	Campaign(ctx context.Context) error
	Propose(ctx context.Context, data []byte) error
	ProposeConfChange(ctx context.Context, cc pb.ConfChange) error
	Step(ctx context.Context, msg pb.Message) error

	Ready() <-chan Ready

	Advance()
	ApplyConfChange(cc pb.ConfChange) *pb.ConfState

	TransferLeadership(ctx context.Context, lead, transferee uint64)

	ReadIndex(ctx context.Context, rctx []byte) error

	Status() Status
	ReportUnreachable(id uint64)
	ReportSnapshot(id uint64, status SnapshotStatus)
	Stop()
}

type Peer struct {
	ID      uint64
	Context []byte
}

func StartNode(c *Config, peers []Peer) Node {
}

func RestartNode(c *Config) Node {
}

type node struct {
	propc      chan pb.Message
	recvc      chan pb.Message
	confc      chan pb.ConfChange
	confstatec chan pb.ConfState
	readyc     chan Ready
	advancec   chan struct{}
	tickc      chan struct{}
	done       chan struct{}
	stop       chan struct{}
	status     chan struct{}

	logger Logger
}

func newNode() node {
}

func (n *node) Stop() {
}

func (n *node) run(r *raft) {
}

func (n *node) Tick() {
}

func (n *node) Campaign(ctx context.Context) error {
}

func (n *node) Propose(ctx context.Context, data []byte) error {
}

func (n *node) Step(ctx context.Context, m pb.Message) error {
}

func (n *node) ProposeConfChange(ctx context.Context, cc pb.ConfChange) error {
}

func (n *node) step(ctx context.Context, m pb.Message) error {
}

func (n *node) Ready() <-chan Ready {
}

func (n *node) Advance() {
}

func (n *node) ApplyConfChange(cc pb.ConfChange) *pb.ConfState {
}

func (n *node) Status() Status {
}

func (n *node) ReportUnreachable(id uint64) {
}

func (n *node) ReportSnapshot(id uint64, status SnapshotStatus) {
}

func (n *node) TransferLeadership(ctx context.Context, lead, transferee uint64) {
}

func (n *node) ReadIndex(ctx context.Context, rctx []byte) error {
}

func newReady(r *raft, prevSoftSt *SoftState, prevHardst pb.HardState) Ready {
}

func MustSync(st, prevst pb.HardState, entsnum int) bool {
}
