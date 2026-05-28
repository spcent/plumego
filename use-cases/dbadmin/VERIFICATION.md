# dbadmin Manual Verification Checklist

Use this checklist to verify P0/P1 functionality before a release.
Mark each step ✅ pass / ❌ fail / ⏭ skipped.

---

## Prerequisites

```bash
# Start the server (default: http://127.0.0.1:8080)
cd use-cases/dbadmin
go run . -addr 127.0.0.1:8080 -data /tmp/dbadmin-test -admin-user admin -admin-password admin
```

Open http://127.0.0.1:8080 in a browser and log in with `admin / admin`.

---

## 1. MySQL Connection

| Step | Action | Expected result |
|------|--------|----------------|
| 1.1 | Click "+ Manage Connections" → "+ Add Connection" | Modal opens |
| 1.2 | Fill in a valid MySQL DSN (host, port, database, username). Leave **Save password** unchecked. Save. | Connection appears in list, no password stored (reload page, edit → password field empty) |
| 1.3 | Click **Test** on the connection | Green "✓ Connected" badge |
| 1.4 | Click **Open** | Redirected to SQL Console for that connection |
| 1.5 | Edit the connection, enable **Save password**, enter password | Amber warning shown; save; reload page, edit → password field shows placeholder (stored encrypted) |

---

## 2. SQLite Connection — Server Path

| Step | Action | Expected result |
|------|--------|----------------|
| 2.1 | Add Connection → Driver: SQLite → Mode: Server file path | File path input shown |
| 2.2 | Enter an absolute path to a local `.db` file | Connection saved |
| 2.3 | Test the connection | ✓ Connected |
| 2.4 | Open → navigate to a table | Rows displayed |

## 2b. SQLite Connection — File Upload

| Step | Action | Expected result |
|------|--------|----------------|
| 2b.1 | Add Connection → Driver: SQLite → Mode: Upload file | Dashed upload area shown |
| 2b.2 | Click the upload area, select a `.db` file | File uploaded; filename shown; path filled automatically |
| 2b.3 | Save and Test | ✓ Connected |
| 2b.4 | Open → browse tables | Rows displayed |
| 2b.5 | Connection list shows **Download SQLite** button | Click it → browser downloads the `.db` file |
| 2b.6 | Delete the connection, check "Also delete uploaded file" | Connection and server temp file both removed |

---

## 3. Table Structure

| Step | Action | Expected result |
|------|--------|----------------|
| 3.1 | Navigate to a table → **Fields** tab | Column list with name, type, nullable, default |
| 3.2 | Switch to **Indexes** tab | Index list or "No indexes" |
| 3.3 | Switch to **Foreign Keys** tab | FK list or "No foreign keys" |
| 3.4 | Switch to **DDL** tab | CREATE TABLE statement displayed |
| 3.5 | Click "Copy as: Markdown" | Schema copied; toast "Copied" |
| 3.6 | Click "Copy DDL" on DDL tab | DDL copied; toast "Copied" |

---

## 4. Data Pagination

| Step | Action | Expected result |
|------|--------|----------------|
| 4.1 | Navigate to a table with many rows → **Data** tab | First page shown (default 50 rows) |
| 4.2 | Click "Next →" | Second page loads; row offset correct |
| 4.3 | Click "← Prev" | Returns to first page |
| 4.4 | Verify row count in footer matches total | Count accurate |

---

## 5. Search & Sort

| Step | Action | Expected result |
|------|--------|----------------|
| 5.1 | Click **Filters** → Add filter → column "id", operator "gt", value "10" → Filter | Rows with id > 10 shown |
| 5.2 | Click a column header to sort ascending | Rows sorted; arrow icon on column |
| 5.3 | Click same header again | Sorted descending |
| 5.4 | Clear all filters | All rows return |
| 5.5 | Enter an invalid filter value (SQL injection attempt: `1' OR '1'='1`) | Value treated as literal string, no injection; rows filtered correctly |

---

## 6. Insert / Edit / Delete Row

