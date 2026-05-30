# Manual Verification Guide

This document provides systematic manual verification steps for the Database Admin application. Use this guide to validate functionality across all supported databases after deployment or major changes.

## Prerequisites

- Backend server running (`go run main.go`)
- Frontend dev server running (`cd web && npm run dev`)
- Access to test instances of: SQLite, MySQL, Redis, MongoDB, Elasticsearch
- Valid credentials for each database type

---

## 1. Connection Management

### 1.1 Create Connections

**Test Steps:**
1. Navigate to Connections page
2. Click "New Connection"
3. For each database type, create a connection with:
   - Valid name (e.g., "test-mysql", "test-redis")
   - Correct driver selection
   - Valid host/port/credentials
   - Test connection before saving

**Expected Results:**
- ✓ All 5 connection types save successfully
- ✓ Test connection button validates credentials
- ✓ Invalid connections show clear error messages
- ✓ Connection list displays all created connections

### 1.2 Edit Connections

**Test Steps:**
1. Click edit icon on existing connection
2. Modify connection name
3. Modify host/port (but keep valid)
4. Save changes

**Expected Results:**
- ✓ Changes persist after save
- ✓ Connection still works after edit
- ✓ Modified timestamp updates

### 1.3 Duplicate Connections

**Test Steps:**
1. Click duplicate icon on existing connection
2. Verify new connection appears with "(copy)" suffix
3. Edit duplicate and modify name
4. Test duplicate connection

