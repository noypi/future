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

// The tests below pin AC-FUTURE-015 (parent FR-FUTURE-010): a panic escaping
// the executor or a success/failure handler is recovered, the promise's state
// becomes FinalRecovered, and the recovery callback observes that state and
// the recovered value. They are white-box (package future) because the
// criterion's subject is the unexported state field; both handlers are
// registered before exec (RULE-FUTURE-004 / Q-FUTURE-008) and each promise is
// executed exactly once (RULE-FUTURE-003).

// TestRecovered_StateExecutorPanic asserts that an executor panic on the
// synchronous path is recovered and observed as FinalRecovered both inside the
// installed catch callback and on q.state after exec returns.
func TestRecovered_StateExecutorPanic(t *testing.T) {
	const sentinel = "boom-exec-state"

	var q *Promise
	var calls int
	var stateInCatch FinalState
	var gotVal interface{}

	exec, q := FutureDeferred(func(resolve func(string) string, reject func(interface{})) {
		panic(sentinel)
	})

	// Both wrappers must exist before exec builds its call arguments.
	q.Then(func(s string) string { return s }, func(o interface{}) {})

	q.SetCatch(func(err error, o interface{}) {
		calls++
		stateInCatch = q.state
		gotVal = o
	})

	exec(false)

	if calls != 1 {
		t.Fatalf("catch calls = %d, want 1", calls)
	}
	if stateInCatch != FinalRecovered {
		t.Fatalf("state observed in catch = %v, want FinalRecovered (%v)", stateInCatch, FinalRecovered)
	}
	if q.state != FinalRecovered {
		t.Fatalf("state after exec = %v, want FinalRecovered (%v)", q.state, FinalRecovered)
	}
	if gotVal != sentinel {
		t.Fatalf("recovered value = %v, want %q", gotVal, sentinel)
	}
}

// TestRecovered_StateHandlerPanic asserts that a success-handler panic on the
// synchronous path is recovered and observed as FinalRecovered both inside the
// installed catch callback and on q.state after exec returns.
func TestRecovered_StateHandlerPanic(t *testing.T) {
	const sentinel = "boom-handler-state"

	var q *Promise
	var calls int
	var stateInCatch FinalState
	var gotVal interface{}
	var laterHandlerRan bool

	exec, q := FutureDeferred(func(resolve func(string) string, reject func(interface{})) {
		resolve("value")
	})

	// The panicking success handler is registered with Then; a later success
	// handler records whether chaining stopped after the panic. That ordering
	// clause is observed only, never asserted (A7; Q-FUTURE-006's ordering
	// consequence is outside AC-FUTURE-015).
	q.Then(func(s string) string { panic(sentinel) }, func(o interface{}) {})
	q.OnSuccess(func(s string) string {
		laterHandlerRan = true
		return s
	})

	q.SetCatch(func(err error, o interface{}) {
		calls++
		stateInCatch = q.state
		gotVal = o
	})

	exec(false)

	if calls != 1 {
		t.Fatalf("catch calls = %d, want 1", calls)
	}
	if stateInCatch != FinalRecovered {
		t.Fatalf("state observed in catch = %v, want FinalRecovered (%v)", stateInCatch, FinalRecovered)
	}
	if q.state != FinalRecovered {
		t.Fatalf("state after exec = %v, want FinalRecovered (%v)", q.state, FinalRecovered)
	}
	if gotVal != sentinel {
		t.Fatalf("recovered value = %v, want %q", gotVal, sentinel)
	}

	t.Logf("handler registered after the panicking one ran: %v", laterHandlerRan)
}

// The async tests below pin the same AC-FUTURE-015 state contract for
// exec(true): the panic escapes on the executor goroutine today, so these
// tests are the expected red (a process crash) until the recovery runs inside
// the execution path. They block on q.Wait() and read captured values only
// after Wait returns (A3). No sleeps, timers, polling, or t.Parallel().

