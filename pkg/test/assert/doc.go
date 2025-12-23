/*
Package assert provides test assertion functions for verifying expected
conditions in Go tests. It offers a cleaner alternative to manually writing
if statements with t.Fatal or t.Error calls.

Use this package when writing unit tests that need to verify equality,
check for errors, validate nil values, or confirm that code panics as expected.
Assertions fail the test immediately by default, but can be configured to
continue execution using the Continue option.
*/
package assert
