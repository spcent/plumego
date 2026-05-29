import { createContext, useCallback, useContext, useState, type ReactNode } from 'react'

export type Locale = 'en' | 'zh'

type Dict = Record<string, string>

const en: Dict = {
  // Login
  'login.username': 'Username',
  'login.password': 'Password',
  'login.submit': 'Sign in',
  'login.submitting': 'Signing in…',

  // Nav
  'nav.manage_connections': '+ Manage Connections',
  'nav.sql_console': 'SQL Console',
  'nav.logout': 'Logout',

  // Connections
  'connections.title': 'Connections',
  'connections.add': '+ Add Connection',
  'connections.empty': 'No connections yet. Add your first database connection.',
  'connections.test': 'Test',
  'connections.edit': 'Edit',
  'connections.delete': 'Delete',
  'connections.add_title': 'Add Connection',
  'connections.edit_title': 'Edit Connection',
  'connections.form.name': 'Name',
  'connections.form.driver': 'Driver',
  'connections.form.file_path': 'File Path',
  'connections.form.host': 'Host',
  'connections.form.port': 'Port',
  'connections.form.database': 'Database',
  'connections.form.username': 'Username',
  'connections.form.password': 'Password',
  'connections.form.save': 'Save',
  'connections.form.cancel': 'Cancel',
  'connections.open': 'Open',
  'connections.delete.title': 'Delete Connection',
  'connections.delete.message': 'Delete "{name}"? This cannot be undone.',
  'connections.form.readonly': 'Read-only connection',
  'connections.form.readonly_hint': 'Disables all write operations',
  'connections.form.save_password': 'Save password',
  'connections.form.save_password_warn': 'Password will be encrypted and stored in the server data directory. Do not use on shared machines without access control.',

  // Tables
  'tables.sql_console': '▶ SQL Console',
  'tables.new_table': '+ New Table',
  'tables.col.name': 'Name',
  'tables.col.type': 'Type',
  'tables.col.engine': 'Engine',
  'tables.col.rows': 'Rows',
  'tables.col.comment': 'Comment',
  'tables.action.data': 'Data',
  'tables.action.fields': 'Fields',
  'tables.action.indexes': 'Indexes',
  'tables.action.drop': 'Drop',
  'tables.empty': 'No tables found',
  'tables.drop.title': 'Drop Table',
  'tables.drop.message': 'Drop table "{table}"? This cannot be undone.',
  'tables.create.title': 'Create Table',
  'tables.create.hint': 'Use the SQL Console for full DDL control.',
  'tables.create.open_console': 'Open SQL Console',

  // Data
  'data.filter.placeholder': 'col:value filter',
  'data.filter.apply': 'Filter',
  'data.refresh': 'Refresh',
  'data.filter.show': 'Filters',
  'data.filter.clear_all': 'Clear all',
  'data.filter.add': 'Add',
  'data.filter.col_placeholder': 'Column…',
  'data.filter.val_placeholder': 'Value',
  'data.insert': '+ Insert',
  'data.prev': '← Prev',
  'data.next': 'Next →',
  'data.edit': 'Edit',
  'data.delete': 'Del',
  'data.delete.title': 'Delete Row',
  'data.delete.message': 'Delete row where {col} = {val}?',
  'data.insert.title': 'Insert Row',
  'data.edit.title': 'Edit Row',
  'data.save': 'Save',
  'data.cancel': 'Cancel',
  'data.no_rows': 'No rows',
  'data.loading': 'loading…',
  'data.no_pk_hint': 'No primary key — edit/delete unavailable',
  'data.load_failed': 'Load failed',
  'data.copy_failed': 'Copy failed',
  'data.delete_failed': 'Delete failed',
  'data.save_failed': 'Save failed',
  'data.insert.success': 'Row inserted',
  'data.edit.success': 'Row updated',
  'data.pagination': 'Page {page} of {total}',
  'data.page_size': '{n} / page',

  // Structure
  'structure.tab.columns': 'Columns',
  'structure.tab.indexes': 'Indexes',
  'structure.tab.fks': 'Foreign Keys',
  'structure.tab.ddl': 'DDL',
  'structure.no_indexes': 'No indexes',
  'structure.no_fks': 'No foreign keys',

  // Query
  'query.title': 'SQL Console',
  'query.db_placeholder': '-- database --',
  'query.run': '▶ Run (Cmd+Enter)',
  'query.running': 'Running…',
  'query.copy': 'Copy SQL',
  'query.clear': 'Clear',
  'query.confirm_danger': 'Execute dangerous SQL?',
  'query.danger_reason': 'Reason',
  'query.confirm_proceed': 'Proceed',
  'query.rows_affected': 'row(s) affected',
  'query.last_insert_id': 'Last insert ID',
  'query.truncated': 'Result truncated to 1000 rows — add LIMIT to your query',
  'query.tab.result': 'Result',
  'query.tab.history': 'History',
  'query.history.empty': 'No history yet',

  // Readonly
  'readonly.badge': 'READ ONLY',
  'readonly.violation': 'This connection is read-only',

  // Settings
  'settings.title': 'Settings',
  'settings.sql_history': 'SQL History',
  'settings.sql_history_enabled': 'Enable SQL history',
  'settings.sql_history_hint': 'Executed SQL is recorded server-side (max 100 entries per connection, 30-day retention). Manage history in the SQL Console history tab.',
  'nav.settings': 'Settings',

  // History (localStorage)
  'history.tab': 'History',
  'history.empty': 'No history yet',
  'history.clear_all': 'Clear all',
  'history.disabled': 'SQL history is disabled. Enable it in Settings.',
  'history.delete': 'Delete',

  // Export
  'export.title': 'Export',
  'export.format': 'Format',
  'export.include_schema': 'Include schema (DDL)',
  'export.include_data': 'Include data (INSERTs)',
  'export.download': 'Download',

  // Import
  'import.title': 'Import SQL',
  'import.paste': 'Paste SQL or upload a .sql file',
  'import.file': 'Upload file',
  'import.run': 'Import',
  'import.running': 'Importing…',
  'import.confirm_danger': 'Import contains dangerous SQL',
  'import.confirm_proceed': 'Proceed',

  // Tables toolbar
  'tables.import': 'Import SQL',
  'tables.copy_schema': 'Copy Schema',
  'tables.copy_schema.success': 'Schema copied',
  'tables.copy_schema.loading': 'Copying…',

  // Connections filter
  'connections.filter.all': 'All',
  'connections.filter.mysql': 'MySQL',
  'connections.filter.sqlite': 'SQLite',
  'connections.filter.redis': 'Redis',
  'connections.duplicate': 'Duplicate',

  // Redis
  'redis.form.password': 'AUTH Password',
  'redis.form.password_placeholder': 'Optional — leave blank if no AUTH',
  'redis.form.db_index': 'DB Index',
  'redis.form.db_index_hint': '0–15, default 0',
  'redis.form.tls': 'Use TLS / SSL',
  'redis.readonly': 'Read-only',
  'redis.driver_coming_soon': 'Redis driver is coming in the next release. Connection config is saved and ready.',
  'redis.panel.title': 'Redis Key Browser',
  'redis.panel.hint': 'Key browsing, TTL inspection, and command execution will be available once the Redis driver is released.',

  // Cell viewer
  'cell.tab.raw': 'Raw',
  'cell.tab.pretty': 'Pretty',
  'cell.copy': 'Copy',
  'cell.copy_pretty': 'Copy Pretty',
  'cell.copy_minified': 'Copy Minified',
  'cell.empty': 'empty string',
  'cell.blob': 'BLOB',
  'cell.blob_note': 'Binary data — download available in a future release.',

  // Copy
  'copy.as': 'Copy as:',
  'copy.markdown': 'Markdown',
  'copy.json_schema': 'JSON Schema',
  'copy.ddl': 'Copy DDL',
  'copy.json': 'JSON',
  'copy.csv': 'CSV',
  'copy.insert': 'SQL INSERT',
  'copy.rows_selected': 'Copy {n} row(s):',
  'copy.cell_success': 'Copied',

  // SQLite upload/download
  'sqlite.mode_path': 'Server file path',
  'sqlite.mode_upload': 'Upload file',
  'sqlite.upload_choose': 'Choose .db / .sqlite file…',
  'sqlite.uploading': 'Uploading…',
  'sqlite.upload_failed': 'Upload failed',
  'sqlite.replace': 'Replace',
  'sqlite.download': 'Download SQLite',
  'sqlite.delete_file': 'Also delete the uploaded SQLite file from server',
  'sqlite.upload_hint': 'File must be a valid SQLite database. Max 512 MiB.',

  // Confirm
  'confirm.cancel': 'Cancel',
  'confirm.delete': 'Delete',
  'confirm.drop': 'Drop',

  // Redis Key Browser
  'redis.browser.pattern': 'Pattern',
  'redis.browser.pattern_placeholder': 'key:* filter (default *)',
  'redis.browser.search': 'Search',
  'redis.browser.load_more': 'Load more',
  'redis.browser.empty': 'No keys found',
  'redis.browser.loading': 'Loading…',
  'redis.browser.key_count': '{n} keys',
  'redis.browser.db': 'DB {n}',
  'redis.browser.ttl_none': 'no expiry',
  'redis.browser.ttl_sec': '{n}s',

  // Redis Key Inspector
  'redis.key.ttl': 'TTL',
  'redis.key.type': 'Type',
  'redis.key.encoding': 'Encoding',
  'redis.key.set_ttl': 'Set TTL',
  'redis.key.ttl_hint': 'seconds; -1 to remove expiry',
  'redis.key.delete': 'Delete Key',
  'redis.key.delete_confirm': 'Delete key "{key}"? This cannot be undone.',
  'redis.key.copy': 'Copy',
  'redis.key.loading': 'Loading…',
  'redis.key.not_found': 'Key not found',
  'redis.key.stream_length': 'Length: {n}',
  'redis.key.stream_groups': 'Groups: {n}',
  'redis.key.list_length': '{n} items',
  'redis.key.set_size': '{n} members',
  'redis.key.hash_search': 'Search fields...',
  'redis.key.hash_no_match': 'No fields match',
  'redis.key.page': 'Page {current} of {total}',
  'redis.key.prev': 'Previous',
  'redis.key.next': 'Next',
  'redis.key.hash_fields': '{n} fields',

  // Redis Command Console
  'redis.command.title': 'Command Console',
  'redis.command.placeholder': 'e.g. GET mykey',
  'redis.command.run': 'Run',
  'redis.command.running': 'Running…',
  'redis.command.readonly_blocked': 'Blocked: write command on read-only connection',
  'redis.command.forbidden': 'Forbidden command',

  // Redis Batch Operations
  'redis.batch.select_keys': 'Select keys to delete',
  'redis.batch.cancel': 'Cancel selection',
  'redis.batch.selected': '{n} keys selected',
  'redis.batch.preview': 'Preview',
  'redis.batch.delete': 'Delete Selected',
  'redis.batch.confirm_delete': 'Delete {n} selected keys? This cannot be undone.',
  'redis.batch.deleted': '{n} keys deleted',
  'redis.batch.no_keys': 'No keys match the pattern',
  'redis.batch.warning': 'Warning: This will delete {n} keys permanently',
  'redis.batch.select_all': 'Select all',
  'redis.batch.deselect_all': 'Deselect all',

  // Redis Pattern Favorites
  'redis.favorites.title': 'Pattern Favorites',
  'redis.favorites.save': 'Save Pattern',
  'redis.favorites.saved': 'Pattern saved',
  'redis.favorites.no_favorites': 'No saved patterns',
  'redis.favorites.delete': 'Remove from favorites',
  'redis.favorites.name': 'Pattern name',

  // Redis Big Key Warning
  'redis.big_key.warning': 'Large key detected ({size})',
  'redis.big_key.confirm_load': 'This key is {size}. Loading it may slow down the interface. Continue?',
  'redis.big_key.skip': 'Skip loading large value',

  // Redis Memory
  'redis.memory': 'Memory: {size}',
  'redis.memory.bytes': '{n} B',
  'redis.memory.kb': '{n} KB',
  'redis.memory.mb': '{n} MB',
  'redis.memory.gb': '{n} GB',

  // Redis Hash Field Search
  'redis.hash.search': 'Search fields',
  'redis.hash.search_placeholder': 'Filter by field name…',
  'redis.hash.no_match': 'No fields match the filter',

  // Redis List/Set/ZSet Pagination
  'redis.collection.range': 'Showing {start}-{end} of {total}',
  'redis.collection.load_more': 'Load more items',
  'redis.collection.first': 'First',
  'redis.collection.prev': 'Previous',
  'redis.collection.next': 'Next',

  // Redis Stream
  'redis.stream.id': 'ID',
  'redis.stream.fields': 'Fields',
  'redis.stream.no_messages': 'No messages in stream',

  // Redis Copy
  'redis.copy.key': 'Copy Key',
  'redis.copy.value': 'Copy Value',
  'redis.copy.key_success': 'Key copied to clipboard',
  'redis.copy.value_success': 'Value copied to clipboard',

  // Redis TTL Presets
  'redis.ttl.presets': 'TTL Presets',
  'redis.ttl.1min': '1 minute',
  'redis.ttl.5min': '5 minutes',
  'redis.ttl.1hour': '1 hour',
  'redis.ttl.1day': '1 day',
  'redis.ttl.1week': '1 week',
  'redis.ttl.never': 'Never expire',
  'redis.ttl.custom': 'Custom',

  // MongoDB (preparation — P0 not yet implemented)
  'mongodb.form.uri': 'Connection URI',
  'mongodb.form.uri_hint': 'Use mongodb:// or mongodb+srv:// format. Leave empty to use host:port below.',
  'mongodb.form.or_basic': '— or use basic host:port —',
  'mongodb.form.auth_db': 'Authentication Database',
  'mongodb.form.tls': 'Use TLS/SSL',
  'mongodb.driver_coming_soon': 'MongoDB document browsing will be available in a future release.',
  'mongodb.readonly': 'READ ONLY',
  'mongodb.coming_soon': 'MongoDB Document Browser',
  'mongodb.coming_soon_hint': 'Document browsing, filtering, and editing will be available soon.',
  'mongodb.p0_features': 'Planned: document list, JSON filter, insert/edit/delete, index viewer',
  'connections.filter.mongodb': 'MongoDB',

  // MongoDB P0 — Document Browser
  'mongodb.browser.title': 'Document Browser',
  'mongodb.browser.filter': 'Filter',
  'mongodb.browser.filter_placeholder': 'JSON filter (e.g., {"age": {"$gt": 18}})',
  'mongodb.browser.projection': 'Projection',
  'mongodb.browser.projection_placeholder': 'JSON projection (e.g., {"name": 1, "age": 1})',
  'mongodb.browser.sort': 'Sort',
  'mongodb.browser.sort_placeholder': 'JSON sort (e.g., {"createdAt": -1})',
  'mongodb.browser.limit': 'Limit',
  'mongodb.browser.skip': 'Skip',
  'mongodb.browser.run_query': 'Run Query',
  'mongodb.browser.querying': 'Querying...',
  'mongodb.browser.insert': 'Insert Document',
  'mongodb.browser.refresh': 'Refresh',
  'mongodb.browser.no_documents': 'No documents found',
  'mongodb.browser.total_documents': 'Total: {count} documents',
  'mongodb.browser.showing': 'Showing {start}-{end} of {total}',
  'mongodb.browser.load_more': 'Load More',
  'mongodb.browser.view_indexes': 'View Indexes',
  'mongodb.browser.query_error': 'Query Error',
  'mongodb.browser.invalid_json': 'Invalid JSON in {field}: {error}',

  // MongoDB P0 — Document Viewer/Editor
  'mongodb.document.viewer_title': 'Document Viewer',
  'mongodb.document.editor_title': 'Edit Document',
  'mongodb.document.insert_title': 'Insert New Document',
  'mongodb.document.save': 'Save Changes',
  'mongodb.document.insert': 'Insert',
  'mongodb.document.delete': 'Delete Document',
  'mongodb.document.delete_confirm': 'Are you sure you want to delete this document? This action cannot be undone.',
  'mongodb.document.copied': 'Document copied to clipboard',
  'mongodb.document.saved': 'Document saved successfully',
  'mongodb.document.inserted': 'Document inserted successfully',
  'mongodb.document.deleted': 'Document deleted successfully',
  'mongodb.document.save_error': 'Failed to save document',
  'mongodb.document.insert_error': 'Failed to insert document',
  'mongodb.document.delete_error': 'Failed to delete document',
  'mongodb.document.invalid_json': 'Invalid JSON format',
  'mongodb.document.objectid': 'ObjectID',

  // MongoDB P0 — Indexes
  'mongodb.indexes.title': 'Indexes',
  'mongodb.indexes.name': 'Index Name',
  'mongodb.indexes.keys': 'Keys',
  'mongodb.indexes.unique': 'Unique',
  'mongodb.indexes.sparse': 'Sparse',
  'mongodb.indexes.no_indexes': 'No indexes found',
  'mongodb.indexes.load_error': 'Failed to load indexes',

  // MongoDB P0 — Common
  'mongodb.loading': 'Loading...',
  'mongodb.error': 'Error',
  'mongodb.retry': 'Retry',
  'mongodb.close': 'Close',
  'mongodb.cancel': 'Cancel',
  'mongodb.confirm': 'Confirm',
  'mongodb.copy': 'Copy',
  'mongodb.edit': 'Edit',
  'mongodb.delete': 'Delete',
  'mongodb.view': 'View',

  // MongoDB P1 - Tabs
  'mongodb.p1.tab.documents': 'Documents',
  'mongodb.p1.tab.aggregation': 'Aggregation',
  'mongodb.p1.tab.schema': 'Schema Analysis',
  'mongodb.p1.tab.stats': 'Statistics',

  // MongoDB P1 - Aggregation
  'mongodb.p1.aggregation.title': 'Aggregation Pipeline',
  'mongodb.p1.aggregation.pipeline_placeholder': 'Enter aggregation pipeline JSON array, e.g.:\n[{"$match": {"age": {"$gt": 18}}}, {"$group": {"_id": "$city", "count": {"$sum": 1}}}]',
  'mongodb.p1.aggregation.run': 'Run Aggregation',
  'mongodb.p1.aggregation.running': 'Running...',
  'mongodb.p1.aggregation.results': 'Aggregation Results',
  'mongodb.p1.aggregation.result_count': 'Returned {count} documents',
  'mongodb.p1.aggregation.duration': 'Duration: {ms} ms',
  'mongodb.p1.aggregation.no_results': 'Aggregation returned no results',
  'mongodb.p1.aggregation.error': 'Aggregation failed',
  'mongodb.p1.aggregation.invalid_json': 'Invalid aggregation pipeline JSON',
  'mongodb.p1.aggregation.empty_pipeline': 'Aggregation pipeline cannot be empty',

  // MongoDB P1 - Explain
  'mongodb.p1.explain.title': 'Query Execution Plan',
  'mongodb.p1.explain.filter_placeholder': 'Query filter JSON (optional)',
  'mongodb.p1.explain.run': 'Analyze Query',
  'mongodb.p1.explain.running': 'Analyzing...',
  'mongodb.p1.explain.winning_plan': 'Execution Plan',
  'mongodb.p1.explain.stage': 'Stage',
  'mongodb.p1.explain.index': 'Index',
  'mongodb.p1.explain.docs_examined': 'Docs Examined',
  'mongodb.p1.explain.docs_returned': 'Docs Returned',
  'mongodb.p1.explain.execution_time': 'Execution Time',
  'mongodb.p1.explain.no_plan': 'No execution plan data',
  'mongodb.p1.explain.error': 'Query analysis failed',

  // MongoDB P1 - Schema
  'mongodb.p1.schema.title': 'Schema Analysis',
  'mongodb.p1.schema.description': 'Analyze field structure and type distribution in collection documents',
  'mongodb.p1.schema.sample_size': 'Sample Size',
  'mongodb.p1.schema.analyze': 'Analyze Schema',
  'mongodb.p1.schema.analyzing': 'Analyzing...',
  'mongodb.p1.schema.fields': 'Fields',
  'mongodb.p1.schema.field_name': 'Field Name',
  'mongodb.p1.schema.field_type': 'Type',
  'mongodb.p1.schema.field_count': 'Count',
  'mongodb.p1.schema.field_percentage': 'Frequency',
  'mongodb.p1.schema.no_fields': 'No fields found',
  'mongodb.p1.schema.error': 'Schema analysis failed',
  'mongodb.p1.schema.sampled_docs': 'Analyzed {count} documents',

  // MongoDB P1 - Stats
  'mongodb.p1.stats.title': 'Collection Statistics',
  'mongodb.p1.stats.load': 'Load Stats',
  'mongodb.p1.stats.loading': 'Loading...',
  'mongodb.p1.stats.document_count': 'Document Count',
  'mongodb.p1.stats.collection_size': 'Collection Size',
  'mongodb.p1.stats.avg_doc_size': 'Average Document Size',
  'mongodb.p1.stats.storage_size': 'Storage Size',
  'mongodb.p1.stats.index_count': 'Index Count',
  'mongodb.p1.stats.index_size': 'Total Index Size',
  'mongodb.p1.stats.indexes': 'Index Details',
  'mongodb.p1.stats.index_name': 'Index Name',
  'mongodb.p1.stats.index_size_col': 'Size',
  'mongodb.p1.stats.no_indexes': 'No indexes',
  'mongodb.p1.stats.error': 'Failed to load statistics',

  // MongoDB P1 - Export
  'mongodb.p1.export.title': 'Export Documents',
  'mongodb.p1.export.format': 'Export Format',
  'mongodb.p1.export.format_json': 'JSON (Formatted)',
  'mongodb.p1.export.format_ndjson': 'NDJSON (One document per line)',
  'mongodb.p1.export.format_csv': 'CSV',
  'mongodb.p1.export.filter': 'Filter (optional)',
  'mongodb.p1.export.filter_placeholder': 'JSON filter, leave empty to export all documents',
  'mongodb.p1.export.limit': 'Limit',
  'mongodb.p1.export.limit_hint': 'Maximum {max} documents',
  'mongodb.p1.export.export': 'Export',
  'mongodb.p1.export.exporting': 'Exporting...',
  'mongodb.p1.export.success': 'Export successful',
  'mongodb.p1.export.error': 'Export failed',
  'mongodb.p1.export.no_data': 'No data to export',

  // MongoDB P1 - Import
  'mongodb.p1.import.title': 'Import Documents',
  'mongodb.p1.import.format': 'Import Format',
  'mongodb.p1.import.format_json': 'JSON Array',
  'mongodb.p1.import.format_ndjson': 'NDJSON (One document per line)',
  'mongodb.p1.import.data': 'Document Data',
  'mongodb.p1.import.data_placeholder_json': 'JSON array format, e.g.:\n[{"name": "Alice", "age": 30}, {"name": "Bob", "age": 25}]',
  'mongodb.p1.import.data_placeholder_ndjson': 'NDJSON format, one document per line, e.g.:\n{"name": "Alice", "age": 30}\n{"name": "Bob", "age": 25}',
  'mongodb.p1.import.import': 'Import',
  'mongodb.p1.import.importing': 'Importing...',
  'mongodb.p1.import.success': 'Successfully imported {count} documents',
  'mongodb.p1.import.error': 'Import failed',
  'mongodb.p1.import.invalid_json': 'Invalid JSON data format',
  'mongodb.p1.import.empty_data': 'Import data cannot be empty',
  'mongodb.p1.import.confirm': 'Confirm Import',
  'mongodb.p1.import.confirm_message': 'About to import {count} documents to collection "{collection}". This operation cannot be undone. Continue?',

  // MongoDB P1 - History
  'mongodb.p1.history.title': 'Query History',
  'mongodb.p1.history.no_history': 'No query history',
  'mongodb.p1.history.clear_all': 'Clear All',
  'mongodb.p1.history.delete': 'Delete',
  'mongodb.p1.history.load_more': 'Load More',
  'mongodb.p1.history.error': 'Failed to load history',
  'mongodb.p1.history.run': 'Run',
  'mongodb.p1.history.running': 'Running...',
  'mongodb.p1.history.query': 'Query',
  'mongodb.p1.history.duration': 'Duration',
  'mongodb.p1.history.result_count': 'Results',
  'mongodb.p1.history.timestamp': 'Time',
  'mongodb.p1.history.clear_confirm': 'Are you sure you want to clear all query history? This cannot be undone.',
  'mongodb.p1.history.delete_confirm': 'Are you sure you want to delete this query history entry?',
  'mongodb.p1.history.cleared': 'Query history cleared',
  'mongodb.p1.history.deleted': 'Query history deleted',

  // MongoDB P1 - ObjectId Parser
  'mongodb.p1.objectid.title': 'ObjectID Parser',
  'mongodb.p1.objectid.input_placeholder': 'Enter 24-character hex ObjectID',
  'mongodb.p1.objectid.parse': 'Parse',
  'mongodb.p1.objectid.parsing': 'Parsing...',
  'mongodb.p1.objectid.hex': 'Hex',
  'mongodb.p1.objectid.timestamp': 'Timestamp',
  'mongodb.p1.objectid.process_id': 'Process ID',
  'mongodb.p1.objectid.machine_id': 'Machine ID',
  'mongodb.p1.objectid.counter': 'Counter',
  'mongodb.p1.objectid.invalid': 'Invalid ObjectID format',
  'mongodb.p1.objectid.error': 'Failed to parse ObjectID',

  // MongoDB P1 - Common
  'mongodb.p1.common.load_error': 'Failed to load',
  'mongodb.p1.common.retry': 'Retry',
  'mongodb.p1.common.refresh': 'Refresh',
  'mongodb.p1.common.no_data': 'No data',
  'mongodb.p1.common.bytes': 'bytes',
  'mongodb.p1.common.kb': 'KB',
  'mongodb.p1.common.mb': 'MB',
  'mongodb.p1.common.gb': 'GB',
  'mongodb.p1.common.ms': 'ms',
  'mongodb.p1.common.seconds': 'seconds',
  'mongodb.p1.common.minutes': 'minutes',
}

