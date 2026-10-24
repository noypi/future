package future_test

// Acceptance tests for AC-FUTURE-015 (parent FR-FUTURE-010). They are
// deliberately disjoint from the unit tests owned by code-unit-tester
// (future_test.go, future_state_test.go): the unit tests pin the sync and async
// executor/success-handler/failure-handler panics with string sentinels. These
// acceptance tests instead exercise the criterion end to end from the public
// API only, on angles the unit tests leave open:
//
//   - many asynchronous promises panicking concurrently, each recovering with
//     its own error and its own recovered value (no cross-talk, no process
//     abort);
//   - panic values that are not strings (error, int, struct pointer, slice,
//     nil), on both execution paths, so "together with the recovered value"
//     is not only a string case;
//   - a panic raised after the executor has already settled, which must still
//     leave the state FinalRecovered;
//   - repeated calling of the Future entry point, whose only observation
//     channel is the default catch log.
//
// The state clause is observed through the exported FinalState value passed to
// the completion callback (SetFinally). That is an observable proxy for
// q.state, not a white-box read; the tests assert only the state value, never
// the completion callback's call count or its variadic arguments, so
// AC-FUTURE-016 remains unclaimed. The panic value is observed through the
// installed SetCatch callback (AC-FUTURE-014 used as an observation channel
// only).
//
// All async tests use exec(true) followed by q.Wait() and read captured values
// only after Wait returns. No sleeps, timers, polling, or t.Parallel().

import (
	stderrors "errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	. "github.com/noypi/future"
	pkgerrors "github.com/pkg/errors"
)

// acceptanceRequireRecoveredError pins the AC-FUTURE-015 error contract from
// the caller's side: the exact message, a WithStack wrapper whose Cause is a
// different error with the same message, and a formatted stack that names
// future.go.
func acceptanceRequireRecoveredError(t *testing.T, err error) {
	t.Helper()

	if err == nil {
		t.Fatal("the recovery callback must receive an error, got nil")
	}
	if got := err.Error(); got != "Promise Recovered" {
		t.Fatalf("recovered error message = %q, want %q", got, "Promise Recovered")
	}
	cause := pkgerrors.Cause(err)
	if cause == err {
		t.Fatal("recovered error is not wrapped: errors.Cause(err) == err, want a WithStack wrapper")
	}
	if got := cause.Error(); got != "Promise Recovered" {
		t.Fatalf("recovered error cause message = %q, want %q", got, "Promise Recovered")
	}
	if formatted := fmt.Sprintf("%+v", err); !strings.Contains(formatted, "future.go") {
		t.Fatalf("formatted recovered error has no future.go stack frame:\n%+v", err)
	}
}

// acceptanceRequireRecoveredValue compares the recovered panic value with the
// value that was panicked. panic(nil) is special: since Go 1.21 the runtime
// substitutes a non-nil *runtime.PanicNilError, so the recovered value is
// asserted non-nil instead of equal.
func acceptanceRequireRecoveredValue(t *testing.T, got, want interface{}) {
	t.Helper()

	if want == nil {
		if got == nil {
			t.Fatal("recovered value = nil for panic(nil); the runtime substitutes a non-nil panic value")
		}
		return
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("recovered value = %#v (%T), want %#v (%T)", got, got, want, want)
	}
}

// TestAcceptance_Recovered_ConcurrentAsyncPanicsSurvive runs sixteen async
// promises that all panic at the same time. The process must survive every
// Wait, each catch must run once with its own recovery error, and each catch
// must receive its own sentinel value; the final state observed through the
// completion callback must be FinalRecovered for every promise.
func TestAcceptance_Recovered_ConcurrentAsyncPanicsSurvive(t *testing.T) {
	const n = 16

	type observation struct {
		calls        int
		err          error
		value        interface{}
		finallyState FinalState
	}

	observations := make([]observation, n)
	qs := make([]*Promise, n)
	sentinels := make([]string, n)

	for i := 0; i < n; i++ {
		i := i
		sentinel := fmt.Sprintf("boom-accept-concurrent-%02d", i)
		sentinels[i] = sentinel

		exec, q := FutureDeferred(func(resolve func(string) string, reject func(interface{})) {
			panic(sentinel)
		})

		// Both wrappers must exist before exec builds its call arguments.
		q.Then(func(s string) string { return s }, func(o interface{}) {})

		q.SetCatch(func(err error, o interface{}) {
			observations[i].calls++
			observations[i].err = err
			observations[i].value = o
		})
		// Observation channel for the state clause; only the state is asserted.
		q.SetFinally(func(st FinalState, args ...interface{}) {
			observations[i].finallyState = st
		})

		qs[i] = q
		exec(true)
	}

	for i, q := range qs {
		q.Wait()

		obs := observations[i]
		if obs.calls != 1 {
			t.Fatalf("promise %d: catch calls = %d, want 1", i, obs.calls)
		}
		if obs.value != sentinels[i] {
			t.Fatalf("promise %d: recovered value = %v, want %q", i, obs.value, sentinels[i])
		}
		acceptanceRequireRecoveredError(t, obs.err)
		if obs.finallyState != FinalRecovered {
			t.Fatalf("promise %d: FinalState passed to the completion callback = %v, want FinalRecovered (%v)",
				i, obs.finallyState, FinalRecovered)
		}
	}
}

