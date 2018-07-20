| Code Factor  | Docs |
| :----| :----|
| [![CodeFactor][1]][2] | [![CodeFactor][3]][4] |

[1]: https://www.codefactor.io/repository/github/noypi/future/badge/master
[2]: https://www.codefactor.io/repository/github/noypi/future/overview/master
[3]: http://img.shields.io/badge/godoc-reference-5272B4.svg?style=flat-square
[4]: https://godoc.org/github.com/noypi/future

Try to run BenchmarkFuture()

# Why?

- if you don't want repeated casting of interface{} to your desired types (in args and results)
- if you don't want repeated initialization of your channels+goroutines just to make a common pattern
- if you wanted control back on your desired parameters in your resolv or reject functions (when using a 3rd party lib)

# Content

- func Future( anyfunc(resolvfunc, rejectfunc) ) (*Promise)
- func FutureDeferred( anyfunc(resolvfunc, rejectfunc) ) (exec func(bAsync bool), *Promise)
- func Race(...*Promise) (*Promise)
- Promise
	- Then(resolvedfunc, rejectedfunc) (*Promise)
	- OnSuccess(resolvedfunc) (*Promise)
	- OnFail(rejectedfunc) (*Promise)
	- Wait() 
	- SetCatch( recoverfn func(...interface{}) )
	- SetFinally( func(state, ...interface{}) )  // state: {resolved, rejected, recovered}

:: resolvedfunc and rejectedfunc can have any function signature

# Example 
```go
	exec, q := FutureDeferred( func(resolvFn, rejectFn){
		// ... do something
		resolvFn("success")
	})
	
	q.OnSuccess(fn1, fn2, fn3)
	q.OnFail(fn1, fn2)
	
	q.SetCatch( onRecoverFn )
	q.SetFinally( onDone )
	
	bAsync := false
	exec(bAsync)
```

