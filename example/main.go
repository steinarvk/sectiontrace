package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"sync"
	"time"

	"github.com/steinarvk/sectiontrace"
)

var sectionFunctionA = sectiontrace.New("FunctionA")
var sectionFunctionA_Subfunction = sectiontrace.New("FunctionA.Subfunction")
var sectionFunctionB = sectiontrace.New("FunctionB")
var sectionFunctionC = sectiontrace.New("FunctionC")
var sectionFunctionC_Setup = sectiontrace.New("FunctionC.Setup")
var sectionFunctionC_CallA = sectiontrace.New("FunctionC.CallA")
var sectionFunctionC_CallB = sectiontrace.New("FunctionC.CallB")

// FunctionA is annotated with a callback function.
func FunctionA(ctx context.Context) error {
	wg := &sync.WaitGroup{}
	wg.Add(5)

	f := func(ctx context.Context) error {
		return sectionFunctionA_Subfunction.Do(ctx, func(ctx context.Context) error {
			time.Sleep(time.Duration(rand.Float64() * float64(time.Second)))
			wg.Done()
			return fmt.Errorf("oops")
		})
	}

	return sectionFunctionA.Do(ctx, func(ctx context.Context) error {
		for i := 0; i < 5; i++ {
			go f(ctx)
		}
		wg.Wait()
		return nil
	})
}

// FunctionB is annotated with a Begin() / End() pair of calls.
func FunctionB(ctx context.Context, n int) (err error) {
	ctx, sec := sectionFunctionB.Begin(ctx)
	defer func() { sec.End(err) }()

	if n > 0 {
		time.Sleep(50 * time.Millisecond)
		return FunctionB(ctx, n-1)
	}

	return
}

// FunctionC
func FunctionC(ctx context.Context) error {
	return sectionFunctionC.Do(ctx, func(ctx context.Context) (returnErr error) {
		ctx, sec := sectionFunctionC_Setup.Begin(ctx)
		defer func() { sec.End(returnErr) }()

		time.Sleep(50 * time.Millisecond)

		ctx, sec = sec.NextPhase(sectionFunctionC_CallA)

		if err := FunctionA(ctx); err != nil {
			returnErr = err
			return
		}

		ctx, sec = sec.NextPhase(sectionFunctionC_CallB)

		if err := FunctionB(ctx, 5); err != nil {
			returnErr = err
			return
		}

		return
	})
}

func main() {
	var records []*sectiontrace.Record
	defer func() {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		enc.Encode(sectiontrace.Export(records))
	}()

	sectiontrace.OnBegin = func(begin *sectiontrace.Record) {
		records = append(records, begin)
	}
	sectiontrace.OnEnd = func(begin, end *sectiontrace.Record) {
		records = append(records, end)
	}

	FunctionC(context.Background())
}
