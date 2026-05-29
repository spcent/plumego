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
