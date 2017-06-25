package raft

import "fmt"

const (
	ProgressStateProbe ProgressStateType = itoa
	ProgressStateReplicate
	ProgressStateSnapshot
)

type ProgressStateType uint64

var prstmap = [...]string{
	"ProgressStateProbe",
	"ProgressStateReplicate",
	"ProgressStateSnapshot",
}

func (st ProgressStateType) String() string { return prstmap[uint64(st)] }

type Progress struct {
	Match, Next     uint64
	State           ProgressStateType
	Paused          bool
	PendingSnapshot uint64
	RecentActive    bool
	ins             *inflights
}

func (pr *Progress) resetState(state ProgressStateType) {
	pr.Paused = false
	pr.PendingSnapshot = 0
	pr.State = state
	pr.ins.reset()
}

func (pr *Progress) becomeProbe() {
	if pr.State == ProgressStateSnapshot {
		pendingSnapshot := pr.PendingSnapshot
		pr.resetState(ProgressStateProbe)
		pr.Next = max(pr.Match+1, pendingSnapshot+1)
	} else {
		pr.resetState(ProgressStateProbe)
		pr.Next = pr.Match + 1
	}
}

func (pr *Progress) becomeReplicate() {
	pr.resetState(ProgressStateReplicate)
	pr.Next = pr.Match + 1
}

func (pr *Progress) becomeSnapshot(snapshoti uint64) {
	pr.resetState(ProgressStateSnapshot)
	pr.PendingSnapshot = snapshoti
}

func (pr *Progress) maybeUpdate(n uint64) bool {
	var updated bool
	if pr.Match < n {
		pr.Match = n
		updated = true
		pr.resume()
	}
	if pr.Next < n+1 {
		pr.Next = n + 1
	}
	return updated
}

func (pr *Progress) optimisticUpdate(n uint64) { pr.Next = n + 1 }

func (pr *Progress) maybeDecrTo(rejected, last uint64) bool {
	if pr.State == ProgressStateReplicate {
		if rejected <= pr.Match {
			return false
		}
		pr.Next = pr.Match + 1
		return true
	}
	if pr.Next-1 != rejected {
		return false
	}
	if pr.Next = min(rejected, last+1); pr.Next < 1 {
		pr.Next = 1
	}
	pr.resume()
	return true
}

func (pr *Progress) pause()  { pr.Paused = true }
func (pr *Progress) resume() { pr.Paused = false }

func (pr *Progress) IsPaused() bool {
	switch pr.State {
	case ProgressStateProbe:
		return pr.Paused
	case ProgressStateReplicate:
		return pr.ins.full()
	case ProgressStateSnapshot:
		return true
	default:
		panic("unexpected state")
	}
}

func (pr *Progress) snapshotFailure() {
	pr.PendingSnapshot = 0
}

func (pr *Progress) needSnapshotAbort() bool {
	return pr.State == ProgressStateSnapshot && pr.Match >= pr.PendingSnapshot
}

func (pr *Progress) String() string {
	return fmt.Sprintf("next = %d, match = %d, state = %s, waiting = %v, pendingSnapshot = %d", pr.Next, pr.Match, pr.State, pr.IsPaused(), pr.PendingSnapshot)
}

type inflights struct {
	start  int
	count  int
	size   int
	buffer []uint64
}

func newInflights(size int) *inflights {
	return &inflights{
		size: size,
	}
}

func (in *inflights) add(inflight uint64) {
	if in.full() {
		panic("cannot add into a full inflights")
	}
	next := in.start + in.count
	size := in.size
	if next >= size {
		next -= size
	}
	if next >= len(in.buffer) {
		in.growBuf()
	}
	in.buffer[next] = inflight
	in.count++
}

func (in *inflights) growBuf() {
	newSize := len(in.buffer) * 2
	if newSize == 0 {
		newSize = 1
	} else if newSize > in.size {
		newSize = in.size
	}
	newBuffer := make([]uint64, newSize)
	copy(newBuffer, in.buffer)
	in.buffer = newBuffer
}

func (in *inflights) freeTo(to uint64) {
	if in.count == 0 || to < in.buffer[in.start] {
		return
	}

	i, idx := 0, in.start
	for i = 0; i < in.count; i++ {
		if to < in.buffer[idx] {
			break
		}
		size := in.size
		if idx++; idx >= size {
			idx -= size
		}
	}
	in.count -= i
	in.start = idx
	if in.count == 0 {
		in.start = 0
	}
}

func (in *inflights) freeFirstOne() {
	in.freeTo(in.buffer[in.start])
}

func (in *inflights) full() bool {
	return in.count == in.size
}

func (in *inflights) reset() {
	in.count = 0
	in.start = 0
}
