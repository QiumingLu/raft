package raft

import (
	"errors"
	"sync"

	pb "./raftpb"
)

var ErrCompacted = errors.New("requested index is unavailable due to compaction")

var ErrSnapOutOfDate = errors.New("requested index is older than the exsiting snapshot")

var ErrUnavailable = errors.New("requested entry at index is unavailable")

var ErrSnapshotTemporarilyUnavailable = errors.New("snapshot is temporarily unavailable")

type Storage interface {
	InitialState() (pb.HardState, pb.ConfState, error)

	Entries(lo, hi, maxSize uint64) ([]pb.Entry, error)

	Term(i uint64) (uint64, error)

	LastIndex() (uint64, error)

	FirstIndex() (uint64, error)

	Snapshot() (pb.Snapshot, error)
}

type MemoryStorage struct {
	sync.Mutex

	hardState pb.HardState
	snapshot  pb.Snapshot
	ents      []pb.Entry
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		ents: make([]pb.Entry, 1),
	}
}

func (ms *MemoryStorage) InitialState() (pb.HardState, pb.ConfState, error) {
	return ms.hardState, ms.snapshot.Metadata.ConfState, nil
}

func (ms *MemoryStorage) SetHardState(st pb.HardState) error {
	ms.Lock()
	defer ms.Unlock()
	ms.hardState = st
	return nil
}

func (ms *MemoryStorage) Entries(lo, hi, maxSize uint64) ([]pb.Entry, error) {
	ms.Lock()
	defer ms.Unlock()
	offset := ms.ents[0].Index
	if lo <= offset {
		return nil, ErrCompacted
	}
	if hi > ms.lastIndex()+1 {
		raftLogger.Panicf("entries' hi(%d) is out of bound lastIndex(%d)",
			hi, ms.lastIndex())
	}

	if len(ms.ents) == 1 {
		return nil, ErrUnavailable
	}

	ents := ms.ents[lo-offset : hi-offset]
	return limitSize(ents, maxSize), nil
}

func (ms *MemoryStorage) Term(i uint64) (uint64, error) {
	ms.Lock()
	defer ms.Unlock()
	offset := ms.ents[0].Index
	if i < offset {
		return 0, ErrCompacted
	}
	if int(i-offset) >= len(ms.ents) {
		return 0, ErrUnavailable
	}
	return ms.ents[i-offset].Term, nil
}

func (ms *MemoryStorage) LastIndex() (uint64, error) {
	ms.Lock()
	defer ms.Unlock()
	return ms.lastIndex(), nil
}

func (ms *MemoryStorage) lastIndex() uint64 {
	return ms.ents[0].Index + uint64(len(ms.ents)) - 1
}

func (ms *MemoryStorage) FirstIndex() (uint64, error) {
	ms.Lock()
	defer ms.Unlock()
	return ms.firstIndex(), nil
}

func (ms *MemoryStorage) firstIndex() uint64 {
	return ms.ents[0].Index + 1
}

func (ms *MemoryStorage) Snapshot() (pb.Snapshot, error) {
	ms.Lock()
	defer ms.Unlock()
	return ms.snapshot, nil
}

func (ms *MemoryStorage) ApplySnapshot(snap pb.Snapshot) error {
	ms.Lock()
	defer ms.Unlock()

	msIndex := ms.snapshot.Metadata.Index
	snapIndex := snap.Metadata.Index
	if msIndex >= snapIndex {
		return ErrSnapOutOfDate
	}
	ms.snapshot = snap
	ms.ents = []pb.Entry{{Term: snap.Metadata.Term, Index: snap.Metadata.Index}}
	return nil
}

func (ms *MemoryStorage) CreateSnapshot(i uint64, cs *pb.ConfState, data []byte) (pb.Snapshot, error) {
	ms.Lock()
	defer ms.Unlock()
	if i <= ms.snapshot.Metadata.Index {
		return pb.Snapshot{}, ErrSnapOutOfDate
	}
	if i > ms.lastIndex() {
		raftLogger.Panicf("snapshot %d is out of bound lastindex(%d)", i, ms.lastIndex)
	}

	offset := ms.ents[0].Index
	ms.snapshot.Metadata.Index = i
	ms.snapshot.Metadata.Term = ms.ents[i-offset].Term
	if cs != nil {
		ms.snapshot.Metadata.ConfState = *cs
	}
	ms.snapshot.Data = data
	return ms.snapshot, nil
}

func (ms *MemoryStorage) Compact(compactIndex uint64) error {
	ms.Lock()
	defer ms.Unlock()
	offset := ms.ents[0].Index
	if compactIndex <= offset {
		return ErrCompacted
	}
	if compactIndex > ms.lastIndex() {
		raftLogger.Panicf("compact %d is out of bound lastindex(%d)",
			compactIndex, ms.lastIndex())
	}
	i := compactIndex - offset
	ents := make([]pb.Entry, 1, 1+uint64(len(ms.ents))-i)
	ents[0].Index = ms.ents[i].Index
	ents[0].Term = ms.ents[i].Term
	ents = append(ents, ms.ents[i+1:]...)
	ms.ents = ents
	return nil
}

func (ms *MemoryStorage) Append(entries []pb.Entry) error {
	if len(entries) == 0 {
		return nil
	}
	ms.Lock()
	defer ms.Unlock()
	first := ms.firstIndex()
	last := entries[0].Index + uint64(len(entries)) - 1

	if last < first {
		return nil
	}

	if first > entries[0].Index {
		entries = entries[first-entries[0].Index:]
	}

	offset := entries[0].Index - ms.ents[0].Index
	switch {
	case uint64(len(ms.ents)) > offset:
		ms.ents = append([]pb.Entry{}, ms.ents[:offset]...)
		ms.ents = append(ms.ents, entries...)
	case uint64(len(ms.ents)) == offset:
		ms.ents = append(ms.ents, entries...)
	default:
		raftLogger.Panicf("missing log entry [last: %d, append at: %d]",
			ms.lastIndex(), entries[0].Index)
	}
	return nil
}
