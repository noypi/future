# FAQ

**Can a promise be reused?** No. A promise settles once.

**What happens if a handler panics?** The panic is recovered and reported to the catch handler.

**Are handlers run on another goroutine?** Only when the promise is executed asynchronously.