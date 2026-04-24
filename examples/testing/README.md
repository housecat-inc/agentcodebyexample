# Testing

## Quick Start

```bash
go test -v examples/testing/main_test.go
```

## Rationale

Writing code and testing code go hand in hand. Software is effectively tested by end-users, so we prefer ways to empirically test it during development.

An agentic coding harness is relentless in its ability to write code, run tests, look at results, and reason about making changes or calling a task complete. Arming it with fast tests that describe system requirements, and strong guidelines about how it can modify the requirements itself is one of the highest leverage things we can do.

Writing tests can be repetitive, so it’s idiomatic to use a table-driven style, where test inputs and expected outputs are listed in a table and a single loop walks over them and performs the test logic.

The resulting test code is easier for a human to review, and less tokens and more guidance for an LLM to write.

## References

- https://gobyexample.com/testing-and-benchmarking
- https://go.dev/wiki/TableDrivenTests
- https://dave.cheney.net/2019/05/07/prefer-table-driven-tests
