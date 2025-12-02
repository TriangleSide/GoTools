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

## Creating a Pull Request
- Before creating a new branch, follow these steps one at a time:
  1. `git add -A`
  2. `git stash`
  3. `git checkout main`
  4. `git pull`
  5. `git stash pop`
  6. `git add -A`
  7. Resolve any merge conflicts.
- Use `git checkout -b <short-description>` to create a new branch.
- Use `git commit -m "<message>"` with a very terse message summarizing the change at a high level, ending with a period.
- Do not co-author commits; keep the current git user as the sole author.
- Push the branch to remote with `git push -u origin <branch_name>`.
- Use `gh pr create --base main` to open pull requests against the main branch.
- Use `git checkout main` to go to the main branch after the PR is created.
