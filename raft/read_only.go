package raft

import pb "./raftpb"

type ReadState struct {
	Index      uint64
	RequestCtx []byte
}

type readIndexStatus struct {
	req   pb.Message
	index uint64
	acks  map[uint64]struct{}
}

type readOnly struct {
	option           ReadOnlyOption
	pendingReadIndex map[string]*readIndexStatus
	readIndexQueue   []string
}

func newReadOnly(option ReadOnlyOption) *readOnly {
}

func (ro *readOnly) addRequest(index uint64, m pb.Message) {
}

func (ro *readOnly) recvAck(m pb.Message) int {
}

func (ro *readOnly) advance(m pb.Message) []*readIndexStatus {
}

func (ro *readOnly) lastPendingRequestCtx() string {
}