| Step | Action | Expected result |
|------|--------|----------------|
| 6.1 | Click "+ Insert" → fill required fields → Save | Row inserted; table refreshes |
| 6.2 | Click **Edit** on a row → change a value → Save | Row updated; new value visible |
| 6.3 | Click **Del** on a row → confirm | Row deleted; table refreshes |
| 6.4 | Navigate to a table without a primary key | Insert/Edit/Delete buttons disabled or absent |
| 6.5 | Click a cell with a long value | Cell detail viewer opens |
| 6.6 | Click a cell containing JSON | Viewer shows Pretty/Raw tabs; JSON formatted |

---

## 7. SQL Console

| Step | Action | Expected result |
|------|--------|----------------|
| 7.1 | Navigate to SQL Console, select a database | Database selector shows databases |
| 7.2 | Type `SELECT 1` → Run | Result table shows one row |
| 7.3 | Type `SELECT * FROM users` (large table, no LIMIT) | At most 1000 rows; truncation banner shown |
| 7.4 | Type `DROP TABLE test_tmp` → Run | Danger confirmation dialog shown; cancel → no action |
| 7.5 | Type `DELETE FROM users` (no WHERE) → Run | Danger confirmation dialog shown; "DELETE without WHERE affects all rows" reason |
| 7.6 | Type multiple statements separated by `;` | Error: multiple statements blocked |
| 7.7 | Check History tab after executing SQL | Entry recorded; can click to re-use |

---

## 8. Import SQL

| Step | Action | Expected result |
|------|--------|----------------|
| 8.1 | Tables page → "Import SQL" | Import dialog opens |
| 8.2 | Paste `INSERT INTO test (id) VALUES (1);` → Import | Row inserted |
| 8.3 | Paste SQL containing `DROP TABLE test;` without confirmation | Danger dialog shown; list of dangerous statements visible |
| 8.4 | Confirm dangerous import | Statements executed |
| 8.5 | Import to a read-only connection | HTTP 403 error shown in UI |

---

## 9. Export

| Step | Action | Expected result |
|------|--------|----------------|
| 9.1 | Table → Export → Format: CSV → Download | `.csv` file downloads; open in spreadsheet: columns + data correct |
| 9.2 | Export → Format: SQL → Download | `.sql` file downloads; contains `CREATE TABLE` DDL and `INSERT INTO` statements |
| 9.3 | Export a table with NULL values | CSV: empty cell; SQL: `NULL` keyword |
| 9.4 | Export a table with strings containing single quotes | SQL: `''` escaped correctly |
| 9.5 | (MySQL) Export table with strings containing backslash | SQL: `\\` escaped (e.g. `'a\\nb'`) |

---

## 10. Read-Only Mode

| Step | Action | Expected result |
|------|--------|----------------|
| 10.1 | Edit a connection → enable **Read-only connection** → Save | "READ ONLY" amber badge on connection card |
| 10.2 | Open the connection → Data page | Insert/Edit/Delete buttons are disabled |
| 10.3 | Tables page | "New Table", "Import SQL", and "Drop" buttons are disabled |
| 10.4 | SQL Console | "READ ONLY" badge visible in page header |
| 10.5 | SQL Console: run `SELECT 1` | Works normally |
| 10.6 | SQL Console: run `INSERT INTO t VALUES (1)` | Error response: "this connection is read-only" (HTTP 403) |
| 10.7 | API: `POST /api/conn/:id/db/:db/tables/:table/rows` | Returns `403 READONLY_VIOLATION` |
| 10.8 | API: `DELETE /api/conn/:id/db/:db/tables/:table/rows` | Returns `403 READONLY_VIOLATION` |
| 10.9 | API: `POST /api/conn/:id/db/:db/import` | Returns `403 READONLY_VIOLATION` |

---

## 11. SQLite Download After Modification

| Step | Action | Expected result |
|------|--------|----------------|
| 11.1 | Open an uploaded SQLite connection | Connects successfully |
| 11.2 | Insert a row via the Data page | Row inserted |
| 11.3 | Download the file (Connections page → Download SQLite) | Browser downloads the `.db` file; open with `sqlite3` CLI; row is present |

