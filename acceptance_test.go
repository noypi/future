package future_test

// Acceptance tests for AC-FUTURE-017 (parent FR-FUTURE-012):
// "After exec(true), Wait returns only once the executor goroutine has
// returned; after exec(false) it returns immediately because no goroutine was
// started."
//
// These tests are authored by the `tester` role and live in their own file so
// they are separable from the hash-pinned unit suite in future_test.go.

import (
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	. "github.com/noypi/future"
)

const (
	// ac017NegativeWindow is how long a conforming Wait must stay blocked while
	// the async executor is still running. Scheduling aid, not a latency
	// contract; the bound is one-sided so it cannot fail a correct Wait.
	ac017NegativeWindow = 150 * time.Millisecond
	// ac017Timeout is a liveness guard, not a latency assertion.
	ac017Timeout = 3 * time.Second
	// ac017SyncWaitBound bounds 100 consecutive Waits after exec(false).
	ac017SyncWaitBound = 2 * time.Second
)

// ac017GoroutineID returns the numeric id of the calling goroutine, parsed from
// the runtime.Stack header ("goroutine 123 [running]: ...").
func ac017GoroutineID() uint64 {
	buf := make([]byte, 64)
	n := runtime.Stack(buf, false)
	fields := strings.Fields(string(buf[:n]))
	if len(fields) < 2 {
		return 0
	}
	id, err := strconv.ParseUint(fields[1], 10, 64)
	if err != nil {
		return 0
	}
	return id
}

// ac017Prime registers one success and one failure handler before exec, because
// exec passes both wrapper reflect.Values to reflect.Call (RULE-FUTURE-004 /
// Q-FUTURE-008). The handlers are setup only; the promises here need not settle.
func ac017Prime(q *Promise) {
	q.OnSuccess(func(v string) string { return v })
	q.OnFail(func(err interface{}) {})
}

// TestAcceptance017_AsyncWaitBlocksUntilExecutorReturns proves the async clause
// in its strictest reading: the executor settles ("resolve") early but keeps
// running, so Wait must remain blocked until the executor goroutine executes its
// final statement and returns -- not merely until settlement.
func TestAcceptance017_AsyncWaitBlocksUntilExecutorReturns(t *testing.T) {
	callerID := ac017GoroutineID()

	var executorID atomic.Uint64
	var finalStatement atomic.Bool

	release := make(chan struct{})
	execReturned := make(chan struct{})
	settled := make(chan struct{})

	exec, q := FutureDeferred(func(resolve func(string) string, reject func(interface{})) {
		executorID.Store(ac017GoroutineID())
		// Wait until exec(true) has returned before settling. exec's deferred
		// finally callback reads state on the caller goroutine, and settling
		// concurrently with it is the pre-existing NFR-FUTURE-002 race; this
		// test is about Wait's blocking, not settlement safety.
		<-execReturned
		resolve("early settlement")
		close(settled)
		<-release
		finalStatement.Store(true)
	})
	ac017Prime(q)

	var releaseOnce sync.Once
	releaseFn := func() { releaseOnce.Do(func() { close(release) }) }
	t.Cleanup(releaseFn)

	exec(true)
	close(execReturned)
	select {
	case <-settled:
	case <-time.After(ac017Timeout):
		t.Fatal("executor did not settle before the Wait probe")
	}
	if finalStatement.Load() {
		t.Fatal("executor reached its final statement before Wait was probed")
	}

	waitReturned := make(chan struct{})
	go func() {
		q.Wait()
		close(waitReturned)
	}()

	select {
	case <-waitReturned:
		t.Fatal("Wait returned while the async executor goroutine was still running (after settlement but before its return)")
	case <-time.After(ac017NegativeWindow):
	}

	releaseFn()

	select {
	case <-waitReturned:
	case <-time.After(ac017Timeout):
		t.Fatal("Wait did not return after the executor goroutine returned")
	}

	if !finalStatement.Load() {
		t.Fatal("Wait returned before the executor goroutine's final statement ran")
	}
	if got := executorID.Load(); got == 0 || got == callerID {
		t.Fatalf("executor ran on goroutine %d; want a new goroutine distinct from the caller's %d", got, callerID)
	}
}