const zh: Dict = {
  // Login
  'login.username': '用户名',
  'login.password': '密码',
  'login.submit': '登录',
  'login.submitting': '登录中…',

  // Nav
  'nav.manage_connections': '+ 管理连接',
  'nav.sql_console': 'SQL 控制台',
  'nav.logout': '退出',

  // Connections
  'connections.title': '连接管理',
  'connections.add': '+ 添加连接',
  'connections.empty': '暂无连接，请添加第一个数据库连接。',
  'connections.test': '测试',
  'connections.edit': '编辑',
  'connections.delete': '删除',
  'connections.add_title': '添加连接',
  'connections.edit_title': '编辑连接',
  'connections.form.name': '名称',
  'connections.form.driver': '驱动',
  'connections.form.file_path': '文件路径',
  'connections.form.host': '主机',
  'connections.form.port': '端口',
  'connections.form.database': '数据库',
  'connections.form.username': '用户名',
  'connections.form.password': '密码',
  'connections.form.save': '保存',
  'connections.form.cancel': '取消',
  'connections.open': '打开',
  'connections.delete.title': '删除连接',
  'connections.delete.message': '删除 "{name}"？此操作不可撤销。',
  'connections.form.readonly': '只读连接',
  'connections.form.readonly_hint': '禁止所有写操作',
  'connections.form.save_password': '保存密码',
  'connections.form.save_password_warn': '密码将加密存储在服务端数据目录。不建议在无访问控制的共享机器上使用。',

  // Tables
  'tables.sql_console': '▶ SQL 控制台',
  'tables.new_table': '+ 新建表',
  'tables.col.name': '表名',
  'tables.col.type': '类型',
  'tables.col.engine': '引擎',
  'tables.col.rows': '行数',
  'tables.col.comment': '注释',
  'tables.action.data': '数据',
  'tables.action.fields': '字段',
  'tables.action.indexes': '索引',
  'tables.action.drop': '删除',
  'tables.empty': '暂无表',
  'tables.drop.title': '删除表',
  'tables.drop.message': '删除表 "{table}"？此操作不可撤销。',
  'tables.create.title': '新建表',
  'tables.create.hint': '推荐使用 SQL 控制台编写完整 DDL。',
  'tables.create.open_console': '打开 SQL 控制台',

  // Data
  'data.filter.placeholder': '列:值 过滤',
  'data.filter.apply': '过滤',
  'data.refresh': '刷新',
  'data.filter.show': '过滤',
  'data.filter.clear_all': '清空',
  'data.filter.add': '添加',
  'data.filter.col_placeholder': '选择列…',
  'data.filter.val_placeholder': '值',
  'data.insert': '+ 插入',
  'data.prev': '← 上一页',
  'data.next': '下一页 →',
  'data.edit': '编辑',
  'data.delete': '删除',
  'data.delete.title': '删除行',
  'data.delete.message': '删除 {col} = {val} 的行？',
  'data.insert.title': '插入行',
  'data.edit.title': '编辑行',
  'data.save': '保存',
  'data.cancel': '取消',
  'data.no_rows': '无数据',
  'data.loading': '加载中…',
  'data.no_pk_hint': '无主键 — 编辑和删除不可用',
  'data.load_failed': '加载失败',
  'data.copy_failed': '复制失败',
  'data.delete_failed': '删除失败',
  'data.save_failed': '保存失败',
  'data.insert.success': '行已插入',
  'data.edit.success': '行已更新',
  'data.pagination': '第 {page} / {total} 页',
  'data.page_size': '{n} 条/页',

  // Structure
  'structure.tab.columns': '字段',
  'structure.tab.indexes': '索引',
  'structure.tab.fks': '外键',
  'structure.tab.ddl': 'DDL',
  'structure.no_indexes': '无索引',
  'structure.no_fks': '无外键',

  // Query
  'query.title': 'SQL 控制台',
  'query.db_placeholder': '-- 选择数据库 --',
  'query.run': '▶ 运行 (Cmd+Enter)',
  'query.running': '运行中…',
  'query.copy': '复制 SQL',
  'query.clear': '清空',
  'query.confirm_danger': '执行危险 SQL？',
  'query.danger_reason': '原因',
  'query.confirm_proceed': '继续执行',
  'query.rows_affected': '行受影响',
  'query.last_insert_id': '最后插入 ID',
  'query.truncated': '结果已截断为 1000 行，请添加 LIMIT',
  'query.tab.result': '结果',
  'query.tab.history': '历史',
  'query.history.empty': '暂无历史记录',

  // Readonly
  'readonly.badge': '只读',
  'readonly.violation': '此连接为只读模式',

  // Export
  'export.title': '导出',
  'export.format': '格式',
  'export.include_schema': '包含结构 (DDL)',
  'export.include_data': '包含数据',
  'export.download': '下载',

  // Import
  'import.title': '导入 SQL',
  'import.paste': '粘贴 SQL 或上传 .sql 文件',
  'import.file': '上传文件',
  'import.run': '导入',
  'import.running': '导入中…',
  'import.confirm_danger': '导入包含危险 SQL',
  'import.confirm_proceed': '继续执行',

  // Tables toolbar
  'tables.import': '导入 SQL',
  'tables.copy_schema': '复制 Schema',
  'tables.copy_schema.success': '已复制 Schema',
  'tables.copy_schema.loading': '复制中…',

  // Connections filter
  'connections.filter.all': '全部',
  'connections.filter.mysql': 'MySQL',
  'connections.filter.sqlite': 'SQLite',
  'connections.filter.redis': 'Redis',
  'connections.duplicate': '复制',

  // Redis
  'redis.form.password': 'AUTH 密码',
  'redis.form.password_placeholder': '可选，无 AUTH 时留空',
  'redis.form.db_index': 'DB 索引',
  'redis.form.db_index_hint': '0–15，默认 0',
  'redis.form.tls': '使用 TLS / SSL',
  'redis.readonly': '只读',
  'redis.driver_coming_soon': 'Redis 驱动将在下一个版本发布。连接配置已保存。',
  'redis.panel.title': 'Redis Key 浏览器',
  'redis.panel.hint': 'Key 浏览、TTL 查看和命令执行功能将在 Redis 驱动发布后可用。',

  // Settings
  'settings.title': '设置',
  'settings.sql_history': 'SQL 历史',
  'settings.sql_history_enabled': '启用 SQL 历史',
  'settings.sql_history_hint': '执行的 SQL 将保存在服务端（每个连接最多 100 条，保留 30 天）。可在 SQL Console 历史 tab 中管理。',
  'nav.settings': '设置',

  // History (localStorage)
  'history.tab': '历史',
  'history.empty': '暂无历史记录',
  'history.clear_all': '清空',
  'history.disabled': 'SQL 历史已关闭，请在设置中启用。',
  'history.delete': '删除',

  // Cell viewer
  'cell.tab.raw': 'Raw',
  'cell.tab.pretty': '格式化',
  'cell.copy': '复制',
  'cell.copy_pretty': '复制格式化',
  'cell.copy_minified': '复制压缩',
  'cell.empty': '空字符串',
  'cell.blob': 'BLOB',
  'cell.blob_note': '二进制数据，下载功能将在后续版本提供。',

  // Copy
  'copy.as': '复制为：',
  'copy.markdown': 'Markdown',
  'copy.json_schema': 'JSON Schema',
  'copy.ddl': '复制 DDL',
  'copy.json': 'JSON',
  'copy.csv': 'CSV',
  'copy.insert': 'SQL INSERT',
  'copy.rows_selected': '复制 {n} 行：',
  'copy.cell_success': '已复制',

  // SQLite upload/download
  'sqlite.mode_path': '服务端文件路径',
  'sqlite.mode_upload': '上传文件',
  'sqlite.upload_choose': '选择 .db / .sqlite 文件…',
  'sqlite.uploading': '上传中…',
  'sqlite.upload_failed': '上传失败',
  'sqlite.replace': '替换',
  'sqlite.download': '下载 SQLite',
  'sqlite.delete_file': '同时从服务器删除上传的 SQLite 文件',
  'sqlite.upload_hint': '文件必须是有效的 SQLite 数据库，最大 512 MiB。',

  // Confirm
  'confirm.cancel': '取消',
  'confirm.delete': '删除',
  'confirm.drop': '删除',

  // Redis Key Browser
  'redis.browser.pattern': '匹配模式',
  'redis.browser.pattern_placeholder': 'key:* 过滤（默认 *）',
  'redis.browser.search': '搜索',
  'redis.browser.load_more': '加载更多',
  'redis.browser.empty': '未找到 Key',
  'redis.browser.loading': '加载中…',
  'redis.browser.key_count': '{n} 个 Key',
  'redis.browser.db': 'DB {n}',
  'redis.browser.ttl_none': '永不过期',
  'redis.browser.ttl_sec': '{n}s',

  // Redis Key Inspector
  'redis.key.ttl': 'TTL',
  'redis.key.type': '类型',
  'redis.key.encoding': '编码',
  'redis.key.set_ttl': '设置 TTL',
  'redis.key.ttl_hint': '秒；-1 表示删除过期',
  'redis.key.delete': '删除 Key',
  'redis.key.delete_confirm': '删除 Key "{key}"？此操作不可撤销。',
  'redis.key.copy': '复制',
  'redis.key.loading': '加载中…',
  'redis.key.not_found': 'Key 不存在',
  'redis.key.stream_length': '长度：{n}',
  'redis.key.stream_groups': '消费组：{n}',
  'redis.key.list_length': '{n} 个元素',
  'redis.key.set_size': '{n} 个成员',
  'redis.key.hash_fields': '{n} 个字段',
  'redis.key.hash_search': '搜索字段...',
  'redis.key.hash_no_match': '没有匹配的字段',
  'redis.key.page': '第 {current} 页，共 {total} 页',
  'redis.key.prev': '上一页',
  'redis.key.next': '下一页',

  // Redis Command Console
  'redis.command.title': '命令控制台',
  'redis.command.placeholder': '例如 GET mykey',
  'redis.command.run': '执行',
  'redis.command.running': '执行中…',
  'redis.command.readonly_blocked': '已拦截：只读连接不允许写命令',
  'redis.command.forbidden': '禁止的命令',

  // Redis Batch Operations
  'redis.batch.select_keys': '选择要删除的 Key',
  'redis.batch.cancel': '取消选择',
  'redis.batch.selected': '已选择 {n} 个 Key',
  'redis.batch.preview': '预览',
  'redis.batch.delete': '删除选中',
  'redis.batch.confirm_delete': '删除 {n} 个选中的 Key？此操作不可撤销。',
  'redis.batch.deleted': '已删除 {n} 个 Key',
  'redis.batch.no_keys': '没有匹配该模式的 Key',
  'redis.batch.warning': '警告：这将永久删除 {n} 个 Key',
  'redis.batch.select_all': '全选',
  'redis.batch.deselect_all': '取消全选',

  // Redis Pattern Favorites
  'redis.favorites.title': '模式收藏',
  'redis.favorites.save': '保存模式',
  'redis.favorites.saved': '模式已保存',
  'redis.favorites.no_favorites': '暂无收藏模式',
  'redis.favorites.delete': '从收藏中移除',
  'redis.favorites.name': '模式名称',

  // Redis Big Key Warning
  'redis.big_key.warning': '检测到大 Key（{size}）',
  'redis.big_key.confirm_load': '此 Key 大小为 {size}，加载可能会拖慢界面。是否继续？',
  'redis.big_key.skip': '跳过大 value 加载',

  // Redis Memory
  'redis.memory': '内存：{size}',
  'redis.memory.bytes': '{n} B',
  'redis.memory.kb': '{n} KB',
  'redis.memory.mb': '{n} MB',
  'redis.memory.gb': '{n} GB',

  // Redis Hash Field Search
  'redis.hash.search': '搜索字段',
  'redis.hash.search_placeholder': '按字段名过滤…',
  'redis.hash.no_match': '没有匹配的字段',

  // Redis List/Set/ZSet Pagination
  'redis.collection.range': '显示 {start}-{end}，共 {total}',
  'redis.collection.load_more': '加载更多',
  'redis.collection.first': '首页',
  'redis.collection.prev': '上一页',
  'redis.collection.next': '下一页',

  // Redis Stream
  'redis.stream.id': 'ID',
  'redis.stream.fields': '字段',
  'redis.stream.no_messages': '流中没有消息',

  // Redis Copy
  'redis.copy.key': '复制 Key',
  'redis.copy.value': '复制 Value',
  'redis.copy.key_success': 'Key 已复制到剪贴板',
  'redis.copy.value_success': 'Value 已复制到剪贴板',

  // Redis TTL Presets
  'redis.ttl.presets': 'TTL 预设',
  'redis.ttl.1min': '1 分钟',
  'redis.ttl.5min': '5 分钟',
  'redis.ttl.1hour': '1 小时',
  'redis.ttl.1day': '1 天',
  'redis.ttl.1week': '1 周',
  'redis.ttl.never': '永不过期',
  'redis.ttl.custom': '自定义',

  // MongoDB (preparation — P0 not yet implemented)
  'mongodb.form.uri': '连接 URI',
  'mongodb.form.uri_hint': '使用 mongodb:// 或 mongodb+srv:// 格式。留空则使用下方的 host:port。',
  'mongodb.form.or_basic': '— 或使用基本 host:port —',
  'mongodb.form.auth_db': '认证数据库',
  'mongodb.form.tls': '使用 TLS/SSL',
  'mongodb.driver_coming_soon': 'MongoDB 文档浏览功能将在后续版本中提供。',
  'mongodb.readonly': '只读',
  'mongodb.coming_soon': 'MongoDB 文档浏览器',
  'mongodb.coming_soon_hint': '文档浏览、过滤和编辑功能即将推出。',
  'mongodb.p0_features': '计划功能：文档列表、JSON 过滤、插入/编辑/删除、索引查看',
  'connections.filter.mongodb': 'MongoDB',

  // MongoDB P0 — Document Browser
  'mongodb.browser.title': '文档浏览器',
  'mongodb.browser.filter': '过滤器',
  'mongodb.browser.filter_placeholder': 'JSON 过滤条件（例如：{"age": {"$gt": 18}}）',
  'mongodb.browser.projection': '投影',
  'mongodb.browser.projection_placeholder': 'JSON 投影（例如：{"name": 1, "age": 1}）',
  'mongodb.browser.sort': '排序',
  'mongodb.browser.sort_placeholder': 'JSON 排序（例如：{"createdAt": -1}）',
  'mongodb.browser.limit': '限制',
  'mongodb.browser.skip': '跳过',
  'mongodb.browser.run_query': '执行查询',
  'mongodb.browser.querying': '查询中...',
  'mongodb.browser.insert': '插入文档',
  'mongodb.browser.refresh': '刷新',
  'mongodb.browser.no_documents': '未找到文档',
  'mongodb.browser.total_documents': '总计：{count} 个文档',
  'mongodb.browser.showing': '显示 {start}-{end}，共 {total}',
  'mongodb.browser.load_more': '加载更多',
  'mongodb.browser.view_indexes': '查看索引',
  'mongodb.browser.query_error': '查询错误',
  'mongodb.browser.invalid_json': '{field} 中的 JSON 无效：{error}',

  // MongoDB P0 — Document Viewer/Editor
  'mongodb.document.viewer_title': '文档查看器',
  'mongodb.document.editor_title': '编辑文档',
  'mongodb.document.insert_title': '插入新文档',
  'mongodb.document.save': '保存更改',
  'mongodb.document.insert': '插入',
  'mongodb.document.delete': '删除文档',
  'mongodb.document.delete_confirm': '确定要删除此文档吗？此操作无法撤销。',
  'mongodb.document.copied': '文档已复制到剪贴板',
  'mongodb.document.saved': '文档保存成功',
  'mongodb.document.inserted': '文档插入成功',
  'mongodb.document.deleted': '文档删除成功',
  'mongodb.document.save_error': '保存文档失败',
  'mongodb.document.insert_error': '插入文档失败',
  'mongodb.document.delete_error': '删除文档失败',
  'mongodb.document.invalid_json': 'JSON 格式无效',
  'mongodb.document.objectid': 'ObjectID',

  // MongoDB P0 — Indexes
  'mongodb.indexes.title': '索引',
  'mongodb.indexes.name': '索引名称',
  'mongodb.indexes.keys': '键',
  'mongodb.indexes.unique': '唯一',
  'mongodb.indexes.sparse': '稀疏',
  'mongodb.indexes.no_indexes': '未找到索引',
  'mongodb.indexes.load_error': '加载索引失败',

  // MongoDB P0 — Common
  'mongodb.loading': '加载中...',
  'mongodb.error': '错误',
  'mongodb.retry': '重试',
  'mongodb.close': '关闭',
  'mongodb.cancel': '取消',
  'mongodb.confirm': '确认',
  'mongodb.copy': '复制',
  'mongodb.edit': '编辑',
  'mongodb.delete': '删除',
  'mongodb.view': '查看',

  // MongoDB P1 - Tabs
  'mongodb.p1.tab.documents': '文档',
  'mongodb.p1.tab.aggregation': '聚合',
  'mongodb.p1.tab.schema': '模式分析',
  'mongodb.p1.tab.stats': '统计信息',

  // MongoDB P1 - Aggregation
  'mongodb.p1.aggregation.title': '聚合管道',
  'mongodb.p1.aggregation.pipeline_placeholder': '输入聚合管道 JSON 数组，例如：\n[{"$match": {"age": {"$gt": 18}}}, {"$group": {"_id": "$city", "count": {"$sum": 1}}}]',
  'mongodb.p1.aggregation.run': '运行聚合',
  'mongodb.p1.aggregation.running': '运行中...',
  'mongodb.p1.aggregation.results': '聚合结果',
  'mongodb.p1.aggregation.result_count': '返回 {count} 个文档',
  'mongodb.p1.aggregation.duration': '耗时 {ms} 毫秒',
  'mongodb.p1.aggregation.no_results': '聚合未返回结果',
  'mongodb.p1.aggregation.error': '聚合执行失败',
  'mongodb.p1.aggregation.invalid_json': '聚合管道 JSON 无效',
  'mongodb.p1.aggregation.empty_pipeline': '聚合管道不能为空',

  // MongoDB P1 - Explain
  'mongodb.p1.explain.title': '查询执行计划',
  'mongodb.p1.explain.filter_placeholder': '查询过滤条件 JSON（可选）',
  'mongodb.p1.explain.run': '分析查询',
  'mongodb.p1.explain.running': '分析中...',
  'mongodb.p1.explain.winning_plan': '执行计划',
  'mongodb.p1.explain.stage': '阶段',
  'mongodb.p1.explain.index': '索引',
  'mongodb.p1.explain.docs_examined': '检查文档数',
  'mongodb.p1.explain.docs_returned': '返回文档数',
  'mongodb.p1.explain.execution_time': '执行时间',
  'mongodb.p1.explain.no_plan': '无执行计划数据',
  'mongodb.p1.explain.error': '查询分析失败',

  // MongoDB P1 - Schema
  'mongodb.p1.schema.title': '模式分析',
  'mongodb.p1.schema.description': '分析集合中文档的字段结构和类型分布',
  'mongodb.p1.schema.sample_size': '采样数量',
  'mongodb.p1.schema.analyze': '分析模式',
  'mongodb.p1.schema.analyzing': '分析中...',
  'mongodb.p1.schema.fields': '字段',
  'mongodb.p1.schema.field_name': '字段名',
  'mongodb.p1.schema.field_type': '类型',
  'mongodb.p1.schema.field_count': '出现次数',
  'mongodb.p1.schema.field_percentage': '出现率',
  'mongodb.p1.schema.no_fields': '未找到字段',
  'mongodb.p1.schema.error': '模式分析失败',
  'mongodb.p1.schema.sampled_docs': '已分析 {count} 个文档',

  // MongoDB P1 - Stats
  'mongodb.p1.stats.title': '集合统计',
  'mongodb.p1.stats.load': '加载统计',
  'mongodb.p1.stats.loading': '加载中...',
  'mongodb.p1.stats.document_count': '文档数量',
  'mongodb.p1.stats.collection_size': '集合大小',
  'mongodb.p1.stats.avg_doc_size': '平均文档大小',
  'mongodb.p1.stats.storage_size': '存储大小',
  'mongodb.p1.stats.index_count': '索引数量',
  'mongodb.p1.stats.index_size': '索引总大小',
  'mongodb.p1.stats.indexes': '索引详情',
  'mongodb.p1.stats.index_name': '索引名称',
  'mongodb.p1.stats.index_size_col': '大小',
  'mongodb.p1.stats.no_indexes': '无索引',
  'mongodb.p1.stats.error': '加载统计信息失败',

  // MongoDB P1 - Export
  'mongodb.p1.export.title': '导出文档',
  'mongodb.p1.export.format': '导出格式',
  'mongodb.p1.export.format_json': 'JSON（格式化）',
  'mongodb.p1.export.format_ndjson': 'NDJSON（每行一个文档）',
  'mongodb.p1.export.format_csv': 'CSV',
  'mongodb.p1.export.filter': '过滤条件（可选）',
  'mongodb.p1.export.filter_placeholder': 'JSON 过滤条件，留空导出所有文档',
  'mongodb.p1.export.limit': '限制数量',
  'mongodb.p1.export.limit_hint': '最多导出 {max} 个文档',
  'mongodb.p1.export.export': '导出',
  'mongodb.p1.export.exporting': '导出中...',
  'mongodb.p1.export.success': '导出成功',
  'mongodb.p1.export.error': '导出失败',
  'mongodb.p1.export.no_data': '没有可导出的数据',

  // MongoDB P1 - Import
  'mongodb.p1.import.title': '导入文档',
  'mongodb.p1.import.format': '导入格式',
  'mongodb.p1.import.format_json': 'JSON 数组',
  'mongodb.p1.import.format_ndjson': 'NDJSON（每行一个文档）',
  'mongodb.p1.import.data': '文档数据',
  'mongodb.p1.import.data_placeholder_json': 'JSON 数组格式，例如：\n[{"name": "Alice", "age": 30}, {"name": "Bob", "age": 25}]',
  'mongodb.p1.import.data_placeholder_ndjson': 'NDJSON 格式，每行一个文档，例如：\n{"name": "Alice", "age": 30}\n{"name": "Bob", "age": 25}',
  'mongodb.p1.import.import': '导入',
  'mongodb.p1.import.importing': '导入中...',
  'mongodb.p1.import.success': '成功导入 {count} 个文档',
  'mongodb.p1.import.error': '导入失败',
  'mongodb.p1.import.invalid_json': 'JSON 数据格式无效',
  'mongodb.p1.import.empty_data': '导入数据不能为空',
  'mongodb.p1.import.confirm': '确认导入',
  'mongodb.p1.import.confirm_message': '即将导入 {count} 个文档到集合 "{collection}"，此操作无法撤销。是否继续？',

  // MongoDB P1 - History
  'mongodb.p1.history.title': '查询历史',
  'mongodb.p1.history.no_history': '暂无查询历史',
  'mongodb.p1.history.clear_all': '清空历史',
  'mongodb.p1.history.delete': '删除',
  'mongodb.p1.history.load_more': '加载更多',
  'mongodb.p1.history.error': '加载历史失败',
  'mongodb.p1.history.run': '运行',
  'mongodb.p1.history.running': '运行中...',
  'mongodb.p1.history.query': '查询',
  'mongodb.p1.history.duration': '耗时',
  'mongodb.p1.history.result_count': '结果数',
  'mongodb.p1.history.timestamp': '时间',
  'mongodb.p1.history.clear_confirm': '确定要清空所有查询历史吗？此操作无法撤销。',
  'mongodb.p1.history.delete_confirm': '确定要删除这条查询历史吗？',
  'mongodb.p1.history.cleared': '查询历史已清空',
  'mongodb.p1.history.deleted': '查询历史已删除',

  // MongoDB P1 - ObjectId Parser
  'mongodb.p1.objectid.title': 'ObjectID 解析',
  'mongodb.p1.objectid.input_placeholder': '输入 24 位十六进制 ObjectID',
  'mongodb.p1.objectid.parse': '解析',
  'mongodb.p1.objectid.parsing': '解析中...',
  'mongodb.p1.objectid.hex': '十六进制',
  'mongodb.p1.objectid.timestamp': '时间戳',
  'mongodb.p1.objectid.process_id': '进程 ID',
  'mongodb.p1.objectid.machine_id': '机器 ID',
  'mongodb.p1.objectid.counter': '计数器',
  'mongodb.p1.objectid.invalid': '无效的 ObjectID 格式',
  'mongodb.p1.objectid.error': '解析 ObjectID 失败',

  // MongoDB P1 - Common
  'mongodb.p1.common.load_error': '加载失败',
  'mongodb.p1.common.retry': '重试',
  'mongodb.p1.common.refresh': '刷新',
  'mongodb.p1.common.no_data': '暂无数据',
  'mongodb.p1.common.bytes': '字节',
  'mongodb.p1.common.kb': 'KB',
  'mongodb.p1.common.mb': 'MB',
  'mongodb.p1.common.gb': 'GB',
  'mongodb.p1.common.ms': '毫秒',
  'mongodb.p1.common.seconds': '秒',
  'mongodb.p1.common.minutes': '分钟',
}

const dicts: Record<Locale, Dict> = { en, zh }

interface I18nContextValue {
  lang: Locale
  setLang: (l: Locale) => void
  t: (key: string, vars?: Record<string, string | number>) => string
}

const I18nContext = createContext<I18nContextValue>({
  lang: 'en',
  setLang: () => {},
  t: k => k,
})

export function I18nProvider({ children }: { children: ReactNode }) {
  const [lang, setLangState] = useState<Locale>(() =>
    (localStorage.getItem('dbadmin_lang') as Locale) || 'en'
  )

  const setLang = (l: Locale) => {
    setLangState(l)
    localStorage.setItem('dbadmin_lang', l)
  }

  const t = useCallback((key: string, vars?: Record<string, string | number>) => {
    let str = dicts[lang][key] ?? dicts['en'][key] ?? key
    if (vars) {
      Object.entries(vars).forEach(([k, v]) => {
        str = str.replaceAll(`{${k}}`, String(v))
      })
    }
    return str
  }, [lang])

  return (
    <I18nContext.Provider value={{ lang, setLang, t }}>
      {children}
    </I18nContext.Provider>
  )
}

export const useI18n = () => useContext(I18nContext)
