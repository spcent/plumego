// Package db contains tenant-aware database adapters.
//
// TenantDB wraps *sql.DB with explicit tenant query scoping for straightforward
// single-statement SQL:
//   - SELECT and DELETE statements get a tenant predicate inserted or prefixed
//     into the WHERE clause
//   - UPDATE statements inject the tenant predicate ahead of existing WHERE
//     filters while preserving SET-argument ordering
//   - INSERT ... VALUES statements add the tenant column and append one tenant
//     argument per inserted row
//
// Sharp edges stay explicit:
//   - INSERT ... SELECT, CTE-heavy SQL, UNIONs, and other query shapes that
//     cannot be rewritten safely are left unchanged; use RawDB with manual
//     tenant scoping for those paths
//   - QueryFromContext, ExecFromContext, and QueryRowFromContext fail closed
//     when the tenant ID is missing from context
//   - invalid tenant-column configuration is recorded on construction and all
//     query helpers return an error-producing result until corrected
package db
