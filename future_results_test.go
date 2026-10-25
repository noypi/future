package future_test

import (
	"testing"

	. "github.com/noypi/future"
	assertpkg "github.com/stretchr/testify/assert"
)

// AC-FUTURE-029 (parent FR-FUTURE-015, API-FUTURE-011): Results() returns a
// fresh slice holding the values a completed settle pipeline recorded. It is
// nil before any pipeline records (an unexecuted promise and a promise
// executed but never settled included), a settled promise that recorded zero
// values returns a non-nil empty slice, and every call allocates a new backing
// array so mutating one returned slice never changes the promise.
//
// At base a01f86f the settle wrappers are still created from the first
// registered handler (FR-FUTURE-014 handler-less settlement is not
// implemented), so every settling row registers both a success and a failure
// handler whose signatures exactly match the executor's declared resolve and
// reject parameter types (RULE-FUTURE-002, RULE-FUTURE-004). Reads happen
// after settlement returns: exec(false) for the synchronous rows and Wait()
// for the asynchronous row (FR-FUTURE-015 Notes, A5). No State() or Err()
// behavior is asserted, and no -race claim is made.
//
// The cases are one table and each row runs as a named subtest
// (FR-UNIT-TESTS-001 / AC-UNIT-TESTS-001); every row builds its own promise
// and depends on no other row (RULE-UNIT-TESTS-002).
func TestOutcome_Results(t *testing.T) {
	// resultsObs records a read taken from inside the executor before the
	// promise settled, plus proof that the executor body actually ran, so the
	// "nil before settlement" case cannot pass on a promise whose executor
	// never executed.
	type resultsObs struct {
		executorRan bool
		midRead     []interface{}
	}

	type testCase struct {
		name string

		// prepare builds and, when the row settles, executes the promise under
		// test. It returns the promise plus what the executor observed.
		prepare func(t *testing.T) (*Promise, resultsObs)

		// Exactly one of wantNil, wantEmpty, or a non-nil want states the
		// row's expected result. wantEmpty is the pinned nil-vs-empty
		// distinction: a non-nil slice with length zero.
		wantNil   bool
		wantEmpty bool
		want      []interface{}

		// verify holds the case-specific checks the want fields cannot express.
		verify func(t *testing.T, assert *assertpkg.Assertions, q *Promise, got []interface{}, obs resultsObs)
	}

	tests := []testCase{
		{
			// (a) No settle pipeline has run, so no results were recorded.
			name: "unexecuted promise returns nil",
			prepare: func(t *testing.T) (*Promise, resultsObs) {
				t.Helper()
				_, q := FutureDeferred(func(resolve func(string) string, reject func(string) string) {})
				return q, resultsObs{}
			},
			wantNil: true,
		},
		{
			// (b) The executor runs and returns without settling; a read from
			// inside the executor (before any resolve could record) is nil too.
			name: "executed but unsettled promise returns nil",
			prepare: func(t *testing.T) (*Promise, resultsObs) {
				t.Helper()

				var q *Promise
				obs := resultsObs{}
				exec, promise := FutureDeferred(func(resolve func(string) string, reject func(string) string) {
					obs.executorRan = true
					obs.midRead = q.Results()
				})
				q = promise

				q.Then(func(s string) string { return "out:" + s }, func(reason string) string { return reason })
				exec(false)

				return q, obs
			},
			wantNil: true,
			verify: func(t *testing.T, assert *assertpkg.Assertions, q *Promise, got []interface{}, obs resultsObs) {
				t.Helper()
				assert.True(obs.executorRan, "the executor must have run so the mid-executor read is meaningful")
				assert.Nil(obs.midRead, "Results() must be nil inside the executor before resolve settles the promise")
			},
		},
		{
			// (c) A completed resolve pipeline records its handler's output:
			// the recorded value is the handler result, not the resolve
			// argument, and it is a concrete value rather than a reflect.Value.
			name: "resolve returns the recorded success pipeline output",
			prepare: func(t *testing.T) (*Promise, resultsObs) {
				t.Helper()
				exec, q := FutureDeferred(func(resolve func(string) string, reject func(string) string) {
					resolve("value")
				})
				q.Then(func(s string) string { return "out:" + s }, func(reason string) string { return reason })
				exec(false)
				return q, resultsObs{}
			},
			want: []interface{}{"out:value"},
			verify: func(t *testing.T, assert *assertpkg.Assertions, q *Promise, got []interface{}, obs resultsObs) {
				t.Helper()
				if !assert.Len(got, 1) {
					return
				}
				assert.IsType("", got[0], "Results() must hold the concrete recorded value, not a reflect.Value wrapper")
				assert.NotEqual("value", got[0], "Results() must expose the pipeline output, not the resolve argument")
			},
		},
		{
			// (d) Every call allocates a new backing array, and mutating or
			// appending to one returned slice leaves the promise unchanged.
			name: "each call returns an independent fresh slice",
			prepare: func(t *testing.T) (*Promise, resultsObs) {
				t.Helper()
				exec, q := FutureDeferred(func(resolve func(string) string, reject func(string) string) {
					resolve("value")
				})
				q.Then(func(s string) string { return "out:" + s }, func(reason string) string { return reason })
				exec(false)
				return q, resultsObs{}
			},
			want: []interface{}{"out:value"},
			verify: func(t *testing.T, assert *assertpkg.Assertions, q *Promise, got []interface{}, obs resultsObs) {
				t.Helper()

				first := q.Results()
				second := q.Results()
				if assert.Len(first, 1) && assert.Len(second, 1) {
					assert.True(&first[0] != &second[0], "two Results() calls must not share a backing array")
				}

				first[0] = "mutated"
				assert.Equal([]interface{}{"out:value"}, q.Results(), "mutating a returned slice must not change the promise's recorded results")

				first = append(first, "extra")
				assert.Len(first, 2, "appending to the returned slice must not affect the promise")
				assert.Equal([]interface{}{"out:value"}, q.Results(), "appending to a returned slice must not change the promise's recorded results")
			},
		},
		{
			// (e) A settled promise that recorded zero values returns a
			// non-nil empty slice, never nil.
			name: "settled with zero recorded results returns a non-nil empty slice",
			prepare: func(t *testing.T) (*Promise, resultsObs) {
				t.Helper()
				exec, q := FutureDeferred(func(resolve func(string), reject func(string) string) {
					resolve("value")
				})
				q.Then(func(s string) {}, func(reason string) string { return reason })
				exec(false)
				return q, resultsObs{}
			},
			wantEmpty: true,
		},
		{
			// (f) A rejected promise returns the failure pipeline's recorded
			// values; the rejection value is not converted to an error.
			name: "rejected promise returns the failure pipeline values without an error",
			prepare: func(t *testing.T) (*Promise, resultsObs) {
				t.Helper()
				exec, q := FutureDeferred(func(resolve func(string) string, reject func(string) string) {
					reject("rejected value")
				})
				q.Then(func(s string) string { return "out:" + s }, func(reason string) string { return reason })
				exec(false)
				return q, resultsObs{}
			},
			want: []interface{}{"rejected value"},
			verify: func(t *testing.T, assert *assertpkg.Assertions, q *Promise, got []interface{}, obs resultsObs) {
				t.Helper()
				if !assert.Len(got, 1) {
					return
				}
				assert.IsType("", got[0], "the recorded rejection value must be preserved as a concrete value")
				_, isErr := got[0].(error)
				assert.False(isErr, "Results() must not convert the rejection value into an error")
			},
		},
		{
			// (g) The asynchronous path: Wait() orders the executor
			// goroutine's settlement before the read.
			name: "async exec records results for a read after Wait",
			prepare: func(t *testing.T) (*Promise, resultsObs) {
				t.Helper()
				exec, q := FutureDeferred(func(resolve func(string) string, reject func(string) string) {
					resolve("value")
				})
				q.Then(func(s string) string { return "out:" + s }, func(reason string) string { return reason })
				exec(true)
				q.Wait()
				return q, resultsObs{}
			},
			want: []interface{}{"out:value"},
		},
		{
			// (h) A success handler may return more than one value; every
			// recorded value is exposed, in recorded order, with the length
			// preserved. The two values differ, so returning only the first
			// value or reversing the order fails this row.
			name: "resolve records every value from a multi-value handler in recorded order",
			prepare: func(t *testing.T) (*Promise, resultsObs) {
				t.Helper()
				exec, q := FutureDeferred(func(resolve func(string) (string, string), reject func(string) string) {
					resolve("value")
				})
				q.Then(func(s string) (string, string) { return s, "out:" + s }, func(reason string) string { return reason })
				exec(false)
				return q, resultsObs{}
			},
			want: []interface{}{"value", "out:value"},
			verify: func(t *testing.T, assert *assertpkg.Assertions, q *Promise, got []interface{}, obs resultsObs) {
				t.Helper()
				if !assert.Len(got, 2, "Results() must expose both recorded values") {
					return
				}
				assert.IsType("", got[0], "the first recorded value must be a concrete value, not a reflect.Value wrapper")
				assert.IsType("", got[1], "the second recorded value must be a concrete value, not a reflect.Value wrapper")
				assert.NotEqual(got[0], got[1], "the recorded values must differ so the recorded-order assertion is meaningful")
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			assert := assertpkg.New(t)

			q, obs := tt.prepare(t)
			got := q.Results()

			switch {
			case tt.wantNil:
				assert.Nil(got, "Results() must be nil before any settle pipeline records results")
			case tt.wantEmpty:
				assert.NotNil(got, "a settled promise that recorded zero values must return a non-nil empty slice")
				assert.Len(got, 0, "a zero-result settlement must return an empty slice, not a nil one")
			default:
				assert.Equal(tt.want, got, "Results() must return the recorded pipeline values in recorded order")
			}

			if tt.verify != nil {
				tt.verify(t, assert, q, got, obs)
			}
		})
	}
}
