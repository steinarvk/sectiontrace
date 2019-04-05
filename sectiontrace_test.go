package sectiontrace

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
	"time"
)

var currentTestTime time.Time
var collectedRecords []*Record
var usageErrors []error

func init() {
	getTimeNow = func() time.Time {
		return currentTestTime
	}
	OnBegin = func(begin *Record) {
		collectedRecords = append(collectedRecords, begin)
	}
	OnEnd = func(_, end *Record) {
		collectedRecords = append(collectedRecords, end)
	}
	OnUsageError = func(err error) {
		usageErrors = append(usageErrors, err)
	}
	DefaultScope = "testscope"
	DefaultOtherData["test"] = "yes"
	ProcessID = 123
}

func TestCallbackTracing(t *testing.T) {
	currentTestTime = time.Unix(1230, 0)
	collectedRecords = nil
	nextNodeID = 0

	mySection1 := New("section1")
	mySection2 := New("section2")
	mySection3 := New("section3")

	ctx := context.Background()

	_ = mySection1.Do(ctx, func(ctx context.Context) error {
		currentTestTime = currentTestTime.Add(time.Second)
		_ = mySection2.Do(ctx, func(ctx context.Context) error {
			currentTestTime = currentTestTime.Add(3 * time.Second)
			_ = mySection3.Do(ctx, func(ctx context.Context) error {
				currentTestTime = currentTestTime.Add(10 * time.Second)
				return nil
			})
			currentTestTime = currentTestTime.Add(4 * time.Second)
			return fmt.Errorf("oops")
		})
		currentTestTime = currentTestTime.Add(2 * time.Second)
		_ = mySection2.Do(ctx, func(ctx context.Context) error {
			currentTestTime = currentTestTime.Add(5 * time.Second)
			return nil
		})
		currentTestTime = currentTestTime.Add(4 * time.Second)
		return nil
	})

	wantJSON := `
{
	"traceEvents": [
		{"cat": "Section", "scope": "testscope", "id": 1,
		 "pid": 123,
		 "name": "section1", "ph": "b",
		 "ts": 1230000000
		},
		{"cat": "Section", "scope": "testscope", "id": 2,
		 "pid": 123,
		 "name": "section2", "ph": "b",
		 "ts": 1231000000,
		 "args": {"p": 1, "a": 1}
		},
		{"cat": "Section", "scope": "testscope", "id": 3,
		 "pid": 123,
		 "name": "section3", "ph": "b",
		 "ts": 1234000000,
		 "args": {"p": 2, "a": 1}
		},
		{"cat": "Section", "scope": "testscope", "id": 3,
		 "pid": 123,
		 "name": "section3", "ph": "e",
		 "ts": 1244000000,
		 "args": {"p": 2, "a": 1, "ok": true}
		},
		{"cat": "Section", "scope": "testscope", "id": 2,
		 "pid": 123,
		 "name": "section2", "ph": "e",
		 "ts": 1248000000,
		 "args": {"p": 1, "a": 1, "ok": false}
		},
		{"cat": "Section", "scope": "testscope", "id": 4,
		 "pid": 123,
		 "name": "section2", "ph": "b",
		 "ts": 1250000000,
		 "args": {"p": 1, "a": 1}
		},
		{"cat": "Section", "scope": "testscope", "id": 4,
		 "pid": 123,
		 "name": "section2", "ph": "e",
		 "ts": 1255000000,
		 "args": {"p": 1, "a": 1, "ok": true}
		},
		{"cat": "Section", "scope": "testscope", "id": 1,
		 "pid": 123,
		 "name": "section1", "ph": "e",
		 "ts": 1259000000,
		 "args": {"ok": true}
		}
	],
	"displayTimeUnit": "ms",
	"otherData": {
		"test": "yes"
	}
}
`

	got := Export(collectedRecords)

	var gotReadback, wantReadback interface{}

	gotMarshalled, err := json.Marshal(got)
	if err != nil {
		t.Fatal(err)
	}

	if err := json.Unmarshal([]byte(gotMarshalled), &gotReadback); err != nil {
		t.Fatal(err)
	}

	if err := json.Unmarshal([]byte(wantJSON), &wantReadback); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(gotReadback, wantReadback) {
		if err != nil {
			t.Fatal(err)
		}
		t.Errorf("got: `\n%v\n` want `\n%v\n", string(gotMarshalled), wantJSON)
	}
}

