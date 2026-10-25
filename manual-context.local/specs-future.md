# SPEC-FUTURE — Async Futures and Promise Chaining

- Version: 1.19
- Status: active
- Purpose: future
- Updated: 2026-10-25

## Summary

`future` is a reflection-based Promise/Future helper for Go. A caller supplies an executor
function that takes two callbacks, `resolve` and `reject`. The package wraps that executor in a
`*Promise`, runs it either on the calling goroutine or on a new goroutine, and invokes success or
failure handlers registered on the promise. Handler signatures are not required to match: argument
and result values are fitted across chain steps by reflection, with mismatched or missing values
replaced by zero values. A promise can settle with no handler registered, and its outcome can be
read through `State`, `Results`, and `Err` after `Wait`.

## Scope

- Includes: `Future`, `FutureDeferred`, `Race`, the `Promise` methods (`Then`, `OnSuccess`,
  `OnFail`, `SetCatch`, `SetFinally`, `Wait`, `State`, `Results`, `Err`), the `FinalState` values,
  the reflective invocation and result-piping rules, and panic recovery.
- Includes (build): the module manifest (`go.mod`, `go.sum`) and the `make test` entry point that
  compile and execute the package's tests.
- Excludes: channel/select primitives, `context` cancellation, timeouts, retries, goroutine-pool
  scheduling, generic (type-parameterized) APIs, cancelable promises, and promise aggregation other
  than `Race`. Third-party dependencies (`github.com/pkg/errors`, `github.com/stretchr/testify`)
  are not specified here; only the behavior the package derives from them is. CI pipelines, release
  automation, linting, and coverage reporting are also excluded.

## Actors

| ID             | Name                                                                                  |
| -------------- | ------------------------------------------------------------------------------------- |
| ACTOR-CONSUMER | Go program importing `github.com/noypi/future` and calling the exported API           |
| ACTOR-EXECUTOR | The user-supplied function `func(resolveFn, rejectFn)` that represents the async work |
| ACTOR-HANDLER  | A user-supplied success or failure function registered on a promise                   |

## Requirements

### FR-FUTURE-001 — `FutureDeferred` validates the executor shape and panics on a malformed one.

- Kind: functional
- Status: active
- Priority: must
- Source: code
- Rationale: The executor's two callbacks are discovered by reflection; a wrong shape cannot be
  used and must fail loudly rather than silently producing a never-settling promise.
- Acceptance:
  - **AC-FUTURE-001** — Any function whose first parameter kind is `func` and second parameter
    kind is `func` is accepted, regardless of the parameters' concrete signatures and regardless
    of the executor's own return values.
  - **AC-FUTURE-002** — A function with a non-`func` parameter kind, a different arity, or a
    non-function value panics with the message
    `Future should pass a function in form of func(resolveFunc, rejectFunc)`.
- Notes: Validation checks `reflect.Kind() == reflect.Func`, `NumIn() == 2`, and the two input
  kinds only. Executor return values are never inspected.

### FR-FUTURE-002 — `FutureDeferred` defers execution and returns an execution handle.

- Kind: functional
- Status: active
- Priority: must
- Source: code
- Acceptance:
  - **AC-FUTURE-003** — `FutureDeferred(fn)` returns `(exec, q)` without invoking the executor.
  - **AC-FUTURE-004** — `exec(false)` invokes the executor on the calling goroutine.
  - **AC-FUTURE-005** — `exec(true)` increments the promise's wait group and invokes the executor
    on a new goroutine, calling `Done()` when the executor returns.
- Notes: The wait group is incremented before the goroutine is started, so a `Wait` started after
  `exec(true)` returns cannot miss the increment.

### FR-FUTURE-003 — `Future` runs the executor asynchronously and returns immediately.

- Kind: functional
- Status: active
- Priority: must
- Source: code
- Acceptance:
  - **AC-FUTURE-006** — `Future(fn)` is equivalent to `FutureDeferred(fn)` followed by
    `exec(true)`; it returns the promise while the executor may still be running.

### FR-FUTURE-004 — `Then` registers both handlers and returns the same promise.

- Kind: functional
- Status: active
- Priority: must
- Source: test
- Acceptance:
  - **AC-FUTURE-007** — `Then(fulfilledFn, rejectedFn)` registers the two functions as a success
    handler and a failure handler respectively, and returns the receiver so calls can be chained.

### FR-FUTURE-005 — `OnSuccess` and `OnFail` register one or more handlers in call order.

- Kind: functional
- Status: active
- Priority: must
- Source: code
- Acceptance:
  - **AC-FUTURE-008** — Both methods are variadic; handlers are appended in the order supplied
    and each is invoked in registration order when the corresponding callback fires.
  - **AC-FUTURE-009** — A typed nil handler is accepted and ignored (it is not appended).
  - **AC-FUTURE-025** — An untyped `nil` handler argument is treated like a typed nil: it is
    accepted and ignored, and registration does not panic.
- Notes: `OnSuccess(nil)` previously reached `reflect.TypeOf(nil)` and panicked on the nil type;
  `AC-FUTURE-025` makes the untyped nil a no-op, settling `Q-FUTURE-005`.

### FR-FUTURE-006 — `resolve` runs the success handlers as a value pipeline.

- Kind: functional
- Status: active
- Priority: must
- Source: test
- Acceptance:
  - **AC-FUTURE-010** — The executor's `resolve` arguments are passed to the first success
    handler, and each handler's return values become the arguments of the next.
  - **AC-FUTURE-011** — The last success handler's return values are recorded as the promise's
    results and are the values handed to `finally`.

### FR-FUTURE-007 — `reject` runs the failure handlers with the same pipeline semantics.

- Kind: functional
- Status: active
- Priority: must
- Source: test
- Acceptance:
  - **AC-FUTURE-012** — The executor's `reject` arguments are passed to the first failure
    handler, and each failure handler's results feed the next.

### FR-FUTURE-008 — The promise records its final state.

- Kind: functional
- Status: active
- Priority: must
- Source: code
- Acceptance:
  - **AC-FUTURE-013** — The state is `FinalUnknown` until a callback fires, becomes
    `FinalResolved` before the success handlers run, and becomes `FinalRejected` before the
    failure handlers run.

### FR-FUTURE-009 — `SetCatch` and `SetFinally` replace the default hooks.

- Kind: functional
- Status: active
- Priority: must
- Source: code
- Acceptance:
  - **AC-FUTURE-014** — `SetCatch` replaces the recovery callback and `SetFinally` replaces the
    completion callback; by default `catch` logs `fmt.Sprintf("%+v", err)` plus `" => "` plus the
    recovered value through `log.Println`, and `finally` does nothing.

### FR-FUTURE-010 — A panic escaping the executor or a handler is recovered.

- Kind: functional
- Status: active
- Priority: must
- Source: code
- Acceptance:
  - **AC-FUTURE-015** — The panic does not reach the caller; the state becomes `FinalRecovered`
    and the recovery callback receives an error created by `errors.New("Promise Recovered")`
    wrapped with `errors.WithStack`, together with the recovered value.

### FR-FUTURE-011 — `finally` runs once per execution with the final state and results.

- Kind: functional
- Status: active
- Priority: must
- Source: code
- Acceptance:
  - **AC-FUTURE-016** — After the executor returns, the completion callback is invoked exactly
    once with the promise's final state and its recorded results as variadic arguments.

### FR-FUTURE-012 — `Wait` blocks for asynchronous execution.

- Kind: functional
- Status: active
- Priority: must
- Source: code
- Acceptance:
  - **AC-FUTURE-017** — After `exec(true)`, `Wait` returns only once the executor goroutine has
    returned; after `exec(false)` it returns immediately because no goroutine was started.

### FR-FUTURE-013 — `Race` returns the first promise to settle.

- Kind: functional
- Status: active
- Priority: must
- Source: code
- Acceptance:
  - **AC-FUTURE-018** — `Race(qs...)` returns the first of its arguments to resolve or reject,
    waits for that promise to finish, and returns it.
- Notes: The source comment states "returns first promise to finish or rejected". The settle
  callback now sends once on a buffered result channel guarded by `sync.Once`, and `race_test.go`
  proves this criterion.

### FR-FUTURE-014 — A promise settles and records its outcome without a pre-registered handler.

