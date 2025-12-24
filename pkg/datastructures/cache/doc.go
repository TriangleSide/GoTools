/*
Package cache provides a thread-safe key/value store with optional expiration.

Use this package when you need a concurrent-safe cache that supports time-to-live
for entries. The cache automatically removes expired entries on access and provides
atomic get-or-set operations to prevent redundant computation for the same key.
*/
package cache
