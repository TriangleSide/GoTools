/*
Package migration provides a framework for orchestrating ordered database migrations.

Use this package when you need to manage database schema or data changes across
application versions. The package handles migration ordering, status tracking,
and distributed locking to ensure only one instance runs migrations at a time.
Migrations are registered with a numeric order and executed sequentially.
*/
package migration