- Kind: functional
- Status: active
- Priority: must
- Source: instruction
- Rationale: The settle callbacks are discovered by reflection from the executor's declared
  parameter types, so a handler is not a prerequisite for the executor to settle. Requiring a
  handler first makes `Future` unusable as a plain awaitable and turns a successful executor into
  `FinalRecovered`.
- Supersedes: none
- Acceptance:
  - **AC-FUTURE-026** — The resolve and reject wrappers are created from the executor's declared
    parameter types (`tFn.In(0)` and `tFn.In(1)`) when the promise is created. An executor with no
    registered handlers that calls `resolve("x")` reaches `FinalResolved` and records `"x"` as the
    result instead of panicking inside `reflect.Call`.
  - **AC-FUTURE-027** — An executor with no registered failure handler that calls `reject(...)`
    reaches `FinalRejected`, records the rejection arguments, and does not panic.
- Notes: Settles `Q-FUTURE-008` in favour of the package supplying the wrappers. The wrapper
  signature follows the executor, not the first handler; handlers of a different signature are
  still fitted positionally under `NFR-FUTURE-001`. When the handler chain is empty the settle
  arguments are the recorded results; a non-empty chain still records the last handler's return
  values (`AC-FUTURE-011`). Delivery estimate: see `## Delivery Estimates` (Feature 2).

### FR-FUTURE-015 — The promise exposes its settled outcome.

- Kind: functional
- Status: active
- Priority: must
- Source: instruction
- Rationale: `Wait` returns no value, so the only way to observe an outcome is to register a
  handler; callers that await a promise need a read path that does not require one.
- Acceptance:
  - **AC-FUTURE-028** — `State()` returns `FinalUnknown` before a settle callback fires and the
    final state afterwards; it does not block.
  - **AC-FUTURE-029** — `Results()` returns a fresh slice holding the recorded results; it is nil
    before settlement, and mutating the returned slice does not change the promise's recorded
    results.
  - **AC-FUTURE-030** — `Err()` returns the same `errors.WithStack(errors.New("Promise Recovered"))`
    error that was passed to `catch` when the state is `FinalRecovered`, and nil in every other
    state, including `FinalRejected`.
- Notes: Reads are safe after `Wait` returns under `NFR-FUTURE-002`; concurrent reads while
  settlement is in flight are outside the supported pattern. `Err()` does not convert a rejection
  value to an error; a rejected promise exposes its value through `Results()`. Delivery estimate:
  see `## Delivery Estimates` (Feature 2).

### FR-FUTURE-016 — A promise settles at most once; the first settlement wins.

- Kind: functional
- Status: active
- Priority: must
- Source: instruction
- Rationale: The package header cites promisesaplus.com, whose rule 2.1.2 allows a promise to
  transition once. `RULE-FUTURE-001` records that nothing enforces it today, so a second `resolve`
  or `reject` overwrites `state` and `results` and runs the other handler chain.
- Acceptance:
  - **AC-FUTURE-031** — After the first settle callback returns, a later call to the other callback
    returns normally and does not change the promise's outcome: the state handed to `finally`,
    `Results()`, and `Err()` remain the first settlement's, and the second callback's handlers are
    not invoked.
  - **AC-FUTURE-032** — Two settle callbacks invoked concurrently from different goroutines settle
    the promise exactly once: exactly one handler chain runs, the recorded outcome is one of the two
    settlements in full and never a mixture, and `go test -race` reports no data race on the
    promise's settlement fields.
  - **AC-FUTURE-033** — A handler that calls `Results()` and `Err()` while it runs neither
    deadlocks the settlement nor panics; the values it reads during settlement are unspecified.
- Notes: The latch is per promise, not per `exec`, so a second `exec` cannot settle a promise whose
  first execution already did (`RULE-FUTURE-003`). While `FR-FUTURE-014`'s executor-typed wrappers
  are not yet in the workspace, the wrappers are created lazily from the first registered handler
  (`RULE-FUTURE-004` is the target rule), so a settlement test registers both a success and a
  failure handler. `State()` is not required by these criteria; the observable surface is the
  `finally` state argument plus `Results()` and `Err()`. Delivery estimate: see
  `## Delivery Estimates` (Feature 3).

### FR-FUTURE-017 — A variadic handler receives the settled values packed into its variadic parameter.

- Kind: functional
- Status: active
- Priority: must
- Source: instruction
- Rationale: `fitFnArgs` builds one argument per declared parameter position and never packs a
  variadic tail, so a handler declared `func(xs ...string)` is called with a single zero `[]string`
  and loses every settled value.
- Acceptance:
  - **AC-FUTURE-034** — When the first success handler is variadic, every value the executor passes
    to `resolve` is packed, in order, into its variadic parameter; a value whose type is not the
    element type is replaced by the element type's zero value (`NFR-FUTURE-001` still applies
    positionally); the handler's return values become the next handler's arguments.
  - **AC-FUTURE-035** — The same holds for a variadic failure handler reached through `OnFail` or
    `Then`, and for a variadic handler later in either chain, which receives the previous handler's
    return values packed the same way.
- Notes: The fix belongs in `fitFnArgs` (plus any helper it needs), not in `wrappedFunc`'s `Call`,
  so the invocation stays a plain `Call` with the variadic parameter already built as a slice. Only
  the handler's own variadic parameter is packed; the wrapper built by `createVFuncWrapper` keeps
  the executor's declared shape. Delivery estimate: see `## Delivery Estimates` (Feature 4).

### FR-FUTURE-018 — A registration call is atomic across the handlers it names.

- Kind: functional
- Status: active
- Priority: must
- Source: instruction
- Rationale: `OnSuccess`, `OnFail`, and `Then` append each handler as they go, so one invalid
  argument leaves the earlier arguments registered and the resolve/reject wrapper type initialised
  from a partially applied call.
- Acceptance:
  - **AC-FUTURE-036** — `OnSuccess(f1, invalid, f2)` panics with `param should be a function` and
    registers none of the three: the handler list and the resolve wrapper type are unchanged, so a
    later `OnSuccess(f3)` treats `f3` as the first registration.
  - **AC-FUTURE-037** — `Then(fulfilledFn, rejectedFn)` registers both handlers or neither: when
    either argument is not a function and not nil, it panics with `param should be a function` and
    neither side changes. The same atomicity holds for `OnFail(f1, invalid, f2)`.
- Notes: A typed or untyped nil is still skipped rather than rejected (`AC-FUTURE-025`), so
  atomicity applies to the non-function rejection; until `AC-FUTURE-025` is implemented in the
  workspace, a test uses a non-function value such as `42` rather than `nil`. `Then` must validate
  both arguments before it mutates either list. Delivery estimate: see
  `## Delivery Estimates` (Feature 5).

### NFR-FUTURE-001 — Reflective invocation tolerates mismatched handler signatures.

- Kind: non-functional
- Status: active
- Priority: must
- Source: test
- Rationale: The library's stated purpose is to avoid repeated `interface{}` casting between
  chain steps, so heterogeneous signatures must not panic.
- Acceptance:
  - **AC-FUTURE-019** — A handler that declares more parameters than it receives is called with
    zero values for the extra parameters.
  - **AC-FUTURE-020** — A handler parameter whose type does not match the incoming value is
    replaced by a zero value of the declared type; surplus incoming values are dropped.
- Notes: Fitting is positional; the comparison is exact Go type identity (`reflect.Type`
  equality), so a named type does not satisfy its underlying type.

### NFR-FUTURE-002 — Settlement is race-free for the supported usage pattern.

- Kind: non-functional
- Status: unverified
- Priority: should
- Source: inferred
- Acceptance:
  - **AC-FUTURE-021** — With a single executor settling from its own goroutine and the caller
    reading results only after `Wait` returns, `go test -race` reports no data race.
- Notes: Incremented as `inferred` because the promise's `state`, `results`, `catch`, and
  `finally` fields are written without synchronization and no race test exists. The code may
  already satisfy this for the single-settler pattern; nothing in the repository demonstrates it.
  `FR-FUTURE-016` (`AC-FUTURE-032`) adds concurrent settle attempts to the supported pattern: two
  callbacks invoked from different goroutines must settle the promise exactly once without a data
  race. `AC-FUTURE-021` still describes the single-settler case and is unchanged.

### NFR-FUTURE-003 — The repository builds and its test suite runs from one command.