type acceptancePayload struct {
	name string
	n    int
}

// TestAcceptance_Recovered_PanicValueTypes drives both execution paths with
// panic values the unit tests do not use: an error, an int, a struct pointer, a
// slice (not comparable), and nil. For every case the caller must continue (sync)
// or Wait must return (async), the recovered value must be preserved, the
// recovery error must keep the AC-FUTURE-015 identity, and the observed final
// state must be FinalRecovered.
func TestAcceptance_Recovered_PanicValueTypes(t *testing.T) {
	cases := []struct {
		name  string
		value interface{}
	}{
		{"string", "boom-accept-value"},
		{"error", stderrors.New("boom-accept-error-value")},
		{"int", 42},
		{"struct-pointer", &acceptancePayload{name: "payload", n: 7}},
		{"slice", []string{"a", "b"}},
		{"nil", nil},
	}

	for _, tc := range cases {
		tc := tc
		for _, async := range []bool{false, true} {
			mode := "sync"
			if async {
				mode = "async"
			}
			t.Run(tc.name+"/"+mode, func(t *testing.T) {
				var (
					calls        int
					gotErr       error
					gotValue     interface{}
					finallyState FinalState
				)

				exec, q := FutureDeferred(func(resolve func(string) string, reject func(interface{})) {
					panic(tc.value)
				})

				// Both wrappers must exist before exec builds its call arguments.
				q.Then(func(s string) string { return s }, func(o interface{}) {})

				q.SetCatch(func(err error, o interface{}) {
					calls++
					gotErr = err
					gotValue = o
				})
				q.SetFinally(func(st FinalState, args ...interface{}) {
					finallyState = st
				})

				returnedNormally := false
				if async {
					exec(true)
					q.Wait()
					returnedNormally = true
				} else {
					exec(false)
					// The panic must not reach the caller: execution continues here.
					returnedNormally = true
				}

				if !returnedNormally {
					t.Fatal("exec did not return to the caller")
				}
				if calls != 1 {
					t.Fatalf("catch calls = %d, want 1", calls)
				}
				acceptanceRequireRecoveredValue(t, gotValue, tc.value)
				acceptanceRequireRecoveredError(t, gotErr)
				if finallyState != FinalRecovered {
					t.Fatalf("FinalState passed to the completion callback = %v, want FinalRecovered (%v)",
						finallyState, FinalRecovered)
				}
			})
		}
	}
}

// TestAcceptance_Recovered_ExecutorPanicAfterSettle covers a panic raised by
// the executor after it has already settled the promise. The criterion says a
// panic escaping the executor is recovered and the state becomes
// FinalRecovered, so a later panic must override the settled state; the
// completion callback observes the state the caller can assert.
func TestAcceptance_Recovered_ExecutorPanicAfterSettle(t *testing.T) {
	const sentinel = "boom-accept-after-settle"

	for _, async := range []bool{false, true} {
		mode := "sync"
		if async {
			mode = "async"
		}
		t.Run(mode, func(t *testing.T) {
			var (
				calls        int
				gotErr       error
				gotValue     interface{}
				finallyState FinalState
			)

			exec, q := FutureDeferred(func(resolve func(string) string, reject func(interface{})) {
				resolve("settled-value")
				panic(sentinel)
			})

			q.Then(func(s string) string { return s }, func(o interface{}) {})
			q.SetCatch(func(err error, o interface{}) {
				calls++
				gotErr = err
				gotValue = o
			})
			q.SetFinally(func(st FinalState, args ...interface{}) {
				finallyState = st
			})

			if async {
				exec(true)
				q.Wait()
			} else {
				exec(false)
			}

			if calls != 1 {
				t.Fatalf("catch calls = %d, want 1", calls)
			}
			if gotValue != sentinel {
				t.Fatalf("recovered value = %v, want %q", gotValue, sentinel)
			}
			acceptanceRequireRecoveredError(t, gotErr)
			if finallyState != FinalRecovered {
				t.Fatalf("FinalState passed to the completion callback = %v, want FinalRecovered (%v)",
					finallyState, FinalRecovered)
			}
		})
	}
}

