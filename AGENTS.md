## General
- Follow Go best practices.
- Write idiomatic Go code.
- This repository is intended to be used as a library.
- No external dependencies beyond the Go standard library.

## Compatibility
- Code must be compatible with Linux or macOS only.

## Errors
- Wrap errors with contextual information using `fmt.Errorf("... (%w)", err)`.
- Use `errors.New("...")` for errors that do not require formatting.

## Style
- Write early exit guards to reduce nesting.

## Comments
- Below is a list of areas that must have a terse one or two line comment. The comment should only describe purpose.
  - Functions.
  - Structs.
  - Variables and constants in a package level block.
- Ensure existing comments remain accurate after making changes.

## Testing
- Run the entire test suite with: `make test`
- Test names should follow a format like `Test<FunctionName>_<Scenario>_<ExpectedBehavior>`.
- Write unit tests for every exported function; test both success and error paths.
- Test files should use the `<package>_test` package name (black-box testing).
- Achieve 100% coverage for all changes made.
- For table-driven tests, define a struct with the subtest name, input, and expected output fields.
- Use `t.Parallel()` in every test and subtest unless there is a specific reason not to.
- Use `t.Helper()` in test helper functions.
- Use `t.Cleanup()` for teardown logic.
- Use `t.Run()` for each scenario in table-driven tests.
- Write concurrency tests when possible.
- Do not add comments in test files.

## Linting
- Check lints with: `make lint` and fix lints related to modified code.

## Pull Requests
- Before creating a new branch, follow these steps one at a time:
  1. `git add -A`
  2. `git stash`
  3. `git checkout main`
  4. `git pull`
  5. `git stash pop`
  6. `git add -A`
  7. Resolve any merge conflicts.
- Use `git checkout -b <short-description>` to create a new branch.
- Use `git commit -m "<message>"` with a very terse commit message summarizing the change at a high level. End the commit message with a period.
- Do not co-author commits; keep the current git user as the sole author.
- Push the branch to remote with `git push -u origin <branch_name>`.
- Use `gh pr create --base main --fill` to open pull requests against the main branch.
- Use `git checkout main` to go to the main branch after the PR is created.