- Kind: non-functional
- Status: active
- Priority: must
- Source: instruction
- Rationale: No module metadata exists, so the package cannot be built and no test in
  `future_test.go` has ever executed. Without a build entry point the behavior recorded here is
  derived from source only and can regress unnoticed. This requirement is the fix path for
  `Q-FUTURE-004`.
- Acceptance:
  - **AC-FUTURE-022** — A `go.mod` at the repository root declares module path
    `github.com/noypi/future` — the import path used by `future_test.go` — and a Go language
    version; `go build ./...` completes with no error.
  - **AC-FUTURE-023** — A `go.sum` at the repository root pins every module required to build and
    test the package, including the direct requirements `github.com/pkg/errors` and
    `github.com/stretchr/testify` and their transitive requirements; `go mod verify` reports that
    all modules are verified.
  - **AC-FUTURE-024** — A `Makefile` at the repository root provides a `test` target that runs the
    package's test suite; `make test` exits 0 when every test passes and non-zero when any test
    fails.
- Notes: This requirement adds build metadata only; it changes no behavior in `future.go` and
  introduces no dependency beyond the two already imported. `make test` executes the `Test`
  functions in `future_test.go`; `BenchmarkFuture` is not executed unless benchmarks are
  requested explicitly, so a green `make test` is not evidence for benchmark cost.

## Data Model

| Entity          | Source          | Fields                                                                                                                                                                                                                                                                                                                                                                                              |
| --------------- | --------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `Promise`       | `future.go:13`  | `wg sync.WaitGroup`; `tResolvedWrapper reflect.Type`, `vResolvedWrapper reflect.Value`; `tRejectedWrapper reflect.Type`, `vRejectedWrapper reflect.Value`; `tFn reflect.Type`, `vFn reflect.Value`; `resolvedFuncs []*fnInfoType`, `rejectedFuncs []*fnInfoType`; `state FinalState`; `results []reflect.Value`; `catch func(err error, o interface{})`; `finally func(FinalState, ...interface{})` |
| `FinalState`    | `future.go:34`  | `int` enum: `FinalUnknown` (0), `FinalResolved` (1), `FinalRejected` (2), `FinalRecovered` (3)                                                                                                                                                                                                                                                                                                      |
| `fnInfoType`    | `future.go:193` | `v reflect.Value`, `t reflect.Type` — one registered handler (unexported)                                                                                                                                                                                                                                                                                                                           |
| `fnInfoTypeArr` | `future.go:198` | `[]*fnInfoType` with an `Append` method (unexported)                                                                                                                                                                                                                                                                                                                                                |

## Interfaces

### API-FUTURE-001 — `func Future(fn interface{}) (q *Promise)`

- Auth: n/a (library)
- Request: an executor `func(resolveFn, rejectFn)`; both parameters must be functions.
- Response: a `*Promise` whose executor is already running on a new goroutine.
- Errors: panics on a malformed executor per `FR-FUTURE-001`.

### API-FUTURE-002 — `func FutureDeferred(fn interface{}) (exec func(bAsync bool), q *Promise)`

- Request: an executor as above.
- Response: an execution handle and an unexecuted promise. The resolve and reject wrappers are
  already built from the executor's parameter types (`FR-FUTURE-014`), so the executor may settle
  before any handler is registered.
- Errors: panics on a malformed executor per `FR-FUTURE-001`.

### API-FUTURE-003 — `func (this *Promise) Then(fulfilledFn, rejectedFn interface{}) *Promise`

- Request: one success handler and one failure handler (each may be a typed nil).
- Response: the receiver.
- Errors: panics with `param should be a function` for a non-function handler.

### API-FUTURE-004 — `func (this *Promise) OnSuccess(fulfilledFns ...interface{}) *Promise`

### API-FUTURE-005 — `func (this *Promise) OnFail(rejectedFns ...interface{}) *Promise`

- Request: zero or more handlers.
- Response: the receiver.
- Errors: panics with `param should be a function` for a non-function, non-nil handler.

### API-FUTURE-006 — `func (this *Promise) SetCatch(recoverFn func(error, interface{}))`

- Request: a recovery callback. Replaces the default logger callback.

### API-FUTURE-007 — `func (this *Promise) SetFinally(finallyFn func(FinalState, ...interface{}))`

- Request: a completion callback. Replaces the default no-op.

### API-FUTURE-008 — `func (this *Promise) Wait()`

- Request: none.
- Response: none. Blocks on the promise's wait group.

### API-FUTURE-009 — `func Race(qs ...*Promise) (q *Promise)`

- Request: one or more promises.
- Response: the first promise to settle, after waiting for it.
- Errors: no explicit error path.

### API-FUTURE-010 — `func (q *Promise) State() FinalState`

- Request: none.
- Response: the promise's current final state; `FinalUnknown` before a settle callback fires.
- Errors: none. Does not block.

### API-FUTURE-011 — `func (q *Promise) Results() []interface{}`

- Request: none.
- Response: a fresh slice with the recorded results; nil before settlement. Mutating it does not
  affect the promise.
- Errors: none.

### API-FUTURE-012 — `func (q *Promise) Err() error`

- Request: none.
- Response: the recovered-panic error when the state is `FinalRecovered`; nil otherwise.
- Errors: none. A rejection is not an error and is exposed through `Results()`.

## Rules

- **RULE-FUTURE-001** — A promise settles at most once; the first settle callback to run claims the
  promise and every later call to either callback is a no-op (`FR-FUTURE-016`,
  `AC-FUTURE-031`/`AC-FUTURE-032`). Before 1.18 nothing enforced this: calling both `resolve` and
  `reject` ran both handler sets, and the last one to run won the `state` and `results` fields.
- **RULE-FUTURE-002** — Handler fitting is positional and type-exact; values are never converted
  between types.
- **RULE-FUTURE-003** — Calling `exec` more than once re-invokes the executor; handler
  registrations already on the promise run again. The settlement latch in `RULE-FUTURE-001` is per
  promise, not per execution, so a second `exec` cannot settle a promise its first execution already
  settled (`FR-FUTURE-016`); `Q-FUTURE-007` records the resulting re-execution semantics.
- **RULE-FUTURE-004** — The wrappers passed to the executor are created from the executor's
  parameter types at `FutureDeferred`, so a promise settles even when no handler is registered
  (`FR-FUTURE-014`). Handlers registered later are fitted to the settle values by the positional
  rules in `RULE-FUTURE-002`. This replaces the earlier lazy-wrapper rule, under which a promise
  with no registered success handler had a zero-value resolve wrapper and `resolve` panicked
  inside `reflect.Call`; `Q-FUTURE-008` is resolved in favour of the executor-typed wrapper.

## Errors

| Code                                                                            | Condition                                                         |
| ------------------------------------------------------------------------------- | ----------------------------------------------------------------- |
| `panic: Future should pass a function in form of func(resolveFunc, rejectFunc)` | executor is not a two-`func`-parameter function                   |
| `panic: param should be a function`                                             | a handler argument is neither a function nor nil (typed or untyped) |
| log line `<err> => <recovered value>`                                           | a panic escaped while the default recovery callback was installed |

## Verification

Test files exist in source. `covered` now means the declared `Method`/`Check` passed on the last
`nfr-audit` run: the last run, `nfr-audit run --all-rows` at 2026-10-25T06:07:02Z over the
`ac-future-029` run checkout (`/media/saif/EXTRA/workspaces/future/.worktrees/ac-future-029/_run`,
commit `2b4706f`), checked 22 rows with no failures — 15 pass, 6 unknown, 1 manual — and passed
`TC-FUTURE-002`, `TC-FUTURE-008`, `TC-FUTURE-009`, `TC-FUTURE-010`, `TC-FUTURE-011`,
`TC-FUTURE-012`, `TC-FUTURE-014`, `TC-FUTURE-015`, `TC-FUTURE-016`, `TC-FUTURE-021`, and
`TC-FUTURE-022` with `exit=0`; `TC-FUTURE-013` stays `manual`. The six rows that declare no `Method`
(`TC-FUTURE-001`, `TC-FUTURE-003` to `TC-FUTURE-007`) are `unknown` to the audit and keep their
recorded `Status` pending a `Method` decision.
`TC-FUTURE-017` to `TC-FUTURE-022` each cover exactly one acceptance criterion and live in their
own test file, so an `AC-…` selector run can close its own row; 1.15 split the two combined rows
that 1.14 added. `TC-FUTURE-021` (`AC-FUTURE-029`) and `TC-FUTURE-022` (`AC-FUTURE-030`) are
`covered`, from the 2026-10-25T06:07:02Z run above; `TC-FUTURE-021`'s
`go test -count=1 -run '^TestOutcome_Results' .` exits 0 with `--- PASS: TestOutcome_Results` and
eight named subtests. `TC-FUTURE-017`, `-018`, `-019`, and `-020` stay `missing`: their named tests
do not exist in that checkout — `future_settle_test.go`, `future_accessor_test.go`, and
`future_reject_test.go` are absent, and `future_test.go` contains no `TestFuture_UntypedNil*`
function — so their audit `verdict=pass`, `detail=exit=0` is vacuous, because
`go test -run '<regex>' .` exits 0 when nothing matches.
`TC-FUTURE-012` is the exception, described after the table: an FR row whose evidence is a recorded
test run, not an NFR pass.

