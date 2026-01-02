/*
Package once provides a mechanism to execute a callback function exactly once
per unique call site during test execution.

Use this package when you need one-time setup that should only execute once
even when called from multiple parallel subtests or test iterations.
*/
package once