// TestAcceptance017_AsyncWaitReleasesAllWaitersAfterExecutorReturns proves that
// several concurrent Wait callers all stay blocked while the executor runs and
// all return after it returns, and that a late Wait does not block.
func TestAcceptance017_AsyncWaitReleasesAllWaitersAfterExecutorReturns(t *testing.T) {
	const waiters = 3

	release := make(chan struct{})
	executorEntered := make(chan struct{})
	executorReturned := make(chan struct{})

	exec, q := FutureDeferred(func(resolve func(string) string, reject func(interface{})) {
		close(executorEntered)
		<-release
		close(executorReturned)
	})
	ac017Prime(q)

	var releaseOnce sync.Once
	releaseFn := func() { releaseOnce.Do(func() { close(release) }) }
	t.Cleanup(releaseFn)

	exec(true)

	select {
	case <-executorEntered:
	case <-time.After(ac017Timeout):
		t.Fatal("executor goroutine did not start after exec(true)")
	}

	dones := make([]chan struct{}, waiters)
	for i := range dones {
		done := make(chan struct{})
		dones[i] = done
		go func() {
			q.Wait()
			close(done)
		}()
	}

	select {
	case <-dones[0]:
		t.Fatal("waiter 0 returned while the executor goroutine was still running")
	case <-time.After(ac017NegativeWindow):
	}
	for i := 1; i < waiters; i++ {
		select {
		case <-dones[i]:
			t.Fatalf("waiter %d returned while the executor goroutine was still running", i)
		default:
		}
	}

	releaseFn()

	for i, done := range dones {
		select {
		case <-done:
		case <-time.After(ac017Timeout):
			t.Fatalf("waiter %d did not return after the executor goroutine returned", i)
		}
	}

	select {
	case <-executorReturned:
	default:
		t.Fatal("waiters returned before the executor goroutine's final statement ran")
	}

	// Late Wait: the executor has already returned, so Wait must not block.
	start := time.Now()
	q.Wait()
	if elapsed := time.Since(start); elapsed >= ac017SyncWaitBound {
		t.Fatalf("late Wait blocked for %v after the executor had already returned", elapsed)
	}
}

// TestAcceptance017_SyncWaitReturnsImmediatelyWithoutGoroutine proves the sync
// clause: exec(false) runs the executor inline on the calling goroutine (so no
// goroutine was started) and the subsequent Waits return promptly instead of
// blocking on a wait-group counter that never reaches zero.
func TestAcceptance017_SyncWaitReturnsImmediatelyWithoutGoroutine(t *testing.T) {
	callerID := ac017GoroutineID()

	var executorID atomic.Uint64
	var runs atomic.Int32

	exec, q := FutureDeferred(func(resolve func(string) string, reject func(interface{})) {
		executorID.Store(ac017GoroutineID())
		runs.Add(1)
	})
	ac017Prime(q)

	exec(false)

	if got := runs.Load(); got != 1 {
		t.Fatalf("executor ran %d times after exec(false); want exactly 1 inline run", got)
	}
	if got := executorID.Load(); got != callerID {
		t.Fatalf("exec(false) invoked the executor on goroutine %d; want the caller's goroutine %d (no goroutine started)", got, callerID)
	}

	const repeatedWaits = 100
	done := make(chan struct{})
	start := time.Now()
	go func() {
		for i := 0; i < repeatedWaits; i++ {
			q.Wait()
		}
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(ac017Timeout):
		t.Fatal("Wait blocked after exec(false); no goroutine was started")
	}
	if elapsed := time.Since(start); elapsed >= ac017SyncWaitBound {
		t.Fatalf("%d Waits after exec(false) took %v; want immediate returns", repeatedWaits, elapsed)
	}
	if got := runs.Load(); got != 1 {
		t.Fatalf("Wait re-invoked the executor (%d runs); Wait must only synchronise", got)
	}
}