| Test          | Covers                                      | File                                                                                                         | Method | Check                                                                                                                                                                                                                                                             | Status  |
| ------------- | ------------------------------------------- | ------------------------------------------------------------------------------------------------------------ | ------ | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------- |
| TC-FUTURE-001 | AC-FUTURE-004, AC-FUTURE-007, AC-FUTURE-010 | `future_test.go` (`TestFuture_x01`)                                                                          | —      | —                                                                                                                                                                                                                                                                 | covered |
| TC-FUTURE-002 | AC-FUTURE-019, AC-FUTURE-020                | `future_test.go` (`TestFuture_resolveFuncSignatures`)                                                        | test   | `go test -run TestFuture_resolveFuncSignatures ./...`                                                                                                                                                                                                             | covered |
| TC-FUTURE-003 | AC-FUTURE-012                               | `future_test.go` (`TestFuture_rejectedFuncSignatures`)                                                       | —      | —                                                                                                                                                                                                                                                                 | covered |
| TC-FUTURE-004 | AC-FUTURE-001, AC-FUTURE-002                | —                                                                                                            | —      | —                                                                                                                                                                                                                                                                 | missing |
| TC-FUTURE-005 | AC-FUTURE-003, AC-FUTURE-005, AC-FUTURE-006 | —                                                                                                            | —      | —                                                                                                                                                                                                                                                                 | missing |
| TC-FUTURE-006 | AC-FUTURE-008, AC-FUTURE-009                | —                                                                                                            | —      | —                                                                                                                                                                                                                                                                 | missing |
| TC-FUTURE-007 | AC-FUTURE-011, AC-FUTURE-016                | —                                                                                                            | —      | —                                                                                                                                                                                                                                                                 | missing |
| TC-FUTURE-008 | AC-FUTURE-013                               | `future_state_test.go` (`TestPromiseState_*`)                                                                | test   | `go test -count=1 -run '^TestPromiseState_' .`                                                                                                                                                                                                                    | covered |
| TC-FUTURE-009 | AC-FUTURE-014                               | `future_test.go` (`TestFuture_DefaultCatchLogsRecoveredValue`, `TestFuture_SetCatchReplacesDefault`, `TestFuture_SetFinallyReplacesDefault`) | test   | `go test -count=1 -run '^TestFuture_' .`                                                                                                                                                                                                                          | covered |
| TC-FUTURE-010 | AC-FUTURE-015                               | `future_state_test.go` (`TestRecovered_StateExecutorPanic`, `TestRecovered_StateHandlerPanic`, `TestRecovered_AsyncStateExecutorPanic`, `TestRecovered_AsyncStateHandlerPanic`), `future_test.go` (`TestRecovered_ErrorAndValueExecutorPanic`, `TestRecovered_ErrorAndValueSuccessHandlerPanic`, `TestRecovered_ErrorAndValueFailureHandlerPanic`, `TestRecovered_AsyncErrorAndValueExecutorPanic`, `TestRecovered_AsyncErrorAndValueSuccessHandlerPanic`, `TestRecovered_AsyncErrorAndValueFailureHandlerPanic`, `TestRecovered_FutureAsyncPanic`) | test   | go test -count=1 -run '^TestRecovered_' .                                                                                                                                                                                                                         | covered |
| TC-FUTURE-011 | AC-FUTURE-017                               | `future_test.go` (`TestFuture_WaitBlocksForAsyncExec`, `TestFuture_WaitReturnsAfterSyncExec`)                | test   | `go test -count=1 -run '^TestFuture_Wait' .`                                                                                                                                                                                                                      | covered |
| TC-FUTURE-012 | AC-FUTURE-018                               | `race_test.go` (`TestRace_ResolvesFirstToSettle`, `TestRace_RejectsFirstToSettle`, `TestRace_SinglePromise`) | test   | `go test -count=1 -v -timeout 60s -run '^TestRace_' .`                                                                                                                                                                                                            | covered |
| TC-FUTURE-013 | AC-FUTURE-021                               | —                                                                                                            | manual | 1. In a checkout with go.mod present, run `go test -race -count=1 ./...`.<br>2. Confirm the suite exercises a single executor settling from its own goroutine.<br>3. Confirm `go test -race` reports no data race.<br>4. Record the exact command and its output. | missing |
| TC-FUTURE-014 | AC-FUTURE-022                               | `go.mod`                                                                                                     | build  | `go build ./...`                                                                                                                                                                                                                                                  | covered |
| TC-FUTURE-015 | AC-FUTURE-023                               | `go.sum`                                                                                                     | build  | `go mod verify`                                                                                                                                                                                                                                                   | covered |
| TC-FUTURE-016 | AC-FUTURE-024                               | `Makefile`                                                                                                   | test   | `make test`                                                                                                                                                                                                                                                       | covered |
| TC-FUTURE-017 | AC-FUTURE-026                               | `future_settle_test.go` (`TestOutcome_SettleResolveNoHandlers`)                                              | test   | `go test -count=1 -run '^TestOutcome_SettleResolveNoHandlers' .`                                                                                                                                                                                                  | missing |
| TC-FUTURE-018 | AC-FUTURE-028                               | `future_accessor_test.go` (`TestOutcome_State`)                                                              | test   | `go test -count=1 -run '^TestOutcome_State' .`                                                                                                                                                                                                                    | missing |
| TC-FUTURE-019 | AC-FUTURE-025                               | `future_test.go` (`TestFuture_UntypedNilHandlerNoop`)                                                        | test   | `go test -count=1 -run '^TestFuture_UntypedNil' .`                                                                                                                                                                                                                | missing |
| TC-FUTURE-020 | AC-FUTURE-027                               | `future_reject_test.go` (`TestOutcome_SettleRejectNoHandlers`)                                               | test   | `go test -count=1 -run '^TestOutcome_SettleRejectNoHandlers' .`                                                                                                                                                                                                   | missing |
| TC-FUTURE-021 | AC-FUTURE-029                               | `future_results_test.go` (`TestOutcome_Results`)                                                             | test   | `go test -count=1 -run '^TestOutcome_Results' .`                                                                                                                                                                                                                  | covered |
| TC-FUTURE-022 | AC-FUTURE-030                               | `future_err_test.go` (`TestOutcome_Err`)                                                                     | test   | `go test -count=1 -run '^TestOutcome_Err' .`                                                                                                                                                                                                                      | covered |
| TC-FUTURE-023 | AC-FUTURE-031                               | `future_settle_once_test.go` (`TestSettleOnce_Sequential`)                                                   | test   | `go test -count=1 -run '^TestSettleOnce_Sequential' .`                                                                                                                                                                                                            | missing |
| TC-FUTURE-024 | AC-FUTURE-032                               | `future_settle_once_test.go` (`TestSettleOnce_Concurrent`)                                                   | test   | `go test -count=1 -race -timeout 30s -run '^TestSettleOnce_Concurrent' .`                                                                                                                                                                                         | missing |
| TC-FUTURE-025 | AC-FUTURE-033                               | `future_settle_once_test.go` (`TestSettleOnce_HandlerReadsOutcome`)                                          | test   | `go test -count=1 -timeout 30s -run '^TestSettleOnce_HandlerReadsOutcome' .`                                                                                                                                                                                      | missing |
| TC-FUTURE-026 | AC-FUTURE-034                               | `future_variadic_test.go` (`TestVariadic_SuccessChain`)                                                      | test   | `go test -count=1 -run '^TestVariadic_SuccessChain' .`                                                                                                                                                                                                            | missing |
| TC-FUTURE-027 | AC-FUTURE-035                               | `future_variadic_test.go` (`TestVariadic_FailureAndChain`)                                                   | test   | `go test -count=1 -run '^TestVariadic_FailureAndChain' .`                                                                                                                                                                                                         | missing |
| TC-FUTURE-028 | AC-FUTURE-036                               | `future_register_atomic_test.go` (`TestRegister_AtomicSuccess`)                                               | test   | `go test -count=1 -run '^TestRegister_AtomicSuccess' .`                                                                                                                                                                                                           | missing |
| TC-FUTURE-029 | AC-FUTURE-037                               | `future_register_atomic_test.go` (`TestRegister_AtomicThenAndFail`)                                           | test   | `go test -count=1 -run '^TestRegister_AtomicThenAndFail' .`                                                                                                                                                                                                       | missing |

