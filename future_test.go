package future_test

import (
	"log"
	"testing"

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

func BenchmarkFuture(b *testing.B) {
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
