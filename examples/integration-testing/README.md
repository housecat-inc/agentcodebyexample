# Integration Testing

## Quick Start

```bash
go run ./examples/integration-testing/
time=2026-04-24T07:39:41.165-07:00 level=INFO msg=listening addr=:8080

open http://localhost:8080

go test -v ./examples/integration-testing
```

## Rationale

In traditional "Test Driven Development" (TDD) and "red/green testing", we write idealized but failing tests first (red), then build an implementation that makes the tests pass (green).

A benefit of this approach is that we **apply software design to the testability of our code base**. If up front we design test code that is easy to read, write and fast to run, this cascades down into the code base to make sure our business logic is well factored and easily testable.

For critical paths, well designed tests are worth coding by hand to start, stubbing out and interating on the shape of the initial failing tests.

After that, an agent can easily crank on the implementation to get the tests passing.

## References

- https://en.wikipedia.org/wiki/Test-driven_development
