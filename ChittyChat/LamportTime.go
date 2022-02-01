package chittychat

import (
	"sync"
)

type LamportTime struct {
	Time  int32
	Mutex sync.Mutex
}

func (lamport *LamportTime) UpdateTime(recievedTime int32) {
	lamport.Mutex.Lock()

	if recievedTime > lamport.Time {
		lamport.Time = recievedTime
	}

	lamport.Mutex.Unlock()
}
