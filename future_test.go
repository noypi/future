package future_test

import (
	"bytes"
	"log"
	"strings"
	"sync"
	"testing"
	"time"

	. "github.com/noypi/future"
	assertpkg "github.com/stretchr/testify/assert"
)

func TestFuture_x01(t *testing.T) {
	assert := assertpkg.New(t)

	exec, q :=
		FutureDeferred(func(
			resolve func(string) string,
			rejected func(interface{})) {
			resolve("message 1")
		})

	var results []string

	q.Then(func(msg string) string {
		log.Println("received msg=", msg)
		results = append(results, msg)
		return "message 2"
	}, func(fail interface{}) {

	}).Then(func(msg string) string {
		log.Println("received msg=", msg)
		results = append(results, msg)
		return ""
	}, func(fail interface{}) {

	})

	exec(false)

	assert.Equal(2, len(results))
	assert.Equal("message 1", results[0])
	assert.Equal("message 2", results[1])
}

func TestFuture_resolveFuncSignatures(t *testing.T) {
	assert := assertpkg.New(t)

	exec, q := FutureDeferred(func(resolve func(string) string, rejected func(interface{})) {
		resolve("message 1")
	})

	var results []interface{}

	q.Then(func(msg string) string {
		log.Println("received msg=", msg)
		results = append(results, msg)
		return "message 2"
	}, func(fail interface{}) {

	}).Then(func(msg, none1 string, none2 int) (ret1, ret2 string) {
		log.Println("received msg=", msg)
		results = append(results, msg)
		results = append(results, none1)
		results = append(results, none2)
		return "ret1", "ret2"
	}, func(fail interface{}) {

	}).Then(func(msg1, msg2 string) int {
		log.Println("received msg1=", msg1, ", msg2=", msg2)
		results = append(results, msg1)
		results = append(results, msg2)
		return 3
	}, func(fail interface{}) {

	})

	exec(false)

	assert.Equal(6, len(results))
	assert.Equal("message 1", results[0])
	assert.Equal("message 2", results[1])
	assert.Equal("", results[2])
	assert.Equal(0, results[3])
	assert.Equal("ret1", results[4])
	assert.Equal("ret2", results[5])
}

func TestFuture_rejectedFuncSignatures(t *testing.T) {
	assert := assertpkg.New(t)

	exec, q :=
		FutureDeferred(func(
			resolve func(string) string,
			rejected func(string) int) {

			rejected("reason 1")
		})

	resolved := func(msg string) string {
		return ""

	}

	var results []interface{}

	q.Then(resolved, func(msg string) int {
		log.Println("reason =", msg)
		results = append(results, msg)
		return 8

	}).Then(resolved, func(msg int) (string, string) {
		log.Println("reason =", msg)
		results = append(results, msg)
		return "reason 3", "reason 4"

	}).Then(resolved, func(msg1, msg2 string) string {
		log.Println("reason =", msg1, ", reason2=", msg2)
		results = append(results, msg1)
		results = append(results, msg2)
		return "reason 5"

	}).Then(resolved, func(msg1, msg2, msg3 string, msg4, msg5 int) string {
		log.Println("reason =", msg1, ", reason2=", msg2, "...etc")
		results = append(results, msg1)
		results = append(results, msg2)
		results = append(results, msg3)
		results = append(results, msg4)
		results = append(results, msg5)
		return "reason 6"

	})

	exec(false)

	assert.Equal(9, len(results))
	assert.Equal("reason 1", results[0])
	assert.Equal(8, results[1])
	assert.Equal("reason 3", results[2])
	assert.Equal("reason 4", results[3])
	assert.Equal("reason 5", results[4])
	assert.Equal("", results[5])
	assert.Equal("", results[6])
	assert.Equal(0, results[7])
	assert.Equal(0, results[8])
}

// captureDefaultLogger redirects the standard library logger to a buffer for the
// duration of the test, with empty flags and prefix, and restores the previous
// output, flags, and prefix when the test finishes.
func captureDefaultLogger(t *testing.T) *bytes.Buffer {
	t.Helper()

	prevWriter := log.Writer()
	prevFlags := log.Flags()
	prevPrefix := log.Prefix()

	buf := &bytes.Buffer{}
	log.SetOutput(buf)
	log.SetFlags(0)
	log.SetPrefix("")

	t.Cleanup(func() {
		log.SetOutput(prevWriter)
		log.SetFlags(prevFlags)
		log.SetPrefix(prevPrefix)
	})

	return buf
}

// AC-FUTURE-014: by default catch logs the formatted recovery error, " => ", and
// the recovered value through log.Println.
func TestFuture_DefaultCatchLogsRecoveredValue(t *testing.T) {
	assert := assertpkg.New(t)
	logBuf := captureDefaultLogger(t)

	const sentinel = "boom-default-catch"
	exec, q := FutureDeferred(func(resolve func(string) string, rejected func(interface{})) {
		panic(sentinel)
	})

	// Both the resolve and the reject wrapper must exist before exec builds its
	// call arguments. Registering handlers is setup only: the executor panics
	// before either handler can run.
	q.Then(func(s string) string { return s }, func(o interface{}) {})

	assert.NotPanics(func() { exec(false) }, "exec(false) must not let the executor panic escape")

	out := logBuf.String()
	assert.Contains(out, "Promise Recovered", "the default catch must log the recovery error text")
	assert.Contains(out, "=>", "the default catch must log the separation between the error and the value")

	errIdx := strings.Index(out, "Promise Recovered")
	arrowIdx := strings.Index(out, "=>")
	valIdx := strings.Index(out, sentinel)

	assert.GreaterOrEqual(errIdx, 0)
	assert.GreaterOrEqual(arrowIdx, 0)
	assert.GreaterOrEqual(valIdx, 0)
	assert.Less(errIdx, arrowIdx, "the error text must be logged before the \"=>\" separator")
	assert.Less(arrowIdx, valIdx, "the recovered value must be logged after the \"=>\" separator")
}