---

## 12. Theme Switching

| Step | Action | Expected result |
|------|--------|----------------|
| 12.1 | Click the dark/light mode toggle | Theme switches immediately |
| 12.2 | Reload the page | Theme preference persists |

---

## 13. Language Switching

| Step | Action | Expected result |
|------|--------|----------------|
| 13.1 | Click the language toggle (EN/ZH) | All visible labels switch language |
| 13.2 | Reload the page | Language preference persists |
| 13.3 | Navigate to Settings page | "Settings" / "设置" title shown in correct language |

---

## Security Spot-Checks

| Check | Verification |
|-------|-------------|
| Server binds to 127.0.0.1 only | `curl http://0.0.0.0:8080` (from another machine) should fail |
| Passwords not saved by default | Edit connection without Save Password; reload; password field empty |
| Filter injection | Filter value `1' OR '1'='1` returns rows matching literal string, not all rows |
| Sort column injection | `?sortColumn=id;DROP TABLE` — sanitized to `idDROPTABLE` (alphanumeric only); column not found → error |
| SQLite upload: only .db files | Upload a `.txt` file → "not a valid SQLite database" error |
| DDL injection via DEFAULT | `POST /api/conn/:id/db/:db/tables` with `"default": "1); DROP TABLE x; --"` → 422 validation error |

---

## Completion Status

### Fully Implemented (P0/P1)

| Feature | Status |
|---------|--------|
| MySQL connection management (CRUD, test) | ✅ |
| SQLite connection — server path | ✅ |
| SQLite connection — file upload + download | ✅ |
| Table list with engines, row counts | ✅ |
| Table structure (columns, indexes, FK, DDL) | ✅ |
| Data browse with pagination | ✅ |
| Column sort (asc/desc) | ✅ |
| Column filters (eq, ne, gt, gte, lt, lte, like, not_like, is_null, is_not_null) | ✅ |
| Insert / Edit / Delete row | ✅ |
| No-PK guard (edit/delete disabled) | ✅ |
| SQL Console with history | ✅ |
| Dangerous SQL confirmation (DROP, TRUNCATE, ALTER, DELETE/UPDATE without WHERE) | ✅ |
| Multi-statement blocking | ✅ |
| Import SQL with danger confirmation | ✅ |
| Export CSV and SQL dump | ✅ |
| Read-only mode (backend enforced, frontend disabled) | ✅ |
| Save password (encrypted AES-GCM, opt-in) | ✅ |
| Cell detail viewer (JSON pretty-print, BLOB metadata, NULL/empty distinction) | ✅ |
| Copy as Markdown schema / JSON schema / DDL | ✅ |
| Copy rows as JSON / CSV / SQL INSERT | ✅ |
| Dark/light theme | ✅ |
| EN/ZH i18n | ✅ |
| Settings page (SQL history toggle) | ✅ |
| Default listen on 127.0.0.1 | ✅ |
| SQL errors not leaked to client | ✅ |
| SQLite path traversal protection (magic-byte re-verify + random upload names) | ✅ |
| Identifier quoting (MySQL backtick, SQLite double-quote) | ✅ |
| Filter values parameterized | ✅ |
| DDL DEFAULT value injection prevention | ✅ |
| Export backslash escaping for MySQL | ✅ |

### Not Yet Implemented / Known Limitations

| Capability | Notes |
|-----------|-------|
| PostgreSQL support | MySQL and SQLite only |
| Per-connection / per-database access control | All authenticated users can access all connections |
| Query execution timeout | Long-running queries are bounded only by HTTP server timeout (60s) |
| Row-level audit log | No record of who modified which row and when |
| Frontend unit tests | TypeScript strict compilation is the baseline; no Vitest suite |
| `IN` / `NOT IN` / `BETWEEN` filter operators | Not yet in the filter builder |
| BLOB download from Data page | Cell viewer shows metadata; actual binary download not implemented |
| Multi-tenant / team access | Single admin user only |
