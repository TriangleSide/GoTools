## General
- Follow Go best practices.
- Write idiomatic Go code.
- This repository is intended to be used as a library.
- No external dependencies beyond the Go standard library.

## Compatibility
- Code must be compatible with Linux or macOS only.

## Errors
- Wrap errors with contextual information using `fmt.Errorf("...: %w", err)`.
- Use `errors.New("...")` for errors that do not require formatting.
- Use custom error structs when callers need error context for graceful handling.

## Panics
- Use `fmt.Errorf("...")` as the panic argument when formatting is needed.
- Use `errors.New("...")` as the panic argument when formatting is not needed.

## Style
- Write early exit guards to reduce nesting.

## Comments
- Comments should have complete sentences starting with a capital letter and ending with a period.
- Functions, structs, package variables, and package constants must have a comment.
- Comments should only describe purpose and behavior.
- Comments should not describe implementation details.
- Comments should include examples of input and output only when necessary for clarity.
- Ensure existing comments remain accurate after making changes.
- Comments must be written in third person. Avoid second-person pronouns (you, your).

## Testing
- Run the entire test suite with: `make test`
- Test names should follow a format like `Test<FunctionName>_<Scenario>_<ExpectedBehavior>`.
- Write unit tests for every exported function; test both success and error paths.
- Use the assert package for assertions.
- Test files should use the `<package>_test` package name (black-box testing).
- Achieve 100% coverage for all changes made.
- For table-driven tests, define a struct with the subtest name, input, and expected output fields.
- Use `t.Parallel()` in every test and subtest unless there is a specific reason not to.
- Use `t.Helper()` in test helper functions.
- Use `t.Cleanup()` for teardown logic.
- Use `t.Run()` for each scenario in table-driven tests.
- Write concurrency tests when possible.
- Do not add comments in test files.

## Package Docs
- Write package documentation using GoDoc conventions.
- Package comments should be placed in a file named `doc.go` within the package directory.
- Use `/* ... */` style comments for package-level documentation.
- Every package must have a doc comment explaining its purpose and when to use it.
- Focus on what the package does and why a user would choose it, not how it works internally.
- Exclude implementation details, function signatures, variables, and constants from package comments.
- Update package comments when modifying a package to ensure they remain accurate.

## Linting
- Check lints with: `make lint` and fix lints related to modified code.

## Pull Requests
- Before creating a new branch, follow these steps one at a time:
  1. `git stash --all`
  2. `git checkout main`
  3. `git pull origin main`
  4. `git stash pop`
  5. Resolve any merge conflicts.
- Use `git checkout -b <short-description>` to create a new branch.
- Use `git commit -m "<message>"` with a very terse commit message summarizing the change at a high level. 
- Commit messages should end with a period.
- Do not co-author commits; keep the current git user as the sole author.
- Push the branch to remote with `git push -u origin <branch_name>`.
- Use `gh pr create --base main --fill` to open pull requests against the main branch.
- Use `git checkout main` to go to the main branch after the PR is created.
