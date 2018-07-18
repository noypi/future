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
