package future_test

import (
	"sync/atomic"
	"testing"
	"time"

	. "github.com/noypi/future"
	assertpkg "github.com/stretchr/testify/assert"
)

// raceDeadlockTimeout bounds how long the tests tolerate Race not returning.
// It is the only time-based failure; the executor delays below are ordering
// aids, not measured assertions.
const raceDeadlockTimeout = 3 * time.Second

// mustRaceReturn runs Race on its own goroutine and waits for its result,
// failing with the deadlock message if Race does not return within 3 s.
func mustRaceReturn(t *testing.T, qs ...*Promise) (got *Promise) {
	t.Helper()

	done := make(chan *Promise, 1)
	go func() {
		done <- Race(qs...)
	}()

	select {
	case got = <-done:
		return got
	case <-time.After(raceDeadlockTimeout):
		t.Fatalf("Race did not return within 3s (deadlock)")
		return nil
	}
}

// primeTypedHandlers registers the first success and failure handlers on q so
// the resolve/reject wrappers handed to the executor keep these typed
// signatures (RULE-FUTURE-004 / Q-FUTURE-008). Race's no-argument callback is
// appended after them by Then(cb, cb).
func primeTypedHandlers(q *Promise) {
	q.OnSuccess(func(v string) string { return v })
	q.OnFail(func(err interface{}) {})
}

// TestRace_ResolvesFirstToSettle proves Race returns the promise that resolves
// first, only after that promise's executor has finished, and tolerates the
// loser settling later.
func TestRace_ResolvesFirstToSettle(t *testing.T) {
	assert := assertpkg.New(t)

	var winnerFinished atomic.Bool
	exec1, q1 := FutureDeferred(func(resolve func(string) string, reject func(interface{})) {
		time.Sleep(100 * time.Millisecond)
		resolve("winner")
		// Post-settlement gate. The send on Race's buffered channel only makes
		// the Race goroutine runnable; the executor keeps running and would set
		// the flag before the receiver is scheduled. Holding the flag back for
		// 100 ms after resolve means it can only be true once Race has waited
		// for this executor to finish, so a Race that returns without Wait()
		// observes false here and fails.
		time.Sleep(100 * time.Millisecond)
		winnerFinished.Store(true)
	})
	primeTypedHandlers(q1)

	loserFinished := make(chan struct{})
	exec2, q2 := FutureDeferred(func(resolve func(string) string, reject func(interface{})) {
		time.Sleep(1 * time.Second)
		resolve("loser")
		close(loserFinished)
	})
	primeTypedHandlers(q2)

	exec1(true)
	exec2(true)

	got := mustRaceReturn(t, q1, q2)

	assert.Same(q1, got, "Race must return the first promise to resolve")
	assert.True(winnerFinished.Load(), "Race returned before the winner's executor finished; it must wait for the winner")

	select {
	case <-loserFinished:
	case <-time.After(2 * time.Second):
		t.Fatalf("late loser promise did not finish within 2s after Race returned")
	}
}

// TestRace_RejectsFirstToSettle proves Race also treats rejection as a
// settlement and returns the promise that rejects first.
func TestRace_RejectsFirstToSettle(t *testing.T) {
	assert := assertpkg.New(t)

	exec1, q1 := FutureDeferred(func(resolve func(string) string, reject func(interface{})) {
		time.Sleep(100 * time.Millisecond)
		reject("winner rejects")
	})
	primeTypedHandlers(q1)

	exec2, q2 := FutureDeferred(func(resolve func(string) string, reject func(interface{})) {
		time.Sleep(1 * time.Second)
		resolve("loser")
	})
	primeTypedHandlers(q2)

	exec1(true)
	exec2(true)

	got := mustRaceReturn(t, q1, q2)

	assert.Same(q1, got, "Race must return the first promise to reject")
}

// TestRace_SinglePromise proves Race returns its only argument once it settles.
func TestRace_SinglePromise(t *testing.T) {
	assert := assertpkg.New(t)

	exec, q := FutureDeferred(func(resolve func(string) string, reject func(interface{})) {
		time.Sleep(100 * time.Millisecond)
		resolve("only")
	})
	primeTypedHandlers(q)

	exec(true)

	got := mustRaceReturn(t, q)

	assert.Same(q, got, "Race must return its single promise")
}