// AC-FUTURE-014: SetCatch replaces the recovery callback, so the default
// logging callback no longer runs.
func TestFuture_SetCatchReplacesDefault(t *testing.T) {
	assert := assertpkg.New(t)
	logBuf := captureDefaultLogger(t)

	const sentinel = "boom-setcatch"
	exec, q := FutureDeferred(func(resolve func(string) string, rejected func(interface{})) {
		panic(sentinel)
	})

	// Both wrappers must exist before exec builds its call arguments; the
	// executor panics before either handler can run.
	q.Then(func(s string) string { return s }, func(o interface{}) {})

	var calls int
	var gotErr error
	var gotVal interface{}
	q.SetCatch(func(err error, o interface{}) {
		calls++
		gotErr = err
		gotVal = o
	})

	assert.NotPanics(func() { exec(false) }, "exec(false) must not let the executor panic escape")

	assert.Equal(1, calls, "the installed catch callback must run for the recovered panic")
	assert.NotNil(gotErr, "the installed catch callback must receive the recovery error")
	assert.Equal(sentinel, gotVal, "the installed catch callback must receive the recovered value")
	assert.NotContains(logBuf.String(), "Promise Recovered", "SetCatch must replace the default logging callback")
}

// AC-FUTURE-014: SetFinally replaces the completion callback.
func TestFuture_SetFinallyReplacesDefault(t *testing.T) {
	assert := assertpkg.New(t)

	exec, q := FutureDeferred(func(resolve func(string) string, rejected func(interface{})) {
		resolve("value")
	})

	// RULE-FUTURE-004: the resolve wrapper only exists once a success handler
	// has been registered before the executor settles. The reject wrapper is
	// also required because exec passes both wrappers to reflect.Call.
	q.Then(func(s string) string { return s }, func(o interface{}) {})

	called := false
	q.SetFinally(func(st FinalState, args ...interface{}) {
		called = true
	})

	assert.NotPanics(func() { exec(false) }, "exec(false) must not let the executor panic escape")

	assert.True(called, "the installed finally callback must run at completion")
}

func BenchmarkFuture(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		exec, q := FutureDeferred(func(resolve func(string) string, rejected func(interface{})) {
			resolve("message 1")
		})

		q.Then(func(msg string) string {
			return "message 2"
		}, func(fail interface{}) {
		})

		exec(false)
	}
}

// waitBlockWindow is the negative window during which Wait must not return
// while the async executor goroutine is still blocked. It is a scheduling aid,
// not a latency contract.
const waitBlockWindow = 100 * time.Millisecond

// waitReturnTimeout is the liveness bound for Wait returning once the executor
// is released or when no goroutine was started. It is a scheduling aid, not a
// latency contract.
const waitReturnTimeout = 3 * time.Second

// AC-FUTURE-017: after exec(true), Wait returns only once the executor goroutine
// has returned.
func TestFuture_WaitBlocksForAsyncExec(t *testing.T) {
	release := make(chan struct{})
	executorReturned := make(chan struct{})

	exec, q := FutureDeferred(func(resolve func(string) string, reject func(interface{})) {
		<-release
		close(executorReturned)
	})

	// RULE-FUTURE-004 / Q-FUTURE-008: both wrapper values must be non-zero before
	// exec because exec passes both to reflect.Call. Registering handlers is setup
	// only; the executor is released without settling.
	q.Then(func(s string) string { return s }, func(o interface{}) {})

	// A failed assertion must not leave the executor goroutine blocked, so the
	// channel close is idempotent and runs at cleanup as well as in the test body.
	var releaseOnce sync.Once
	releaseFn := func() {
		releaseOnce.Do(func() {
			close(release)
		})
	}
	t.Cleanup(releaseFn)

	exec(true)

	// Do not wait for the executor to start: calling Wait immediately after
	// exec(true) returns is exactly the ordering AC-FUTURE-017 promises.
	waiterReady := make(chan struct{})
	waitReturned := make(chan struct{})
	go func() {
		close(waiterReady)
		q.Wait()
		close(waitReturned)
	}()
	<-waiterReady

	// Negative window: the executor is still blocked on release, so a Wait that
	// returns now does not block for asynchronous execution.
	select {
	case <-waitReturned:
		t.Fatal("Wait returned before the executor was released")
	case <-time.After(waitBlockWindow):
	}

	releaseFn()

	select {
	case <-waitReturned:
	case <-time.After(waitReturnTimeout):
		t.Fatal("Wait did not return after the executor returned")
	}

	// Ordering: executorReturned must already be closed, because Wait can only
	// return after the executor's final statement closed it.
	select {
	case <-executorReturned:
	default:
		t.Fatal("Wait returned before the executor goroutine's final statement ran")
	}
}

// AC-FUTURE-017: after exec(false) no goroutine was started, so Wait returns
// immediately instead of blocking.
func TestFuture_WaitReturnsAfterSyncExec(t *testing.T) {
	exec, q := FutureDeferred(func(resolve func(string) string, reject func(interface{})) {})

	// RULE-FUTURE-004 / Q-FUTURE-008: both wrapper values must be non-zero before
	// exec because exec passes both to reflect.Call. The executor never settles.
	q.Then(func(s string) string { return s }, func(o interface{}) {})

	exec(false)

	waitReturned := make(chan struct{})
	go func() {
		q.Wait()
		close(waitReturned)
	}()

	select {
	case <-waitReturned:
	case <-time.After(waitReturnTimeout):
		t.Fatal("Wait blocked after exec(false); no goroutine was started")
	}
}
