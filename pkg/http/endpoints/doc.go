/*
Package endpoints provides types and utilities for defining and registering HTTP endpoints.

Use this package when building HTTP APIs that require structured route registration
with path validation. It enables grouping middleware with handlers, validating route
paths at registration time, and organizing endpoints through a builder pattern.

The package supports parameterized paths using curly brace syntax (e.g., "/users/{id}")
and validates that paths follow correct formatting conventions. Types implementing the
Registrar interface can register their endpoints with a shared builder, making it easy
to compose routes from multiple sources.
*/
package endpoints