`TC-FUTURE-012` is the one row whose `covered` status comes from an actual executed run, not from an
`nfr-audit` pass. The evidence was produced in the run checkout
`/media/saif/EXTRA/workspaces/future/.worktrees/ac-future-018/_run` at commit `e5783cf`, with the
mandatory writable Go environment prefix (`GOTMPDIR`, `GOCACHE`, `GOMODCACHE`, `GOPATH` under
`workflow.local/.go`, `GOFLAGS=-mod=mod`): the `Check` above ran 3/3 `TestRace_*` green
(`ok github.com/noypi/future 1.203s`), and the full suite `go test -count=1 -v ./...` passed 6/6 (the
three pre-existing `TestFuture_*` tests plus the three new `TestRace_*` tests), exit 0. That checkout
carries `go.mod`/`go.sum` adopted byte-identical from `run/nfr-future-003`; `NFR-FUTURE-003` has since
landed on `master` (PR #1 merged as `095bc13`), so the workspace root now carries `go.mod`, `go.sum`,
and the `Makefile`. This row's evidence is still that run-checkout execution at commit `e5783cf`, not a
re-run of the same root-level command, and it makes no new root-level claim. `TC-FUTURE-012` covers an
FR acceptance criterion rather than an NFR, so its `Status` is owned by the recorded test run above;
`nfr-audit` reports the row only for information (the 2026-10-24T13:31:05Z all-rows run also lists it
`exit=0`).

`BenchmarkFuture` in `future_test.go` measures allocation and chain cost; it asserts no behavior and
is not mapped to a criterion.

## Code Map

| File             | Role                                                                                      |
| ---------------- | ----------------------------------------------------------------------------------------- |
| `future.go`      | implementation — exported API, reflection fitting, panic recovery                         |
| `future_test.go` | unit-test — three table-less tests plus one benchmark, package `future_test`              |
| `README.md`      | doc — feature list and a usage example                                                    |
| `LICENSE`        | license                                                                                   |
| `go.mod`         | build — present at the repository root; `NFR-FUTURE-003` (`AC-FUTURE-022`)                |
| `go.sum`         | build — present at the repository root; `NFR-FUTURE-003` (`AC-FUTURE-023`)                |
| `Makefile`       | build — present at the repository root; `test` target (`NFR-FUTURE-003`, `AC-FUTURE-024`) |

## Conflicts

- **CONFLICT-FUTURE-001** — within `SPEC-FUTURE`, between `RULE-FUTURE-004` and `FR-FUTURE-014` /
  `AC-FUTURE-026`, status `resolved`, resolution `supersedes`. `RULE-FUTURE-004` described wrappers
  created lazily from the first handler, under which a handler-less settle panicked; `FR-FUTURE-014`
  requires executor-typed wrappers created when the promise is created. Resolved in 1.14 by
  replacing the rule text and recording the decision in `Q-FUTURE-008`.
- **CONFLICT-FUTURE-002** — within `SPEC-FUTURE`, between `AC-FUTURE-011` and `AC-FUTURE-026`,
  status `resolved`, resolution `scope assignment`. `AC-FUTURE-011` records the last handler's
  return values as results; `AC-FUTURE-026` covers the no-handler case, where there is no handler
  chain to take a last value from. The empty chain records the settle arguments instead, as stated
  in the `FR-FUTURE-014` Notes; the two criteria apply to disjoint conditions.

## Open Questions

- **Q-FUTURE-001** (`needs-decision`) — `README.md` documents `SetCatch( recoverfn func(...interface{}) )`,
  but the code requires `func(error, interface{})`. Which signature is the public contract? The
  code is the runtime truth and `FR-FUTURE-009` records it; the README is a stale doc. Settling it
  needs a maintainer confirmation, then a doc fix or a signature change.
- **Q-FUTURE-002** (`resolved`) — For `exec(true)`, the settlement block (recover, catch, finally)
  now runs inside the execution closure on the executor goroutine, so `finally` runs after
  settlement and observes `FinalRecovered`; `Wait` returns after `Done`. **Resolved in run
  `ac-future-015` (commit `8c2b993`):** the async `TestRecovered_AsyncState*` tests record
  `finally observation: calls=1 state=3 args=[]` after `Wait` returns (see `unit-tests.md`), where
  `state=3` is the `FinalRecovered` enum value. This is a code fix against unchanged requirement
  text: `FR-FUTURE-011` and `AC-FUTURE-016` are untouched.
- **Q-FUTURE-003** (`resolved`) — `Race` self-deadlocked: the settle callback took `RLock` and then
  `Lock` on the same `sync.RWMutex` before unlocking, so every `Race` call with at least one
  settling promise hung and `AC-FUTURE-018` was not satisfiable by that code. **Resolved in run
  `ac-future-018` (commit `e5783cf`):** the settle callback now sends once on a buffered result
  channel guarded by `sync.Once` (latest loser settlements are no-ops), so the first promise to
  resolve or reject is returned after `Wait`. Evidence: `race_test.go`
  (`TestRace_ResolvesFirstToSettle`, `TestRace_RejectsFirstToSettle`, `TestRace_SinglePromise`)
  failed 3/3 with `Race did not return within 3s (deadlock)` against the old `future.go` and passed
  3/3 after the fix in the run checkout, and the full suite passed 6/6; the green command is
  `go test -count=1 -v -timeout 60s -run '^TestRace_' .` (see the `TC-FUTURE-012` note under
  `## Verification` for its provenance and environment dependency). This is a code fix against
  unchanged requirement text: `FR-FUTURE-013` and `AC-FUTURE-018` already stated the behavior.
- **Q-FUTURE-004** (`resolved`) — The repository has no `go.mod` and imports
  `github.com/pkg/errors` (production) and `github.com/stretchr/testify` (tests); in this
  environment neither is resolvable and the Go build temp directory is on a read-only filesystem.
  `go build .` fails with `cannot find package "github.com/pkg/errors"` and `go build ./...` fails
  with `directory prefix . does not contain main module`. Every test in `future_test.go` is
  therefore unexecuted here. **Decision (2026-10-23):** add `go.mod`, `go.sum`, and a `Makefile`
  whose `test` target runs the suite, as required by `NFR-FUTURE-003` (`AC-FUTURE-022` to
  `AC-FUTURE-024`). Settling it needs the three files created, `go build ./...` and
  `go mod verify` to succeed, and `make test` to run to completion; the coverage rows above can
  then be promoted from "asserted in source" to an actual pass/fail result, and `TC-FUTURE-014` to
  `TC-FUTURE-016` from `missing` to `covered`. **Resolved (2026-10-24):** the three files now exist at
  the repository root, and the 2026-10-24T12:34:47Z `nfr-audit` run passed `go build ./...`,
  `go mod verify`, and `make test` with `exit=0`, so `TC-FUTURE-014` to `TC-FUTURE-016` are `covered`
  and the package builds and its suite runs here.
- **Q-FUTURE-005** (`resolved`) — An untyped `nil` handler (`q.OnSuccess(nil)`) used to reach
  `reflect.TypeOf(nil)`, which returns `nil`, and then dereference it — a nil-pointer panic rather
  than the documented `param should be a function` panic. **Resolved (2026-10-24, instruction):**
  an untyped `nil` is treated like a typed nil and ignored (`AC-FUTURE-025`, `FR-FUTURE-005`), so
  registration no longer panics. `TC-FUTURE-019` is the pending verification row.
- **Q-FUTURE-006** (`needs-evidence`) — A panic inside a success or failure handler is recovered by
  the same `defer` in `exec`, so handlers registered after the panicking one do not run and
  `catch` receives the wrapped "Promise Recovered" error. This is inferred from the single
  `defer`/`recover` around the whole call chain; a test would confirm it.
- **Q-FUTURE-007** (`unverified`) — `RULE-FUTURE-003` (calling `exec` twice re-runs the executor
  and the registered handlers) is inferred from the absence of a guard; no test or doc states
  whether re-execution is supported. `FR-FUTURE-016` makes the settlement latch per promise, so a
  second `exec` re-runs the executor and its handlers but cannot change the recorded outcome;
  whether re-execution is supported at all stays unverified.
- **Q-FUTURE-008** (`resolved`) — An executor that settles without any handler registered could
  not succeed: the resolve/reject wrapper was a zero `reflect.Value` until the first
  `OnSuccess`/`OnFail` call, so `resolve("x")` panicked in `reflect.Call` and the promise ended
  `FinalRecovered` instead of `FinalResolved`. **Resolved (2026-10-24, instruction):** the package
  supplies the wrappers, derived from the executor's parameter types at `FutureDeferred`
  (`FR-FUTURE-014`, `AC-FUTURE-026`/`AC-FUTURE-027`); `RULE-FUTURE-004` was replaced accordingly.
  `TC-FUTURE-017` is the pending verification row.

## Delivery Estimates

> Non-normative planning estimate. It records expected delivery effort for the `/build` pipeline, not
> a requirement or a guarantee. The model named in the request, "DeepSeek 4.1 (low)", is not measured
> in this repository; the ranges below are engineering estimates from the number of files, tests, and
> reflection/concurrency risk, and must not be quoted as benchmark results.

`/build` runs Preflight → planner → batch tester (tests + contracts) → coder (per batch) →
code-reviewer → security-reviewer → acceptance tester → Close (`nfr-audit`, reconciliation, PR).
Estimates assume one feature per run, one batch, the existing Go environment prefix (`GOTMPDIR`,
`GOCACHE`, `GOMODCACHE`, `GOPATH` under `workflow.local/.go`), and one review fix cycle.

| Feature | Requirements | Stages (expected minutes) | Total | Fits 30 min? |
| ------- | ------------ | ------------------------- | ----- | ------------ |
| Feature 1 — lifecycle hardening | `FR-FUTURE-010`, `FR-FUTURE-011`, `FR-FUTURE-012`, `NFR-FUTURE-002` (existing; no new ID) | preflight 3, planner 5, tester 10, coder 15, code-review 5, security-review 3, acceptance 7, Close 5 | 45–65 min | No; only the coder slice (recover inside the worker) fits 15–25 min, and that part is already implemented in run `ac-future-015` |
| Feature 2 — observable outcome | `FR-FUTURE-014`, `FR-FUTURE-015`, `AC-FUTURE-025` (new) | preflight 3, planner 5, tester 15, coder 25, code-review 7, security-review 3, acceptance 8, Close 6 | 70–100 min | No; the coder-sized slice alone is 20–30 min |
| Feature 3 — settle-once latch | `FR-FUTURE-016` (new) | preflight 3, planner 5, tester 10, coder 15, code-review 5, security-review 3, acceptance 5, Close 4 | 45–55 min | No; one batch, and a clean first review keeps it near 45 |
| Feature 4 — variadic fitting | `FR-FUTURE-017` (new) | preflight 3, planner 4, tester 10, coder 12, code-review 5, security-review 2, acceptance 4, Close 4 | 40–48 min | No; one function plus one test file |
| Feature 5 — atomic registration | `FR-FUTURE-018` (new) | preflight 3, planner 4, tester 10, coder 10, code-review 5, security-review 2, acceptance 4, Close 4 | 38–45 min | No; three methods plus one test file |

- Feature 1 adds no requirement: `FR-FUTURE-010`, `FR-FUTURE-011`, and `FR-FUTURE-012` already
  specify asynchronous recovery, `finally` after settlement, and `Wait`; `NFR-FUTURE-002` owns
  race-freedom. The remaining work is a conformance fix plus the `-race` evidence, which is why its
  estimate is smaller than Feature 2's.
- The workspace root at `13c22ea` still carries the pre-`ac-future-015` `exec`; Feature 1's estimate
  assumes that run's fix (commit `8c2b993`) is merged first. If it must be ported again, add 20–30
  minutes.
- Feature 2 is two batches if the tester splits settlement (`FR-FUTURE-014`) from the accessors
  (`FR-FUTURE-015`); splitting adds one coder dispatch and roughly 10–15 minutes but lowers the risk
  of a reflection regression.
- A 30-minute budget covers a single stage, not a feature. On the earlier question, only a coder-only
  slice — not a full `/build` run — fits that window.
- The 1.15 row split gives each new acceptance criterion its own test file, so `AC-FUTURE-025`,
  `-026`, and `-028` can run as three isolated single-AC pipelines without a test-file collision;
  the `future.go` hunks still overlap at merge time.
- Feature 3 has no skeleton to port: the latch is a new unexported field plus one claim helper, and
  its three tests share one file, so it is a single batch of one item. The 45–55 minute range
  assumes one review cycle; a race-detector failure in the concurrency test adds 10–15 minutes. Its
  three verification rows start `missing`, so the estimate is for delivery, not for promoted coverage.
- Features 3 to 5 are disjoint in `future.go` and can run as three concurrent `/build` runs from the
  same base: Feature 3 edits the `Promise` struct plus `resolvedWrapped`/`rejectedWrapped` (lines
  13–33 and 253–265), Feature 4 edits `fitFnArgs` only (line 296), and Feature 5 edits `Then`,
  `OnSuccess`, `OnFail`, `onsuccess`, and `onfail` (lines 79–151). Only Feature 3 adds fields, so a
  merge at that block is the one expected conflict; merge order is 3, 4, 5. No feature depends on
  `FR-FUTURE-014`, and none requires the other two's code.

## Changelog

- **1.19** (2026-10-25, added) — Added `FR-FUTURE-017` (a variadic handler receives the settled
  values packed into its variadic parameter, `AC-FUTURE-034`/`AC-FUTURE-035`) and `FR-FUTURE-018`
  (a registration call is atomic across the handlers it names, `AC-FUTURE-036`/`AC-FUTURE-037`),
  with pending rows `TC-FUTURE-026` to `TC-FUTURE-029`. Added the non-normative Feature 4 (40–48
  min) and Feature 5 (38–45 min) rows to `## Delivery Estimates` and a note recording that
  Features 3 to 5 edit disjoint regions of `future.go` and may run as concurrent `/build` runs.
  No existing requirement, acceptance criterion, `Covers`, `File`, `Method`, `Check`, or row
  `Status` changed; all four new rows are `missing`. No production code, test file, or
  `workflow.local/` artifact was modified. `Verified` stays 13 and the index `Requirements` count
  moves 18 to 20.
- **1.18** (2026-10-25, added) — Added `FR-FUTURE-016` (a promise settles at most once; the first
  settlement wins) with `AC-FUTURE-031` (a later settle callback is a normal-returning no-op),
  `AC-FUTURE-032` (concurrent settle callbacks settle exactly once, without a data race), and
  `AC-FUTURE-033` (a handler reading `Results()`/`Err()` neither deadlocks nor panics), plus three
  pending verification rows `TC-FUTURE-023` to `TC-FUTURE-025` in `future_settle_once_test.go`.
  Rewrote `RULE-FUTURE-001` from "nothing enforces it" to the latch, and recorded in
  `RULE-FUTURE-003` and `Q-FUTURE-007` that the latch is per promise so a second `exec` cannot
  settle; extended the `NFR-FUTURE-002` note to include concurrent settle attempts. Added the
  non-normative Feature 3 row (45–55 min) to `## Delivery Estimates`. No existing requirement,
  acceptance criterion, `Covers`, `File`, `Method`, or `Check` changed, no row's `Status` changed
  (all three new rows are `missing`), the Data Model table is untouched because the latch field is
  unexported and named at implementation, and no production code, test file, or `workflow.local/`
  artifact was modified. `Verified` stays 13 and the index `Requirements` count moves 17 to 18.
- **1.17** (2026-10-25, updated) — Close reconciliation of verification row `TC-FUTURE-021`
  (`AC-FUTURE-029`, `FR-FUTURE-015`) from `nfr-audit run --all-rows` at 2026-10-25T06:07:02Z over the
  `ac-future-029` run checkout
  (`/media/saif/EXTRA/workspaces/future/.worktrees/ac-future-029/_run`, commit `2b4706f`, scope
  `all`, 22 rows: 15 pass, 6 unknown, 1 manual, 0 fail; `nfr-audit status` reports `stale: no`),
  where the row's `verdict=pass`, `detail=exit=0` and
  `go test -count=1 -run '^TestOutcome_Results' .` exits 0 with `TestOutcome_Results` present in
  `future_results_test.go` and passing eight named subtests: the row's `Status` moved from `missing`
  to `covered`. `TC-FUTURE-017` to `TC-FUTURE-020` also report `verdict=pass`, but those passes are
  vacuous — `future_settle_test.go`, `future_accessor_test.go`, and `future_reject_test.go` do not
  exist in the run checkout, `future_test.go` has no `TestFuture_UntypedNil*` function, and
  `go test -run '<regex>' .` exits 0 when nothing matches — so all four deliberately stay `missing`;
  promoting them would claim coverage for tests that do not exist. `TC-FUTURE-022` was already
  `covered` and is unchanged. The Verification intro now cites this run, its counts, and the
  vacuous-pass reason. No requirement text, acceptance criterion, `Covers`, `File`, `Method`,
  `Check`, or any other row's `Status` changed, and the Data Model table is untouched. `Verified`
  moves 12 to 13.
- **1.16** (2026-10-25, updated) — Close reconciliation of verification row `TC-FUTURE-022`
  (`AC-FUTURE-030`, `FR-FUTURE-015`) from `nfr-audit run --all-rows` at 2026-10-25T04:10:40Z over
  the `ac-future-030` run checkout
  (`/media/saif/EXTRA/workspaces/future/.worktrees/ac-future-030/_run`, commit `ccaacd0`, scope
  `all`, 22 rows: 15 pass, 6 unknown, 1 manual, 0 fail, `uncovered_acs: 0`; `nfr-audit status`
  reports `stale: no`), where the row's `verdict=pass`, `detail=exit=0` and
  `go test -count=1 -run '^TestOutcome_Err' .` exits 0 with `TestOutcome_Err` present in
  `future_err_test.go`: the row's `Status` moved from `missing` to `covered`, giving
  `FR-FUTURE-015` its first covered row. `TC-FUTURE-017` to `TC-FUTURE-021` also report
  `verdict=pass`, but those passes are vacuous — `future_settle_test.go`,
  `future_accessor_test.go`, `future_reject_test.go`, and `future_results_test.go` do not exist in
  the run checkout, `future_test.go` has no `TestFuture_UntypedNil*` function, and
  `go test -run '<regex>' .` exits 0 when nothing matches — so all five deliberately stay
  `missing`; promoting them would claim coverage for tests that do not exist. The Verification
  intro now cites this run and its counts and names the vacuous-pass reason. No requirement text,
  acceptance criterion, `Covers`, `File`, `Method`, `Check`, or any other row's `Status` changed,
  and the Data Model table is untouched (the new unexported `err` field is a known follow-up).
  `Verified` moves 11 to 12.
- **1.15** (2026-10-25, updated) — Split the combined verification rows added in 1.14 into one row
  per acceptance criterion, each in its own test file, so a single-`AC-…` selector run can close its
  own row: `TC-FUTURE-017` now covers only `AC-FUTURE-026` (`future_settle_test.go`),
  `TC-FUTURE-018` only `AC-FUTURE-028` (`future_accessor_test.go`), and new `TC-FUTURE-020`,
  `-021`, and `-022` cover `AC-FUTURE-027` (`future_reject_test.go`), `AC-FUTURE-029`
  (`future_results_test.go`), and `AC-FUTURE-030` (`future_err_test.go`); `TC-FUTURE-019`
  (`AC-FUTURE-025`) is unchanged. All six remain `missing`. No requirement, acceptance criterion,
  or `Status` changed; `Verified` stays 11.
- **1.14** (2026-10-24, added) — Added `FR-FUTURE-014` (handler-less settlement,
  `AC-FUTURE-026`/`AC-FUTURE-027`), `FR-FUTURE-015` (outcome accessors `State`, `Results`, `Err`,
  `AC-FUTURE-028` to `AC-FUTURE-030`), `AC-FUTURE-025` (untyped `nil` handler) under
  `FR-FUTURE-005`, three interfaces (`API-FUTURE-010` to `API-FUTURE-012`), and three pending
  verification rows (`TC-FUTURE-017` to `TC-FUTURE-019`, all `missing`). Replaced
  `RULE-FUTURE-004` with the executor-typed-wrapper rule, resolved `Q-FUTURE-005` and
  `Q-FUTURE-008` at the user's instruction, and recorded `CONFLICT-FUTURE-001` and
  `CONFLICT-FUTURE-002`. Added the non-normative `## Delivery Estimates` section for the two
  features (Feature 1: 45–65 min; Feature 2: 70–100 min under `/build`). No existing requirement,
  acceptance criterion, or `covered` row changed; `Verified` stays 11.
- **1.13** (2026-10-24, updated) — Close reconciled verification row `TC-FUTURE-010` to `covered` from
  the 2026-10-24T15:32:29Z `nfr-audit run --all-rows` over the `ac-future-015` run checkout
  (`/media/saif/EXTRA/workspaces/future/.worktrees/ac-future-015/_run`, scope `all`, 16 rows,
  `rows_pass: 9`, `rows_fail: 0`, `rows_external: 0`, `rows_manual: 1`, `rows_unknown: 6`), where the
  row's `verdict=pass`, `detail=exit=0`: `FR-FUTURE-010` gained its first covered row. The Verification
  intro now cites that run and its counts, naming the passing rows `TC-FUTURE-002`, `TC-FUTURE-008`,
  `TC-FUTURE-009`, `TC-FUTURE-010`, `TC-FUTURE-011`, `TC-FUTURE-012`, `TC-FUTURE-014`,
  `TC-FUTURE-015`, and `TC-FUTURE-016`, with `TC-FUTURE-013` still `manual`; the obsolete parenthetical
  saying the cited run predated the row's `Method` declaration was dropped, and the six-row no-`Method`
  sentence (`TC-FUTURE-001`, `TC-FUTURE-003` to `TC-FUTURE-007`) is unchanged. Nothing else changed: no
  requirement, acceptance criterion, `Covers`, `File`, `Method`, `Check`, or other row's `Status`.
- **1.12** (2026-10-24, updated) — Declared how `TC-FUTURE-010` (`AC-FUTURE-015`,
  `FR-FUTURE-010`) is proven after run `ac-future-015` (commit `8c2b993`) published eleven
  `TestRecovered_*` tests across `future_state_test.go` (4) and `future_test.go` (7): the row's
  `File` names those two files and the eleven tests, `Method` is `test`, and `Check` is
  `go test -count=1 -run '^TestRecovered_' .`, which the run's `unit-tests.md` records as 11/11
  `PASS`, exit 0. `Status` deliberately stays `missing` — the run's Close audit derives it.
  `Q-FUTURE-002` moved from `needs-evidence` to `resolved`: the settlement block now runs in the
  execution closure on the executor goroutine for `exec(true)`, so `finally` runs after settlement
  and observes `FinalRecovered`, and `Wait` returns after `Done`; the async
  `TestRecovered_AsyncState*` tests record `finally observation: calls=1 state=3 args=[]` after
  `Wait` (run `ac-future-015`, commit `8c2b993`). The Verification intro's no-`Method` count moved
  from seven rows to six (`TC-FUTURE-001`, `TC-FUTURE-003` to `TC-FUTURE-007`); the cited
  2026-10-24T15:18:05Z `ac-future-017` run and its historical counts and provenance are unchanged,
  with a parenthetical noting that run predates this declaration. No requirement, acceptance
  criterion, `Covers`, other `Method`/`Check`, or other row's `Status` changed, and `Verified` stays
  10 because the row is still `missing`.
- **1.11** (2026-10-24, updated) — Close reconciled verification row `TC-FUTURE-011` to `covered` from
  the 2026-10-24T15:18:05Z `nfr-audit run --all-rows` over the `ac-future-017` run checkout
  (`/media/saif/EXTRA/workspaces/future/.worktrees/ac-future-017/_run`, scope `all`, 16 rows, 8 pass,
  1 manual, 7 unknown), where the row's `verdict=pass`, `detail=exit=0`: `FR-FUTURE-012` gained its
  first covered row. The Verification intro now cites that run and its counts, naming the passing rows
  `TC-FUTURE-002`, `TC-FUTURE-008`, `TC-FUTURE-009`, `TC-FUTURE-011`, `TC-FUTURE-012`,
  `TC-FUTURE-014`, `TC-FUTURE-015`, and `TC-FUTURE-016`. Nothing else changed: no requirement,
  acceptance criterion, `Covers`, `File`, `Method`, `Check`, or other row's `Status`.
- **1.10** (2026-10-24, updated) — Declared how `TC-FUTURE-011` (`AC-FUTURE-017`, `FR-FUTURE-012`) is
  proven after run `ac-future-017` published `TestFuture_WaitBlocksForAsyncExec` and
  `TestFuture_WaitReturnsAfterSyncExec` in `future_test.go`: the row's `File` names those two tests,
  `Method` is `test`, and `Check` is `go test -count=1 -run '^TestFuture_Wait' .`, which the run's
  `test-report.md` records as 2/2 `PASS`, exit 0. `Status` deliberately stays `missing` — the run's
  Close reconciles the row from the audit. The Verification intro's no-`Method` count moved from eight
  rows to seven (it now ends at `TC-FUTURE-010` instead of `TC-FUTURE-010` to `TC-FUTURE-011`); the
  cited 2026-10-24T14:39:29Z run and its historical counts and provenance are unchanged and no new run
  was invented. No requirement, acceptance criterion, `Covers`, `Status`, or other row changed.
- **1.9** (2026-10-24, updated) — Reconciled verification row `TC-FUTURE-009` with the
  2026-10-24T14:39:29Z `nfr-audit run --all-rows` (workspace and code root
  `/media/saif/EXTRA/workspaces/future`, HEAD `13c22ea`, scope `all`, 16 rows): the row's `Status`
  moved from `missing` to `covered` on `verdict=pass`, `detail=exit=0`. The Verification intro now
  cites that run and its counts (7 pass, 1 manual, 8 unknown), naming the passing rows
  `TC-FUTURE-002`, `TC-FUTURE-008`, `TC-FUTURE-009`, `TC-FUTURE-012`, `TC-FUTURE-014`,
  `TC-FUTURE-015`, and `TC-FUTURE-016`. No requirement, acceptance criterion, `Covers`, `File`,
  `Method`, `Check`, or other row's `Status` changed, and the `TC-FUTURE-012` provenance paragraph
  is unchanged.
- **1.8** (2026-10-24, updated) — Declared how `TC-FUTURE-009` evidence is produced after run
  `ac-future-014` (commit `bb61245`; PR #5 merged as `72fc530`) added three `AC-FUTURE-014` tests to
  `future_test.go`: `TestFuture_DefaultCatchLogsRecoveredValue`,
  `TestFuture_SetCatchReplacesDefault`, and `TestFuture_SetFinallyReplacesDefault`. `File` is
  `future_test.go` naming those three tests, `Method` is `test`, and `Check` is
  `go test -count=1 -run '^TestFuture_' .`, which runs every `TestFuture_*` in the package. That
  run's `test-report.md` records `PASS` 3/3 on the targeted command and 9/9 on the full suite.
  `Status` deliberately stays `missing` — the row is not `covered` until a passing
  `nfr-audit run --all-rows` records it. The Verification intro's no-`Method` count moved from nine
  rows to eight (it now names `TC-FUTURE-010` to `TC-FUTURE-011` instead of `TC-FUTURE-009` to
  `TC-FUTURE-011`). No requirement, acceptance criterion, `Covers`, or other row changed.
- **1.7** (2026-10-24, updated) — Reconciled verification row `TC-FUTURE-008` with the
  2026-10-24T13:31:05Z `nfr-audit run --all-rows` (workspace and code root
  `/media/saif/EXTRA/workspaces/future`, branch `feat/test-branch` at `1f71353`): the row's `Status`
  moved from `missing` to `covered` on `verdict=pass`, `detail=exit=0`, and the Verification intro
  now cites that run and its counts. No requirement, acceptance criterion, `Covers`, `Method`,
  `Check`, or `File` value changed.
- **1.6** (2026-10-24, updated) — Declared how `TC-FUTURE-008` evidence is produced after run
  `ac-future-013` added `future_state_test.go` (`TestPromiseState_UnknownUntilCallback`,
  `TestPromiseState_ResolvedBeforeSuccessHandlers`, `TestPromiseState_RejectedBeforeFailureHandlers`;
  PR #3 merged as `1f71353`): `File` is `future_state_test.go`, `Method` is `test`, and `Check` is
  `go test -count=1 -run '^TestPromiseState_' .`. `Status` deliberately stays `missing` — the row is
  not `covered` until a passing `nfr-audit run --all-rows` records it. No requirement, acceptance
  criterion, `Covers`, or other row changed.
- **1.5** (2026-10-24, updated) — Recorded the full `/verify-nfr all` audit run at
  2026-10-24T12:34:47Z: verification row `TC-FUTURE-002` moved from `missing` to `covered`
  (`go test -run TestFuture_resolveFuncSignatures ./...`, `exit=0`), and the stale records were
  reconciled with the landed `NFR-FUTURE-003` (PR #1, `095bc13`), which put `go.mod`, `go.sum`, and
  the `Makefile` at the repository root: the Verification intro, the `TC-FUTURE-012` provenance
  paragraph, the Code Map build rows, and `Q-FUTURE-004` (now `resolved`). No requirement, acceptance
  criterion, `Method`, `Check`, `File`, or `Covers` value changed.
- **1.4** (2026-10-24, updated) — Recorded the 2026-10-24T12:31:39Z `nfr-audit` run: verification
  rows `TC-FUTURE-014` to `TC-FUTURE-016` moved from `missing` to `covered` (`go build ./...`,
  `go mod verify`, and `make test` each `exit=0`) after `NFR-FUTURE-003` landed, and the Verification
  intro was corrected to match. No requirement, acceptance criterion, `Method`, `Check`, or ID changed.
- **1.3** (2026-10-24, updated) — Reconciled the `Race` descriptive records with the implemented fix
  from run `ac-future-018` (commit `e5783cf`): `FR-FUTURE-013` Notes now describe the buffered result
  channel plus `sync.Once` settle callback and point at `race_test.go`; `API-FUTURE-009` Errors no
  longer claims the implementation blocks indefinitely; `Q-FUTURE-003` is marked resolved with the
  red/green evidence and the green command; verification row `TC-FUTURE-012` records `race_test.go`,
  `Method: test`, the green command, and `Status: covered`, with its run-checkout provenance noted in
  the Verification section. No requirement, acceptance criterion, `RULE-*`, Data Model, or Code Map
  text or ID changed.
- **1.2** (2026-10-24, added) — Verification table: added Method and Check columns for the five
  NFR-covering rows (TC-FUTURE-002, -013, -014, -015, -016) per approved /verify-nfr decisions.
- **1.1** (2026-10-23, added) — Added `NFR-FUTURE-003` and `AC-FUTURE-022` to `AC-FUTURE-024`
  for the module manifest and the `make test` entry point, at the user's instruction, to settle
  `Q-FUTURE-004`. Added `TC-FUTURE-014` to `TC-FUTURE-016` (all `missing`), extended the build
  scope, and updated the code map. No existing requirement, criterion, or ID was changed.
- **1.0** (2026-10-23, created) — Initial spec derived from `future.go`, `future_test.go`,
  `README.md`, and `git log`. 13 functional and 2 non-functional requirements recorded; 8 open
  questions raised, including the `Race` deadlock, the handler-before-execute requirement, and the
  missing module file.
