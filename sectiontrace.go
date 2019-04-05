package sectiontrace

import (
	"context"
	"fmt"
	"time"
)

type ActiveSection interface {
	End(err error)
	NextPhase(Section) (context.Context, ActiveSection)
	GetBeginRecord() *Record
}

type Section interface {
	Do(context.Context, func(context.Context) error) error
	Begin(context.Context) (context.Context, ActiveSection)
	Subsection(string) Section
}

func New(name string) Section {
	if err := maybeCheckName(name); err != nil {
		doUsageError(err)
	}
	return &namedSection{name: name}
}

func (n *namedSection) Subsection(name string) Section {
	return New(fmt.Sprintf("%s.%s", n.name, name))
}

type namedSection struct {
	name string
}

type activeSection struct {
	kind            *namedSection
	t0, t1          time.Time
	nodeID          int32
	beginRec        *Record
	hasParent       bool
	originalContext context.Context
	wasClosed       bool
}

func (a *activeSection) GetBeginRecord() *Record {
	return a.beginRec
}

func makeRecord(name string, id int32, phase Phase, t time.Time) *Record {
	return &Record{
		Category:        DefaultCategory,
		Name:            name,
		ID:              id,
		Phase:           phase,
		Scope:           DefaultScope,
		TimestampMicros: t.UnixNano() / 1000,
		Args:            map[string]interface{}{},
		ProcessID:       ProcessID,
	}
}

func doUsageError(err error) error {
	if DebugMode {
		doPanic(fmt.Errorf("Usage error in DebugMode: %v", err))
	}
	OnUsageError(err)
	return err
}

func doPanic(err error) error {
	OnPanic(err)
	return err
}

var getTimeNow func() time.Time = func() time.Time {
	return time.Now()
}

func (n *namedSection) Begin(ctx context.Context) (context.Context, ActiveSection) {
	originalCtx := ctx

	t0 := getTimeNow()
	thisNodeID := generateNodeID()
	rec := makeRecord(n.name, thisNodeID, Begin, t0)

	if ctx != nil {
		if err := setArgsFromContext(ctx, rec.Args); err != nil {
			doPanic(err)
			return nil, nil
		}
	}

	_, hasParent := rec.Args[ArgParent]

	if ctx != nil {
		ctx = context.WithValue(ctx, ParentNodeContextKey, thisNodeID)
		if !hasParent {
			ctx = context.WithValue(ctx, AncestorNodeContextKey, thisNodeID)
		}
	}

	if OnBegin != nil {
		OnBegin(rec)
	}

	t1 := getTimeNow()

	return ctx, &activeSection{
		kind:            n,
		t0:              t0,
		t1:              t1,
		nodeID:          thisNodeID,
		beginRec:        rec,
		hasParent:       hasParent,
		originalContext: originalCtx,
	}
}

func (a *activeSection) NextPhase(next Section) (context.Context, ActiveSection) {
	a.End(nil)
	return next.Begin(a.originalContext)
}

func (a *activeSection) End(sectionError error) {
	if a.wasClosed {
		doUsageError(fmt.Errorf("Section %q closed twice (did variable get resolved before rebinding?)", a.kind.name))
		return
	}
	a.wasClosed = true

	t2 := getTimeNow()

	sectionOK := sectionError == nil
	endRec := makeRecord(a.kind.name, a.nodeID, End, t2)
	for k, v := range a.beginRec.Args {
		endRec.Args[k] = v
	}
	endRec.Args[ArgOK] = sectionOK

	if OnEnd != nil {
		OnEnd(a.beginRec, endRec)
	}

	timeSpentInternal := t2.Sub(a.t1)

	t3 := getTimeNow()

	timeSpentOverhead := t3.Sub(a.t0) - timeSpentInternal
	if OnTimeSpent != nil {
		OnTimeSpent(timeSpentOverhead, timeSpentInternal, a.hasParent)
	}
}

func (n *namedSection) Do(ctx context.Context, callback func(context.Context) error) error {
	ctx, running := n.Begin(ctx)

	callbackError := callback(ctx)

	running.End(callbackError)

	return callbackError
}
