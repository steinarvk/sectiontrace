package sectiontrace

import "sync/atomic"

var nextNodeID uint32

func generateNodeID() int32 {
	if OnNodeGenerated != nil {
		OnNodeGenerated()
	}
	return int32(atomic.AddUint32(&nextNodeID, 1))
}