**Expected Results:**
- ✓ Duplicate created with all settings copied
- ✓ Duplicate is independent (editing one doesn't affect other)
- ✓ Both connections work independently

### 1.4 Delete Connections

**Test Steps:**
1. Click delete icon on connection
2. Confirm deletion in dialog
3. Verify connection removed from list

**Expected Results:**
- ✓ Confirmation dialog appears
- ✓ Cancel aborts deletion
- ✓ Confirm removes connection
- ✓ Associated resources (tables, keys) no longer accessible

### 1.5 Readonly Mode

**Test Steps:**
1. Create connection with readonly=true
2. Navigate to query/data page
3. Attempt write operations (INSERT, UPDATE, DELETE, SET, etc.)

**Expected Results:**
- ✓ Readonly badge shows in header
- ✓ Write operations blocked with clear error
- ✓ Read operations work normally
- ✓ Delete/drop buttons hidden or disabled

---

## 2. SQLite Verification

### 2.1 Database Discovery

**Test Steps:**
1. Create SQLite connection to test database file
2. Open connection
3. Expand resource tree

**Expected Results:**
- ✓ All tables listed in tree
- ✓ System tables (sqlite_*) hidden or marked
- ✓ Table icons correct

### 2.2 Data Browsing

**Test Steps:**
1. Click on table with data
2. Verify data loads in table view
3. Test pagination (next/prev/first/last)
4. Test page size changes (25, 50, 100)

**Expected Results:**
- ✓ Data displays correctly
- ✓ Column headers match schema
- ✓ Pagination works
- ✓ Large text values truncated with "View" button
- ✓ NULL values displayed as "NULL" badge

### 2.3 Filtering

**Test Steps:**
1. Click filter button
2. Add filter: column = value
3. Add multiple filters (AND logic)
4. Clear filters

**Expected Results:**
- ✓ Filter builder opens
- ✓ Column dropdown populated
- ✓ Operator dropdown shows valid operators
- ✓ Results update after applying filter
- ✓ Clear filters resets view

### 2.4 Sorting

**Test Steps:**
1. Click column header to sort ASC
2. Click again for DESC
3. Click third time to clear sort

**Expected Results:**
- ✓ Sort indicator shows (↑/↓)
- ✓ Data reorders correctly
- ✓ Multi-column sort not supported (single click only)

### 2.5 SQL Query Execution

**Test Steps:**
1. Navigate to Query page
2. Execute: `SELECT * FROM table LIMIT 10`
3. Execute: `SELECT COUNT(*) FROM table`
4. Execute invalid SQL

**Expected Results:**
- ✓ Valid queries return results
- ✓ Result count displays
- ✓ Execution time shows
- ✓ Invalid SQL shows error with line number
- ✓ Query history saves successful queries

### 2.6 Dangerous Operations

**Test Steps:**
1. Execute: `DELETE FROM table` (no WHERE)
2. Execute: `DROP TABLE table`
3. Execute: `UPDATE table SET col = val` (no WHERE)

**Expected Results:**
- ✓ Confirmation dialog appears
- ✓ Dialog shows affected row count estimate
- ✓ Cancel aborts operation
- ✓ Confirm executes operation
- ✓ Readonly mode blocks all dangerous ops

### 2.7 Structure View

**Test Steps:**
1. Navigate to Structure page for table
2. View column definitions
3. View indexes
4. View foreign keys

**Expected Results:**
- ✓ Column names, types, nullable, defaults shown
- ✓ Primary key marked
- ✓ Indexes listed with columns
- ✓ Foreign keys show referenced table/column

---

## 3. MySQL Verification

### 3.1 Database Selection

**Test Steps:**
1. Create MySQL connection (no database specified)
2. Open connection
3. Verify database list in tree

**Expected Results:**
- ✓ All accessible databases listed
- ✓ System databases (information_schema, mysql, performance_schema) hidden or marked
- ✓ Click database to expand tables

### 3.2 Data Operations

**Test Steps:**
1. Select table with data
2. Browse data (same as SQLite 2.2)
3. Test filtering (same as SQLite 2.3)
4. Test sorting (same as SQLite 2.4)

**Expected Results:**
- ✓ All SQLite data operations work identically
- ✓ MySQL-specific types (ENUM, SET, JSON) display correctly
- ✓ Large BLOB values show size, not content

### 3.3 SQL Query

**Test Steps:**
1. Execute multi-statement query:
   ```sql
   SELECT * FROM table1;
   SELECT * FROM table2;
   ```
2. Execute query with variables:
   ```sql
   SET @var = 10;
   SELECT * FROM table WHERE id > @var;
   ```

**Expected Results:**
- ✓ Multi-statement queries show multiple result tabs
- ✓ Variables persist within session
- ✓ Session state maintained across queries

### 3.4 Transactions

**Test Steps:**
1. Execute: `START TRANSACTION`
2. Execute: `INSERT INTO table VALUES (...)`
3. Execute: `ROLLBACK`
4. Verify data not inserted

**Expected Results:**
- ✓ Transaction commands work
- ✓ Rollback reverts changes
- ✓ Commit persists changes
- ✓ Connection maintains transaction state

### 3.5 Character Sets

**Test Steps:**
1. Create table with UTF-8 data
2. Insert non-ASCII characters (中文, 日本語, العربية)
3. Query and verify display

**Expected Results:**
- ✓ Unicode characters display correctly
- ✓ No mojibake or encoding errors
- ✓ Emoji display correctly

---

## 4. Redis Verification

### 4.1 Key Discovery

**Test Steps:**
1. Create Redis connection
2. Open connection
3. Verify key list in tree

**Expected Results:**
- ✓ Keys listed (using SCAN, not KEYS *)
- ✓ Key types shown (string, hash, list, set, zset)
- ✓ TTL shown for keys with expiry
- ✓ Search/filter keys by pattern

### 4.2 String Operations

**Test Steps:**
1. Click string key
2. View value
3. Execute: `SET newkey "value"`
4. Execute: `GET newkey`
5. Execute: `SET keywithttl "value" EX 60`

**Expected Results:**
- ✓ String values display correctly
- ✓ Large strings truncated with "View" button
- ✓ SET/GET work
- ✓ TTL displays correctly
- ✓ Binary data shows hex dump

### 4.3 Hash Operations

**Test Steps:**
1. Click hash key
2. View fields and values
3. Execute: `HSET myhash field1 value1`
4. Execute: `HGET myhash field1`
5. Execute: `HGETALL myhash`

**Expected Results:**
- ✓ Hash fields display in table
- ✓ Field values editable
- ✓ HSET/HGET work
- ✓ HGETALL shows all fields

### 4.4 List Operations

**Test Steps:**
1. Click list key
2. View elements
3. Execute: `LPUSH mylist value1`
4. Execute: `RPUSH mylist value2`
5. Execute: `LRANGE mylist 0 -1`
6. Execute: `LPOP mylist`

**Expected Results:**
- ✓ List elements display in order
- ✓ Index shown for each element
- ✓ LPUSH/RPUSH add to correct end
- ✓ LPOP/RPOP remove from correct end
- ✓ LRANGE shows range

### 4.5 Set Operations

**Test Steps:**
1. Click set key
2. View members
3. Execute: `SADD myset member1 member2`
4. Execute: `SMEMBERS myset`
5. Execute: `SISMEMBER myset member1`
6. Execute: `SREM myset member1`

**Expected Results:**
- ✓ Set members display (unordered)
- ✓ Duplicate members ignored
- ✓ SADD adds new members
- ✓ SMEMBERS shows all
- ✓ SISMEMBER returns 1 or 0
- ✓ SREM removes member

### 4.6 Sorted Set Operations

**Test Steps:**
1. Click zset key
2. View members with scores
3. Execute: `ZADD myzset 1 member1 2 member2`
4. Execute: `ZRANGE myzset 0 -1 WITHSCORES`
5. Execute: `ZRANGEBYSCORE myzset 1 2`
6. Execute: `ZREM myzset member1`

**Expected Results:**
- ✓ Members display sorted by score
- ✓ Scores shown
- ✓ ZADD adds/updates members
- ✓ ZRANGE shows range with scores
- ✓ ZRANGEBYSCORE filters by score
- ✓ ZREM removes member

### 4.7 Command History

**Test Steps:**
1. Execute several commands
2. Open command history panel
3. Click history entry
4. Clear history

**Expected Results:**
- ✓ History shows executed commands
- ✓ Timestamp and execution time shown
- ✓ Click populates command input
- ✓ Clear removes all history
- ✓ History persists in localStorage

### 4.8 Dangerous Operations

**Test Steps:**
1. Execute: `DEL importantkey`
2. Execute: `FLUSHDB`
3. Execute: `FLUSHALL`

**Expected Results:**
- ✓ Confirmation dialog appears for DEL
- ✓ Strong warning for FLUSHDB/FLUSHALL
- ✓ Readonly mode blocks all write operations
- ✓ System keys (if any) protected

---

## 5. MongoDB Verification

### 5.1 Database & Collection Discovery

**Test Steps:**
1. Create MongoDB connection
2. Open connection
3. Verify database and collection list

**Expected Results:**
- ✓ All accessible databases listed
- ✓ Collections listed under each database
- ✓ System collections (system.*) hidden or marked
- ✓ Document count shown for each collection

### 5.2 Document Browsing

**Test Steps:**
1. Click collection
2. View documents
3. Test pagination
4. Test page size changes

**Expected Results:**
- ✓ Documents display in table or JSON view
- ✓ _id field shown first
- ✓ Nested objects expandable
- ✓ Arrays shown with length
- ✓ Pagination works

### 5.3 Query Filtering

**Test Steps:**
1. Open filter builder
2. Simple filter: `{field: "value"}`
3. Complex filter: `{age: {$gt: 18}, status: "active"}`
4. Regex filter: `{name: /^John/}`
5. Array filter: `{tags: {$in: ["tag1", "tag2"]}}`

**Expected Results:**
- ✓ Filter syntax validated
- ✓ Results update correctly
- ✓ Complex queries work
- ✓ Regex patterns supported
- ✓ Array operators work

### 5.4 Projection

**Test Steps:**
1. Add projection: `{name: 1, age: 1, _id: 0}`
2. Execute query
3. Verify only projected fields shown

**Expected Results:**
- ✓ Projection syntax validated
- ✓ Only specified fields returned
- ✓ _id exclusion works

### 5.5 Sorting

**Test Steps:**
1. Add sort: `{age: 1}` (ascending)
2. Add sort: `{age: -1}` (descending)
3. Multi-field sort: `{status: 1, age: -1}`

**Expected Results:**
- ✓ Sort syntax validated
- ✓ Results ordered correctly
- ✓ Multi-field sort works

### 5.6 Aggregation Pipeline

**Test Steps:**
1. Execute aggregation:
   ```javascript
   [
     {$match: {status: "active"}},
     {$group: {_id: "$category", count: {$sum: 1}}},
     {$sort: {count: -1}}
   ]
   ```
2. Execute with $lookup:
   ```javascript
   [
     {$lookup: {from: "orders", localField: "_id", foreignField: "userId", as: "orders"}}
   ]
   ```

**Expected Results:**
- ✓ Pipeline stages execute correctly
- ✓ Results display
- ✓ Complex aggregations work
- ✓ $lookup joins collections

### 5.7 Document Operations

**Test Steps:**
1. Insert document: `db.collection.insertOne({name: "test", value: 123})`
2. Update document: `db.collection.updateOne({_id: ...}, {$set: {value: 456}})`
3. Delete document: `db.collection.deleteOne({_id: ...})`

**Expected Results:**
- ✓ Insert adds document
- ✓ Update modifies document
- ✓ Delete removes document
- ✓ _id auto-generated if not provided
- ✓ ObjectId displayed correctly

### 5.8 ObjectId Handling

**Test Steps:**
1. Query by ObjectId: `{_id: ObjectId("...")}`
2. Verify ObjectId clickable to copy
3. Verify ObjectId in results

**Expected Results:**
- ✓ ObjectId syntax recognized
- ✓ ObjectId converted correctly
- ✓ Copy button works
- ✓ ObjectId displays as hex string

### 5.9 Dangerous Operations

**Test Steps:**
1. Execute: `db.collection.drop()`
2. Execute: `db.collection.deleteMany({})`
3. Execute: `db.collection.updateMany({}, {$set: {field: "value"}})`

**Expected Results:**
- ✓ Confirmation dialog for drop
- ✓ Strong warning for deleteMany with empty filter
- ✓ Strong warning for updateMany with empty filter
- ✓ Readonly mode blocks all write operations

---

## 6. Elasticsearch Verification

### 6.1 Index Discovery

**Test Steps:**
1. Create Elasticsearch connection
2. Open connection
3. Verify index list

**Expected Results:**
- ✓ All accessible indices listed
- ✓ System indices (starting with .) hidden or marked
- ✓ Document count shown
- ✓ Index size shown
- ✓ Health status shown (green/yellow/red)

### 6.2 Document Search

**Test Steps:**
1. Click index
2. Execute match_all query:
   ```json
   {
     "query": {
       "match_all": {}
     },
     "size": 10
   }
   ```
3. Execute match query:
   ```json
   {
     "query": {
       "match": {
         "field": "value"
       }
     }
   }
   ```
4. Execute bool query:
   ```json
   {
     "query": {
       "bool": {
         "must": [{"match": {"status": "active"}}],
         "must_not": [{"match": {"deleted": true}}],
         "filter": [{"range": {"age": {"gte": 18}}}]
       }
     }
   }
   ```

**Expected Results:**
- ✓ Query syntax validated
- ✓ Results display with _score
- ✓ Highlighting works (if configured)
- ✓ Total hit count shown
- ✓ Execution time shown

### 6.3 Aggregations

**Test Steps:**
1. Execute terms aggregation:
   ```json
   {
     "size": 0,
     "aggs": {
       "categories": {
         "terms": {"field": "category.keyword"}
       }
     }
   }
   ```
2. Execute date histogram:
   ```json
   {
     "size": 0,
     "aggs": {
       "over_time": {
         "date_histogram": {
           "field": "timestamp",
           "calendar_interval": "day"
         }
       }
     }
   }
   ```

**Expected Results:**
- ✓ Aggregation results display
- ✓ Bucket counts shown
- ✓ Nested aggregations work
- ✓ Metrics aggregations (avg, sum, etc.) work

### 6.4 Document Operations

**Test Steps:**
1. Get document by ID: `GET /index/_doc/id`
2. Index document: `PUT /index/_doc/id` with body
3. Delete document: `DELETE /index/_doc/id`

**Expected Results:**
- ✓ Get returns full document
- ✓ Index creates/updates document
- ✓ Delete removes document
- ✓ Version conflict handled

### 6.5 Mapping View

**Test Steps:**
1. Navigate to Structure page
2. View index mapping
3. Verify field types

**Expected Results:**
- ✓ Mapping displays in tree or JSON view
- ✓ Field types shown (text, keyword, integer, etc.)
- ✓ Nested objects expandable
- ✓ Multi-fields shown

### 6.6 Query DSL History

**Test Steps:**
1. Execute several queries
2. Open query history panel
3. Click history entry
4. Clear history

**Expected Results:**
- ✓ History shows executed queries
- ✓ Timestamp and execution time shown
- ✓ Result count shown
- ✓ Click populates query editor
- ✓ Clear removes all history
- ✓ History persists in localStorage

### 6.7 Dangerous Operations

**Test Steps:**
1. Execute: `DELETE /index/_doc/id` (delete document)
2. Execute: `DELETE /index` (delete index)
3. Execute: `_delete_by_query` with broad filter

**Expected Results:**
- ✓ Confirmation dialog for delete document
- ✓ Strong warning for delete index
- ✓ Strong warning for delete_by_query
- ✓ Readonly mode blocks all write operations

---

## 7. Cross-Database Features

### 7.1 Data Export

**Test Steps:**
1. Select table/collection/index
2. Export as CSV
3. Export as JSON
4. Verify exported files

**Expected Results:**
- ✓ CSV export includes headers
- ✓ CSV handles commas in values
- ✓ JSON export is valid
- ✓ Large exports don't timeout
- ✓ Binary data handled correctly

### 7.2 Query History

**Test Steps:**
1. Execute queries across all database types
2. Open query history panel
3. Filter by database type
4. Search history
5. Delete individual entries
6. Clear all history

**Expected Results:**
- ✓ History saved per connection
- ✓ Filter works
- ✓ Search finds queries
- ✓ Delete removes entry
- ✓ Clear removes all
- ✓ History persists across sessions

### 7.3 Theme Toggle

**Test Steps:**
1. Click theme toggle button
2. Verify all pages in light mode
3. Switch to dark mode
4. Verify all pages in dark mode
5. Refresh page

**Expected Results:**
- ✓ All pages render correctly in both themes
- ✓ Text readable (sufficient contrast)
- ✓ Theme persists across refresh
- ✓ No layout shifts

### 7.4 Language Switch

**Test Steps:**
1. Click language switch (if available)
2. Verify UI labels change
3. Verify error messages change
4. Verify tooltips change
5. Refresh page

**Expected Results:**
- ✓ All UI labels translated
- ✓ No untranslated strings
- ✓ Language persists across refresh
- ✓ RTL languages (if supported) display correctly

### 7.5 Keyboard Shortcuts

**Test Steps:**
1. Test Ctrl/Cmd+Enter (execute query)
2. Test Ctrl/Cmd+S (save)
3. Test Escape (close dialogs)
4. Test Tab (navigate form fields)

**Expected Results:**
- ✓ Shortcuts work
- ✓ No conflicts
- ✓ Focus indicators visible
- ✓ Tab order logical

---

## 8. Security Verification

### 8.1 Authentication

**Test Steps:**
1. Access application without login
2. Login with valid credentials
3. Login with invalid credentials
4. Logout
5. Try accessing protected resource after logout

**Expected Results:**
- ✓ Redirected to login page
- ✓ Valid credentials grant access
- ✓ Invalid credentials show error
- ✓ Logout clears session
- ✓ Protected resources redirect to login

### 8.2 Session Management

**Test Steps:**
1. Login
2. Wait for session timeout (if configured)
3. Try executing query
4. Refresh page

**Expected Results:**
- ✓ Expired session redirects to login
- ✓ Active session persists across refresh
- ✓ Session token not exposed in URL
- ✓ Session token has expiry

### 8.3 SQL Injection Prevention

**Test Steps:**
1. Try SQL injection in filter: `' OR '1'='1`
2. Try SQL injection in query parameters
3. Verify parameterized queries used

**Expected Results:**
- ✓ Injection attempts fail
- ✓ No data leakage
- ✓ Error messages don't reveal schema
- ✓ Parameterized queries prevent injection

### 8.4 XSS Prevention

**Test Steps:**
1. Insert data with script tag: `<script>alert('xss')</script>`
2. View data in table
3. View data in detail view

**Expected Results:**
- ✓ Script tags escaped
- ✓ No script execution
- ✓ HTML entities encoded
- ✓ Content Security Policy (if configured) blocks inline scripts

### 8.5 CSRF Protection

**Test Steps:**
1. Login
2. Open another tab
3. Try executing action from another origin
4. Verify CSRF token present

**Expected Results:**
- ✓ CSRF token in forms
- ✓ Cross-origin requests blocked
- ✓ Token validated on server
- ✓ Missing token returns 403

### 8.6 Credential Redaction

**Test Steps:**
1. Create connection with password
2. View connection in API response
3. Check browser network tab
4. Check server logs

**Expected Results:**
- ✓ Passwords redacted in API responses
- ✓ Passwords not in network tab
- ✓ Passwords not in server logs
- ✓ Only connection name/host visible

---

## 9. Performance Verification

### 9.1 Large Dataset Handling

**Test Steps:**
1. Table with 1M+ rows
2. Browse with pagination
3. Apply filter
4. Sort by indexed column
5. Sort by non-indexed column

**Expected Results:**
- ✓ Pagination loads quickly (<2s)
- ✓ Filter with index fast
- ✓ Filter without index shows warning
- ✓ Sort with index fast
- ✓ Sort without index slow but works

### 9.2 Query Execution Time

**Test Steps:**
1. Execute fast query (<100ms)
2. Execute slow query (>5s)
3. Execute query that times out

**Expected Results:**
- ✓ Fast queries return immediately
- ✓ Slow queries show progress
- ✓ Timeout shows clear error
- ✓ Execution time displayed

### 9.3 Memory Usage

**Test Steps:**
1. Open multiple connections
2. Browse large tables
3. Execute queries returning large result sets
4. Monitor browser memory

**Expected Results:**
- ✓ Memory usage reasonable (<500MB)
- ✓ No memory leaks
- ✓ Large results paginated
- ✓ Old results garbage collected

---

## 10. Error Handling

### 10.1 Network Errors

**Test Steps:**
1. Disconnect network
2. Try executing query
3. Reconnect network
4. Try again

**Expected Results:**
- ✓ Clear network error message
- ✓ Retry button shown
- ✓ Reconnect works
- ✓ No data loss

### 10.2 Database Errors

**Test Steps:**
1. Execute query with syntax error
2. Try connecting to invalid host
3. Try accessing non-existent table
4. Try operation without permission

**Expected Results:**
- ✓ Syntax errors show line number
- ✓ Connection errors show reason
- ✓ Not found errors clear
- ✓ Permission errors clear
- ✓ No stack traces in UI

### 10.3 Concurrency

**Test Steps:**
1. Open same table in two tabs
2. Update in one tab
3. Refresh other tab
4. Try updating same row

**Expected Results:**
- ✓ Last write wins (or optimistic locking)
- ✓ No data corruption
- ✓ Clear conflict message (if any)

---

## 11. Accessibility

### 11.1 Keyboard Navigation

**Test Steps:**
1. Tab through all pages
2. Use arrow keys in dropdowns
3. Use Enter to select
4. Use Escape to close

**Expected Results:**
- ✓ All interactive elements focusable
- ✓ Focus indicator visible
- ✓ Dropdowns work with keyboard
- ✓ Modals trap focus
- ✓ Escape closes modals

### 11.2 Screen Reader

**Test Steps:**
1. Use screen reader (VoiceOver, NVDA)
2. Navigate pages
3. Execute queries
4. View results

**Expected Results:**
- ✓ Page structure announced
- ✓ Buttons have labels
- ✓ Tables have headers
- ✓ Errors announced
- ✓ Dynamic content announced

### 11.3 Color Contrast

**Test Steps:**
1. Check all text in light mode
2. Check all text in dark mode
3. Use contrast checker tool

**Expected Results:**
- ✓ All text meets WCAG AA (4.5:1)
- ✓ Large text meets 3:1
- ✓ Icons have sufficient contrast
- ✓ Error states distinguishable

---

## 12. Mobile & Responsive

### 12.1 Small Screens

**Test Steps:**
1. Resize browser to mobile width (375px)
2. Navigate all pages
3. Try executing queries
4. Try filtering

**Expected Results:**
- ✓ Layout adapts
- ✓ Tables horizontal scroll
- ✓ Buttons touch-friendly (44px+)
- ✓ Text readable
- ✓ No horizontal overflow

### 12.2 Touch Interactions

**Test Steps:**
1. Use on mobile device or touch emulator
2. Tap buttons
3. Scroll tables
4. Swipe (if supported)

**Expected Results:**
- ✓ Tap targets large enough
- ✓ Scroll smooth
- ✓ No accidental taps
- ✓ Gestures work (if implemented)

---

## 13. Browser Compatibility

### 13.1 Chrome/Edge

**Test Steps:**
1. Test all features in Chrome
2. Test all features in Edge

**Expected Results:**
- ✓ All features work
- ✓ No console errors
- ✓ Performance acceptable

### 13.2 Firefox

**Test Steps:**
1. Test all features in Firefox

**Expected Results:**
- ✓ All features work
- ✓ No console errors
- ✓ Performance acceptable

### 13.3 Safari

**Test Steps:**
1. Test all features in Safari (macOS/iOS)

**Expected Results:**
- ✓ All features work
- ✓ No console errors
- ✓ Performance acceptable
- ✓ Safari-specific quirks handled

---

## 14. Data Integrity

### 14.1 Transaction Safety

**Test Steps:**
1. Start transaction
2. Execute multiple operations
3. Rollback
4. Verify no changes

**Expected Results:**
- ✓ Rollback reverts all changes
- ✓ No partial updates
- ✓ Transaction isolation maintained

### 14.2 Concurrent Access

**Test Steps:**
1. Two users edit same record
2. First saves
3. Second saves
4. Verify result

**Expected Results:**
- ✓ Last write wins (or conflict detected)
- ✓ No data corruption
- ✓ Clear behavior

### 14.3 Backup & Restore

**Test Steps:**
1. Export database
2. Modify data
3. Import backup
4. Verify restoration

**Expected Results:**
- ✓ Export complete
- ✓ Import restores data
- ✓ No data loss

---

## 15. Logging & Monitoring

### 15.1 Audit Logs

**Test Steps:**
1. Execute various operations
2. Check audit log (if available)
3. Verify logged actions

**Expected Results:**
- ✓ Login/logout logged
- ✓ Queries logged
- ✓ Dangerous operations logged
- ✓ Timestamps accurate
- ✓ User identified

### 15.2 Error Logs

**Test Steps:**
1. Trigger various errors
2. Check server logs
3. Verify error details

**Expected Results:**
- ✓ Errors logged with stack trace
- ✓ User context included
- ✓ No sensitive data in logs
- ✓ Log level appropriate

### 15.3 Performance Metrics

**Test Steps:**
1. Execute queries
2. Check metrics endpoint (if available)
3. Verify metrics

**Expected Results:**
- ✓ Query count tracked
- ✓ Execution time tracked
- ✓ Error rate tracked
- ✓ Metrics accurate

---

## Verification Checklist

Use this checklist to track overall verification progress:

### Connection Management
- [ ] Create connections (all 5 types)
- [ ] Edit connections
- [ ] Duplicate connections
- [ ] Delete connections
- [ ] Readonly mode works

### SQLite
- [ ] Database discovery
- [ ] Data browsing
- [ ] Filtering
- [ ] Sorting
- [ ] SQL query execution
- [ ] Dangerous operations blocked
- [ ] Structure view

### MySQL
- [ ] Database selection
- [ ] Data operations
- [ ] SQL query (multi-statement)
- [ ] Transactions
- [ ] Character sets

### Redis
- [ ] Key discovery
- [ ] String operations
- [ ] Hash operations
- [ ] List operations
- [ ] Set operations
- [ ] Sorted set operations
- [ ] Command history
- [ ] Dangerous operations blocked

### MongoDB
- [ ] Database & collection discovery
- [ ] Document browsing
- [ ] Query filtering
- [ ] Projection
- [ ] Sorting
- [ ] Aggregation pipeline
- [ ] Document operations
- [ ] ObjectId handling
- [ ] Dangerous operations blocked

### Elasticsearch
- [ ] Index discovery
- [ ] Document search
- [ ] Aggregations
- [ ] Document operations
- [ ] Mapping view
- [ ] Query DSL history
- [ ] Dangerous operations blocked

### Cross-Database Features
- [ ] Data export (CSV/JSON)
- [ ] Query history
- [ ] Theme toggle
- [ ] Language switch
- [ ] Keyboard shortcuts

### Security
- [ ] Authentication
- [ ] Session management
- [ ] SQL injection prevention
- [ ] XSS prevention
- [ ] CSRF protection
- [ ] Credential redaction

### Performance
- [ ] Large dataset handling
- [ ] Query execution time
- [ ] Memory usage

### Error Handling
- [ ] Network errors
- [ ] Database errors
- [ ] Concurrency

### Accessibility
- [ ] Keyboard navigation
- [ ] Screen reader
- [ ] Color contrast

### Mobile & Responsive
- [ ] Small screens
- [ ] Touch interactions

### Browser Compatibility
- [ ] Chrome/Edge
- [ ] Firefox
- [ ] Safari

### Data Integrity
- [ ] Transaction safety
- [ ] Concurrent access
- [ ] Backup & restore

### Logging & Monitoring
- [ ] Audit logs
- [ ] Error logs
- [ ] Performance metrics

---

## Reporting Issues

When reporting issues found during manual verification:

1. **Database type**: Which database (SQLite, MySQL, Redis, MongoDB, Elasticsearch)
2. **Operation**: What you were trying to do
3. **Steps**: Exact steps to reproduce
4. **Expected**: What should happen
5. **Actual**: What actually happened
6. **Error message**: Any error messages shown
7. **Browser**: Browser name and version
8. **Screenshots**: If applicable
9. **Console logs**: Any errors in browser console
10. **Server logs**: Any errors in server logs

---

## Updates

This document should be updated when:
- New database types are added
- New features are implemented
- UI changes significantly
- Security policies change
- New browsers need support

Last updated: 2026-05-30
