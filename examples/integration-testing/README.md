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

### Browser Testing

One of the most advanced forms of integration testing is testing from the perspective of a user nagivating, clicking and typing in a web page. Structuring our tests and code that make these scenarios easy to describe, execute, and capture screenshots from lets the LLM test and add behaviors to our web app and verify the resulting screenshots visually.

![toggle strikes through](testdata/toggle-strikes-through.png)

## References

- https://en.wikipedia.org/wiki/Test-driven_development