// TestRecovered_AsyncStateExecutorPanic asserts that an executor panic on the
// asynchronous path is recovered and observed as FinalRecovered inside the
// installed catch callback and on q.state after Wait returns.
func TestRecovered_AsyncStateExecutorPanic(t *testing.T) {
	const sentinel = "boom-exec-state-async"

	var q *Promise
	var calls int
	var stateInCatch FinalState
	var gotVal interface{}
	var finallyCalls int
	var finallyState FinalState
	var finallyArgs []interface{}

	exec, q := FutureDeferred(func(resolve func(string) string, reject func(interface{})) {
		panic(sentinel)
	})

	// Both wrappers must exist before exec builds its call arguments.
	q.Then(func(s string) string { return s }, func(o interface{}) {})

	q.SetCatch(func(err error, o interface{}) {
		calls++
		stateInCatch = q.state
		gotVal = o
	})

	// Record (never assert) the finally observation for Q-FUTURE-002.
	q.SetFinally(func(st FinalState, args ...interface{}) {
		finallyCalls++
		finallyState = st
		finallyArgs = args
	})

	exec(true)
	q.Wait()

	t.Logf("finally observation: calls=%d state=%v args=%v", finallyCalls, finallyState, finallyArgs)

	if calls != 1 {
		t.Fatalf("catch calls = %d, want 1", calls)
	}
	if stateInCatch != FinalRecovered {
		t.Fatalf("state observed in catch = %v, want FinalRecovered (%v)", stateInCatch, FinalRecovered)
	}
	if q.state != FinalRecovered {
		t.Fatalf("state after Wait = %v, want FinalRecovered (%v)", q.state, FinalRecovered)
	}
	if gotVal != sentinel {
		t.Fatalf("recovered value = %v, want %q", gotVal, sentinel)
	}
}

// TestRecovered_AsyncStateHandlerPanic asserts that a success-handler panic on
// the asynchronous path is recovered and observed as FinalRecovered inside the
// installed catch callback and on q.state after Wait returns.
func TestRecovered_AsyncStateHandlerPanic(t *testing.T) {
	const sentinel = "boom-handler-state-async"

	var q *Promise
	var calls int
	var stateInCatch FinalState
	var gotVal interface{}
	var laterHandlerRan bool
	var finallyCalls int
	var finallyState FinalState
	var finallyArgs []interface{}

	exec, q := FutureDeferred(func(resolve func(string) string, reject func(interface{})) {
		resolve("value")
	})

	// The panicking success handler is registered with Then; a later success
	// handler records whether chaining stopped after the panic (A7).
	q.Then(func(s string) string { panic(sentinel) }, func(o interface{}) {})
	q.OnSuccess(func(s string) string {
		laterHandlerRan = true
		return s
	})

	q.SetCatch(func(err error, o interface{}) {
		calls++
		stateInCatch = q.state
		gotVal = o
	})

	// Record (never assert) the finally observation for Q-FUTURE-002.
	q.SetFinally(func(st FinalState, args ...interface{}) {
		finallyCalls++
		finallyState = st
		finallyArgs = args
	})

	exec(true)
	q.Wait()

	t.Logf("finally observation: calls=%d state=%v args=%v", finallyCalls, finallyState, finallyArgs)
	t.Logf("handler registered after the panicking one ran: %v", laterHandlerRan)

	if calls != 1 {
		t.Fatalf("catch calls = %d, want 1", calls)
	}
	if stateInCatch != FinalRecovered {
		t.Fatalf("state observed in catch = %v, want FinalRecovered (%v)", stateInCatch, FinalRecovered)
	}
	if q.state != FinalRecovered {
		t.Fatalf("state after Wait = %v, want FinalRecovered (%v)", q.state, FinalRecovered)
	}
	if gotVal != sentinel {
		t.Fatalf("recovered value = %v, want %q", gotVal, sentinel)
	}
}
