# Known limitations

- The callback pair must match the expected two-function shape.
- The library relies on `reflect`, so argument checking is runtime rather than compile time.
- A promise settles exactly once; later settlement attempts are ignored.