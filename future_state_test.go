package future

import "testing"

// The tests in this file pin AC-FUTURE-013 (parent FR-FUTURE-008): the
// promise's final state is FinalUnknown until a settle callback fires,
// becomes FinalResolved before the success handlers run, and becomes
// FinalRejected before the failure handlers run.
//
// They are white-box (package future) because the specification's subject is
// the unexported data-model field state, and no exported accessor exposes it
// at the moments the criterion names. Each handler is registered before exec
// (RULE-FUTURE-004) and each test settles exactly once.

// TestPromiseState_UnknownUntilCallback asserts that q.state is FinalUnknown
// before the executor calls resolve, and FinalResolved once resolve returns.
func TestPromiseState_UnknownUntilCallback(t *testing.T) {
	var q *Promise
	var stateBeforeResolve, stateAfterResolve FinalState

	fn := func(resolve func(string) string, reject func(string)) {
		stateBeforeResolve = q.state
		resolve("value")
		stateAfterResolve = q.state
	}

	exec, q := FutureDeferred(fn)

	q.OnSuccess(func(value string) string { return value })
	q.OnFail(func(reason string) {})

	exec(false)

	if stateBeforeResolve != FinalUnknown {
		t.Fatalf("state before resolve = %v, want FinalUnknown (%v)", stateBeforeResolve, FinalUnknown)
	}
	if stateAfterResolve != FinalResolved {
		t.Fatalf("state after resolve = %v, want FinalResolved (%v)", stateAfterResolve, FinalResolved)
	}
}

// TestPromiseState_ResolvedBeforeSuccessHandlers asserts that every chained
// success handler observes FinalResolved as its first statement, and that
// q.state is FinalResolved after exec returns.
func TestPromiseState_ResolvedBeforeSuccessHandlers(t *testing.T) {
	var q *Promise
	var firstState, secondState FinalState

	fn := func(resolve func(string) string, reject func(string)) {
		resolve("value")
	}

	exec, q := FutureDeferred(fn)

	q.OnSuccess(
		func(value string) string {
			firstState = q.state
			return "first"
		},
		func(value string) string {
			secondState = q.state
			return "second"
		},
	)
	q.OnFail(func(reason string) {})

	exec(false)

	if firstState != FinalResolved {
		t.Fatalf("first success handler observed state = %v, want FinalResolved (%v)", firstState, FinalResolved)
	}
	if secondState != FinalResolved {
		t.Fatalf("second success handler observed state = %v, want FinalResolved (%v)", secondState, FinalResolved)
	}
	if q.state != FinalResolved {
		t.Fatalf("state after exec = %v, want FinalResolved (%v)", q.state, FinalResolved)
	}
}

// TestPromiseState_RejectedBeforeFailureHandlers asserts that every chained
// failure handler observes FinalRejected as its first statement, and that
// q.state is FinalRejected after exec returns.
func TestPromiseState_RejectedBeforeFailureHandlers(t *testing.T) {
	var q *Promise
	var firstState, secondState FinalState

	fn := func(resolve func(string) string, reject func(string) string) {
		reject("reason")
	}

	exec, q := FutureDeferred(fn)

	q.OnSuccess(func(value string) string { return value })
	q.OnFail(
		func(reason string) string {
			firstState = q.state
			return "first"
		},
		func(reason string) string {
			secondState = q.state
			return "second"
		},
	)

	exec(false)

	if firstState != FinalRejected {
		t.Fatalf("first failure handler observed state = %v, want FinalRejected (%v)", firstState, FinalRejected)
	}
	if secondState != FinalRejected {
		t.Fatalf("second failure handler observed state = %v, want FinalRejected (%v)", secondState, FinalRejected)
	}
	if q.state != FinalRejected {
		t.Fatalf("state after exec = %v, want FinalRejected (%v)", q.state, FinalRejected)
	}
}
