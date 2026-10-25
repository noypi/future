# Design

A `Promise` wraps a function that takes a resolve and a reject callback.

- The callbacks are declared with `reflect` types inferred from the stored function.
- Settlement state is tracked with `FinalState`.
- Handlers registered before settlement are queued and run after it.