package sectiontrace

import "time"

var OnNodeGenerated func()
var OnTimeSpent func(overhead, internal time.Duration, hadParent bool)

var OnPanic func(error) = func(err error) {
	panic(err)
}
var OnUsageError func(error) = OnPanic

var OnBegin func(begin *Record)
var OnEnd func(begin, end *Record)
