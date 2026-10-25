package future_test

import (
	"testing"

	. "github.com/noypi/future"
	assertpkg "github.com/stretchr/testify/assert"
)

// AC-FUTURE-030 (parent FR-FUTURE-015, API-FUTURE-012): Err() exposes the
// recovered-panic error when the promise is in FinalRecovered, and nil in
// every other final state, including FinalRejected.
//
// Reads happen after settlement returns: Wait() for the asynchronous path and
// the synchronous exec(false) return for the others (FR-FUTURE-015 Notes, A3).
// The recovered error is compared by identity, so an implementation that
// re-creates or re-wraps the error per call fails. The test does not assert a
// specific %+v stack rendering and makes no -race claim.
func TestOutcome_Err(t *testing.T) {
	assert := assertpkg.New(t)

	// (a) FinalRecovered: Err() is the exact instance handed to the installed
	// catch callback, and it keeps the unchanged recovery message.
	const sentinel = "boom-err-async"

	execRecovered, qRecovered := FutureDeferred(func(resolve func(string) string, reject func(interface{})) {
		panic(sentinel)
	})

	// The handlers must exist before exec builds its call arguments; the
	// executor panics before either handler can run (RULE-FUTURE-004).
	qRecovered.Then(func(s string) string { return s }, func(o interface{}) {})

	var catchCalls int
	var caughtErr error
	qRecovered.SetCatch(func(err error, o interface{}) {
		catchCalls++
		caughtErr = err
	})

	execRecovered(true)
	qRecovered.Wait()

	assert.Equal(1, catchCalls, "the installed catch callback must run exactly once for the recovered panic")
	assert.NotNil(caughtErr, "the catch callback must receive the recovery error")

	errRecovered := qRecovered.Err()
	if assert.NotNil(errRecovered, "Err() must return the recovered-panic error in FinalRecovered") {
		assert.True(errRecovered == caughtErr, "Err() must return the same error instance the catch callback received")
		assert.Equal("Promise Recovered", errRecovered.Error(), "Err() must expose the unchanged recovery message")
	}

	// (b) FinalResolved: a successful outcome carries no error.
	execResolved, qResolved := FutureDeferred(func(resolve func(string) string, reject func(interface{})) {
		resolve("value")
	})
	qResolved.Then(func(s string) string { return s }, func(o interface{}) {})

	execResolved(false)

	assert.Nil(qResolved.Err(), "Err() must be nil after a resolve (FinalResolved)")

	// (c) FinalRejected: a rejection value is not an error.
	execRejected, qRejected := FutureDeferred(func(resolve func(string) string, reject func(interface{})) {
		reject("rejected value")
	})
	qRejected.Then(func(s string) string { return s }, func(o interface{}) {})

	execRejected(false)

	assert.Nil(qRejected.Err(), "Err() must be nil after a reject (FinalRejected)")

	// (d) FinalUnknown: nothing has executed, so there is no error to report.
	_, qUnknown := FutureDeferred(func(resolve func(string) string, reject func(interface{})) {})

	assert.Nil(qUnknown.Err(), "Err() must be nil before settlement (FinalUnknown)")
}
