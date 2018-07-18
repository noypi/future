Try to run BenchmarkFuture()

# Goal

- func Future( anyfunc(resolvfunc, rejectfunc) ) (*Promise)
- func FutureDeferred( anyfunc(resolvfunc, rejectfunc) ) (exec func(bAsync bool), *Promise)
- Promise.Then(resolvedfunc, rejectedfunc) (*Promise)
- Promise.OnSuccess(resolvedfunc) (*Promise)
- Promise.OnError(rejectedfunc) (*Promise)
- Promise.Wait() 
- Promise.SetCatch( recoverfn func(...interface{}) )
- Promise.SetFinally( func(state, ...interface{}) )  // state: {resolved, rejected, recovered}
- func Race(...*Promise) (*Promise)

:: resolvedfunc and rejectedfunc can have any function signature


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