func TestBeginEndTracing(t *testing.T) {
	currentTestTime = time.Unix(1230, 0)
	collectedRecords = nil
	nextNodeID = 0

	mySection1 := New("section1")
	mySection2 := New("section2")

	ctx := context.Background()

	func() {
		ctx, sec := mySection1.Begin(ctx)
		defer sec.End(nil)

		currentTestTime = currentTestTime.Add(time.Second)

		func() {
			_, sec := mySection2.Begin(ctx)
			defer sec.End(nil)

			currentTestTime = currentTestTime.Add(time.Second)
		}()

		currentTestTime = currentTestTime.Add(time.Second)
	}()

	wantJSON := `
{
	"traceEvents": [
		{"cat": "Section", "scope": "testscope", "id": 1,
		 "pid": 123,
		 "name": "section1", "ph": "b",
		 "ts": 1230000000
		},
		{"cat": "Section", "scope": "testscope", "id": 2,
		 "pid": 123,
		 "name": "section2", "ph": "b",
		 "ts": 1231000000,
		 "args": {"p": 1, "a": 1}
		},
		{"cat": "Section", "scope": "testscope", "id": 2,
		 "pid": 123,
		 "name": "section2", "ph": "e",
		 "ts": 1232000000,
		 "args": {"p": 1, "a": 1, "ok": true}
		},
		{"cat": "Section", "scope": "testscope", "id": 1,
		 "pid": 123,
		 "name": "section1", "ph": "e",
		 "ts": 1233000000,
		 "args": {"ok": true}
		}
	],
	"displayTimeUnit": "ms",
	"otherData": {
		"test": "yes"
	}
}
`

	got := Export(collectedRecords)

	var gotReadback, wantReadback interface{}

	gotMarshalled, err := json.Marshal(got)
	if err != nil {
		t.Fatal(err)
	}

	if err := json.Unmarshal([]byte(gotMarshalled), &gotReadback); err != nil {
		t.Fatal(err)
	}

	if err := json.Unmarshal([]byte(wantJSON), &wantReadback); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(gotReadback, wantReadback) {
		if err != nil {
			t.Fatal(err)
		}
		t.Errorf("got: `\n%v\n` want `\n%v\n", string(gotMarshalled), wantJSON)
	}
}

func TestPhaseTracing(t *testing.T) {
	currentTestTime = time.Unix(1230, 0)
	collectedRecords = nil
	nextNodeID = 0

	myContainer := New("container")
	mySection1 := New("section1")
	mySection2 := New("section2")
	mySection21 := New("section2.1")
	mySection3 := New("section3")

	ctx := context.Background()

	func() {
		ctx, sec := myContainer.Begin(ctx)
		defer func() { sec.End(nil) }()

		currentTestTime = currentTestTime.Add(time.Second)

		func() {
			_, sec := mySection1.Begin(ctx)
			defer func() { sec.End(nil) }()

			currentTestTime = currentTestTime.Add(2 * time.Second)

			ctx, sec = sec.NextPhase(mySection2)

			_ = mySection21.Do(ctx, func(ctx context.Context) error {
				currentTestTime = currentTestTime.Add(3 * time.Second)
				return nil
			})

			_, sec = sec.NextPhase(mySection3)

			currentTestTime = currentTestTime.Add(4 * time.Second)
		}()

		currentTestTime = currentTestTime.Add(5 * time.Second)
	}()

	wantJSON := `
{
	"traceEvents": [
		{"cat": "Section", "scope": "testscope", "id": 1,
		 "pid": 123,
		 "name": "container", "ph": "b",
		 "ts": 1230000000
		},
		{"cat": "Section", "scope": "testscope", "id": 2,
		 "pid": 123,
		 "name": "section1", "ph": "b",
		 "args": {"p": 1, "a": 1},
		 "ts": 1231000000
		},
		{"cat": "Section", "scope": "testscope", "id": 2,
		 "pid": 123,
		 "name": "section1", "ph": "e",
		 "args": {"p": 1, "a": 1, "ok": true},
		 "ts": 1233000000
		},
		{"cat": "Section", "scope": "testscope", "id": 3,
		 "pid": 123,
		 "name": "section2", "ph": "b",
		 "args": {"p": 1, "a": 1},
		 "ts": 1233000000
		},
		{"cat": "Section", "scope": "testscope", "id": 4,
		 "pid": 123,
		 "name": "section2.1", "ph": "b",
		 "args": {"p": 3, "a": 1},
		 "ts": 1233000000
		},
		{"cat": "Section", "scope": "testscope", "id": 4,
		 "pid": 123,
		 "name": "section2.1", "ph": "e",
		 "args": {"p": 3, "a": 1, "ok": true},
		 "ts": 1236000000
		},
		{"cat": "Section", "scope": "testscope", "id": 3,
		 "pid": 123,
		 "name": "section2", "ph": "e",
		 "args": {"p": 1, "a": 1, "ok": true},
		 "ts": 1236000000
		},
		{"cat": "Section", "scope": "testscope", "id": 5,
		 "pid": 123,
		 "name": "section3", "ph": "b",
		 "args": {"p": 1, "a": 1},
		 "ts": 1236000000
		},
		{"cat": "Section", "scope": "testscope", "id": 5,
		 "pid": 123,
		 "name": "section3", "ph": "e",
		 "args": {"p": 1, "a": 1, "ok": true},
		 "ts": 1240000000
		},
		{"cat": "Section", "scope": "testscope", "id": 1,
		 "pid": 123,
		 "name": "container", "ph": "e",
		 "args": {"ok": true},
		 "ts": 1245000000
		}
	],
	"displayTimeUnit": "ms",
	"otherData": {
		"test": "yes"
	}
}
`

	got := Export(collectedRecords)

	var gotReadback, wantReadback interface{}

	gotMarshalled, err := json.Marshal(got)
	if err != nil {
		t.Fatal(err)
	}

	if err := json.Unmarshal([]byte(gotMarshalled), &gotReadback); err != nil {
		t.Fatal(err)
	}

	if err := json.Unmarshal([]byte(wantJSON), &wantReadback); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(gotReadback, wantReadback) {
		if err != nil {
			t.Fatal(err)
		}
		t.Errorf("got: `\n%v\n` want `\n%v\n", string(gotMarshalled), wantJSON)
	}
}