// TestAcceptance_Recovered_HandlerPanicState covers the handler half of
// FR-FUTURE-010 on both execution paths, including the failure-handler panic
// for which no white-box state assertion exists. The final state is observed
// through the exported FinalState passed to the completion callback, and the
// recovered value/error through the installed catch callback.
func TestAcceptance_Recovered_HandlerPanicState(t *testing.T) {
	cases := []struct {
		name     string
		settle   string // "resolve" or "reject"
		handlers func(panicking func(interface{})) (success, failure interface{})
		value    string
	}{
		{
			name:   "success-handler",
			settle: "resolve",
			handlers: func(panicking func(interface{})) (interface{}, interface{}) {
				return func(s string) string { panicking("boom-accept-success-handler"); return s }, func(o interface{}) {}
			},
			value: "boom-accept-success-handler",
		},
		{
			name:   "failure-handler",
			settle: "reject",
			handlers: func(panicking func(interface{})) (interface{}, interface{}) {
				return func(s string) string { return s }, func(o interface{}) { panicking("boom-accept-failure-handler") }
			},
			value: "boom-accept-failure-handler",
		},
	}

	for _, tc := range cases {
		tc := tc
		for _, async := range []bool{false, true} {
			mode := "sync"
			if async {
				mode = "async"
			}
			t.Run(tc.name+"/"+mode, func(t *testing.T) {
				var (
					calls        int
					gotErr       error
					gotValue     interface{}
					finallyState FinalState
				)

				exec, q := FutureDeferred(func(resolve func(string) string, reject func(interface{})) {
					if tc.settle == "resolve" {
						resolve("ignored")
						return
					}
					reject("ignored")
				})

				panicWith := func(v interface{}) { panic(v) }
				successFn, failureFn := tc.handlers(panicWith)
				q.Then(successFn, failureFn)

				q.SetCatch(func(err error, o interface{}) {
					calls++
					gotErr = err
					gotValue = o
				})
				q.SetFinally(func(st FinalState, args ...interface{}) {
					finallyState = st
				})

				if async {
					exec(true)
					q.Wait()
				} else {
					exec(false)
				}

				if calls != 1 {
					t.Fatalf("catch calls = %d, want 1", calls)
				}
				if gotValue != tc.value {
					t.Fatalf("recovered value = %v, want %q", gotValue, tc.value)
				}
				acceptanceRequireRecoveredError(t, gotErr)
				if finallyState != FinalRecovered {
					t.Fatalf("FinalState passed to the completion callback = %v, want FinalRecovered (%v)",
						finallyState, FinalRecovered)
				}
			})
		}
	}
}

// TestAcceptance_Recovered_FutureEntryPointRepeatedPanics exercises the Future
// entry point (FutureDeferred + exec(true)) eight times. Future installs no
// catch callback before execution, so the only observation channel is the
// default catch log; each promise must survive its panic, Wait must return, and
// each recovery must produce a "Promise Recovered" log line.
func TestAcceptance_Recovered_FutureEntryPointRepeatedPanics(t *testing.T) {
	const n = 8

	logBuf := captureDefaultLogger(t) // helper from the existing external test file

	qs := make([]*Promise, n)
	for i := 0; i < n; i++ {
		i := i
		qs[i] = Future(func(resolve func(string) string, reject func(interface{})) {
			panic(fmt.Sprintf("boom-accept-future-%02d", i))
		})
	}

	// Wait must return for every promise even though every executor panicked.
	// A hang here is bounded by the package-level go test -timeout.
	for i, q := range qs {
		q.Wait()
		if q == nil {
			t.Fatalf("promise %d: Future returned nil", i)
		}
	}

	if got := strings.Count(logBuf.String(), "Promise Recovered"); got < n {
		t.Fatalf("default catch logged %d %q lines, want at least %d; log:\n%s",
			got, "Promise Recovered", n, logBuf.String())
	}
}
