## General
- Follow Go best practices.
- Write idiomatic Go code.
- The package is intended to be used as a library.

## Development
- Every function, struct, variable in a package block, and every const at package level must have a terse one or two line comment describing its purpose.
- Ensure existing comments remain accurate after making changes.
- Code must be compatible with Linux or macOS only.
- No external dependencies beyond the Go standard library.
- Wrap errors with contextual information using `fmt.Errorf("... (%w)", err)`.
- Use `errors.New("...")` for errors that do not require formatting.
- Write early exit guards to reduce nesting.

## Testing
- Run tests with: `make test`
- Run linting with: `make lint` and fix lints related to modified code.
- Write unit tests for every exported function.
- Test files should use the <package>_test package name.
- Achieve 100% coverage for all changes.
- Each test file should contain exactly one top-level test, with subtests for each scenario using t.Run(...).
- Subtest names should follow the pattern "when ... it should ..." or "it should ...".
- Use t.Parallel() in every test and subtest unless there is a specific reason not to.
- Table-driven tests should define a test-case struct with fields for name, input, expected output, and any relevant data, and each test case should run in its own subtest via t.Run(...).
- Write concurrency tests where applicable.
- Do not add comments in test files.

## Pull Requests
- Create and use a branch named in the format <short_description>.
- Write a very terse commit message summarizing the change at a high level.
- End the commit message with a period.
- Assume the remote is configured correctly and push the branch with `git push -u origin <branch_name>`.
