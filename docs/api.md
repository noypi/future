# API overview

- `Future(fn)` — run `fn` immediately.
- `FutureDeferred(fn)` — return the executor and the promise, without starting it.
- `Then`, `OnSuccess`, `OnFail` — register handlers.
- `State`, `Results`, `Err` — read the settled outcome.