func TestDetectDoubleEndError(t *testing.T) {
	currentTestTime = time.Unix(1230, 0)
	collectedRecords = nil
	nextNodeID = 0

	myContainer := New("container")
	mySection1 := New("section1")
	mySection2 := New("section2")
	mySection21 := New("section2.1")
	mySection3 := New("section3")

	ctx := context.Background()

	func() {
		ctx, sec := myContainer.Begin(ctx)
		defer sec.End(nil) // WRONG: sec is resolved immediately

		currentTestTime = currentTestTime.Add(time.Second)

		func() {
			_, sec := mySection1.Begin(ctx)
			defer sec.End(nil) // WRONG: sec is resolved immediately

			currentTestTime = currentTestTime.Add(2 * time.Second)

			ctx, sec = sec.NextPhase(mySection2)

			_ = mySection21.Do(ctx, func(ctx context.Context) error {
				currentTestTime = currentTestTime.Add(3 * time.Second)
				return nil
			})

			_, sec = sec.NextPhase(mySection3)

			currentTestTime = currentTestTime.Add(4 * time.Second)
		}()

		currentTestTime = currentTestTime.Add(5 * time.Second)
	}()

	wantJSON := `
{
	"traceEvents": [
		{"cat": "Section", "scope": "testscope", "id": 1,
		 "pid": 123,
		 "name": "container", "ph": "b",
		 "ts": 1230000000
		},
		{"cat": "Section", "scope": "testscope", "id": 2,
		 "pid": 123,
		 "name": "section1", "ph": "b",
		 "args": {"p": 1, "a": 1},
		 "ts": 1231000000
		},
		{"cat": "Section", "scope": "testscope", "id": 2,
		 "pid": 123,
		 "name": "section1", "ph": "e",
		 "args": {"p": 1, "a": 1, "ok": true},
		 "ts": 1233000000
		},
		{"cat": "Section", "scope": "testscope", "id": 3,
		 "pid": 123,
		 "name": "section2", "ph": "b",
		 "args": {"p": 1, "a": 1},
		 "ts": 1233000000
		},
		{"cat": "Section", "scope": "testscope", "id": 4,
		 "pid": 123,
		 "name": "section2.1", "ph": "b",
		 "args": {"p": 3, "a": 1},
		 "ts": 1233000000
		},
		{"cat": "Section", "scope": "testscope", "id": 4,
		 "pid": 123,
		 "name": "section2.1", "ph": "e",
		 "args": {"p": 3, "a": 1, "ok": true},
		 "ts": 1236000000
		},
		{"cat": "Section", "scope": "testscope", "id": 3,
		 "pid": 123,
		 "name": "section2", "ph": "e",
		 "args": {"p": 1, "a": 1, "ok": true},
		 "ts": 1236000000
		},
		{"cat": "Section", "scope": "testscope", "id": 5,
		 "pid": 123,
		 "name": "section3", "ph": "b",
		 "args": {"p": 1, "a": 1},
		 "ts": 1236000000
		},
		{"cat": "Section", "scope": "testscope", "id": 1,
		 "pid": 123,
		 "name": "container", "ph": "e",
		 "args": {"ok": true},
		 "ts": 1245000000
		}
	],
	"displayTimeUnit": "ms",
	"otherData": {
		"test": "yes"
	}
}
`

	got := Export(collectedRecords)

	var gotReadback, wantReadback interface{}

	gotMarshalled, err := json.Marshal(got)
	if err != nil {
		t.Fatal(err)
	}

	if err := json.Unmarshal([]byte(gotMarshalled), &gotReadback); err != nil {
		t.Fatal(err)
	}

	if err := json.Unmarshal([]byte(wantJSON), &wantReadback); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(gotReadback, wantReadback) {
		if err != nil {
			t.Fatal(err)
		}
		t.Errorf("got: `\n%v\n` want `\n%v\n", string(gotMarshalled), wantJSON)
	}

	if len(usageErrors) != 1 {
		t.Fatalf("want 1 usage errors, got: %v", usageErrors)
	}
}
