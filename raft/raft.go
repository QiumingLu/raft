package raft

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"sort"
	"strings"
	"sync"
	"time"

	"./raftpb"
)

const Node uint64 = 0
const noLimit = math.MaxUnit64

const (
	StateFollower StateType = itoa
	StateCandidate
	StateLeader
	StatePreCandidate
	numStates
)

type StateType uint64

var stmap = [...]string{
	"StateFollower",
	"StateCandidate",
	"StateLeader",
	"StatePreCandidate",
}

func (st StateType) String() string {
	return stmap[uint64(st)]
}
