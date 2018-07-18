Try to run BenchmarkFuture()

todo: then(funcA).then(funcA2).catch(funcB).finally(???)

# Example
```go

	exec, q :=
		FutureDeferred(func(
			resolve func(string, SomeAction) string,
			rejected func(interface{})) {

			resolve("message 1", SomeAction{})

		})


	q.Then(func(msg string, action SomeAction{}) (string, SomeType) {
		
		// handle msg and action

		return "message 2", SomeType{}
		
	}, func(fail interface{}) {

	}).Then(func(msg string, sometype SomeType) string {
		
		// handle msg and sometype
		return ""
		
	}, func(fail interface{}) {

	})

	bAsync := false
	exec(bAsync)

```
