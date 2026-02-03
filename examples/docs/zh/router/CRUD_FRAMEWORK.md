# CRUD Framework - Database Integration Guide

完整的数据库 CRUD 框架,将 RESTful API 模式与 `store/db` 包深度集成,提供企业级的可复用组件。

## 目录

- [核心概念](#核心概念)
- [快速开始](#快速开始)
- [组件详解](#组件详解)
- [完整示例](#完整示例)
- [最佳实践](#最佳实践)
- [API 参考](#api-参考)

## 核心概念

### 架构图

```
HTTP Request
    ↓
DBResourceController (控制器层)
    ├── QueryBuilder (查询参数解析)
    ├── Validator (数据验证)
    ├── Hooks (生命周期钩子)
    └── Transformer (数据转换)
    ↓
Repository (数据访问层)
    ├── SQLBuilder (SQL 构建器)
    └── store/db (数据库抽象)
    ↓
Database (PostgreSQL/MySQL/SQLite)
```

### 核心组件

1. **Repository[T]** - 泛型数据仓库接口
2. **BaseRepository[T]** - 通用数据库操作实现
3. **SQLBuilder** - 动态 SQL 查询构建器
4. **DBResourceController[T]** - 数据库集成的控制器
5. **ResourceHooks** - 生命周期钩子
6. **ResourceTransformer** - 资源转换器

## 快速开始

### 1. 定义数据模型

```go
type UserModel struct {
    ID        string    `json:"id" db:"id"`
    Name      string    `json:"name" db:"name"`
    Email     string    `json:"email" db:"email"`
    Status    string    `json:"status" db:"status"`
    CreatedAt time.Time `json:"created_at" db:"created_at"`
    UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}
```

### 2. 创建 Repository

```go
func NewUserRepository(database db.DB) *router.BaseRepository[UserModel] {
    builder := router.NewSQLBuilder("users", "id").
        WithColumns("id", "name", "email", "status", "created_at", "updated_at").
        WithScanFunc(scanUser).
        WithInsertFunc(userInsertValues).
        WithUpdateFunc(userUpdateValues)

    return router.NewBaseRepository[UserModel](database, builder)
}

func scanUser(rows *sql.Rows) (any, error) {
    var user UserModel
    err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.Status, &user.CreatedAt, &user.UpdatedAt)
    if err != nil {
        return nil, err
    }
    return &user, nil
}

func userInsertValues(data any) ([]any, error) {
    user := data.(*UserModel)
    return []any{user.ID, user.Name, user.Email, user.Status, user.CreatedAt, user.UpdatedAt}, nil
}

func userUpdateValues(data any) ([]any, error) {
    user := data.(*UserModel)
    return []any{user.Name, user.Email, user.Status, user.UpdatedAt}, nil
}
```

### 3. 创建控制器

```go
func NewUserController(database db.DB) *router.DBResourceController[UserModel] {
    repo := NewUserRepository(database)
    ctrl := router.NewDBResourceController[UserModel]("user", repo)

    // 配置查询选项
    ctrl.QueryBuilder.
        WithPageSize(20, 100).
        WithAllowedSorts("name", "email", "created_at").
        WithAllowedFilters("status")

    // 配置钩子和转换器
    ctrl.Hooks = &UserHooks{}
    ctrl.Transformer = &UserTransformer{}

    return ctrl
}
```

### 4. 注册路由

```go
func RegisterUserRoutes(r *router.Router, database db.DB) {
    ctrl := NewUserController(database)

    r.Get("/users", ctrl.IndexCtx)        // 列表查询
    r.Get("/users/:id", ctrl.ShowCtx)     // 获取单条
    r.Post("/users", ctrl.CreateCtx)      // 创建
    r.Put("/users/:id", ctrl.UpdateCtx)   // 更新
    r.Delete("/users/:id", ctrl.DeleteCtx) // 删除
}
```

## 组件详解

### Repository 接口

Repository 提供了标准的 CRUD 操作接口:

```go
type Repository[T any] interface {
    // 查询所有记录(支持分页、排序、过滤)
    FindAll(ctx context.Context, params *QueryParams) ([]T, int64, error)

    // 根据 ID 查询单条记录
    FindByID(ctx context.Context, id string) (*T, error)

    // 创建新记录
    Create(ctx context.Context, data *T) error

    // 更新记录
    Update(ctx context.Context, id string, data *T) error

    // 删除记录
    Delete(ctx context.Context, id string) error

    // 统计记录数
    Count(ctx context.Context, params *QueryParams) (int64, error)

    // 检查记录是否存在
    Exists(ctx context.Context, id string) (bool, error)
}
```

### SQLBuilder 查询构建器

SQLBuilder 根据 QueryParams 动态构建 SQL 查询:

```go
builder := router.NewSQLBuilder("users", "id").
    WithColumns("id", "name", "email", "created_at").
    WithScanFunc(scanUser).
    WithInsertFunc(userInsertValues).
    WithUpdateFunc(userUpdateValues)

// 构建 SELECT 查询
query, args := builder.BuildSelectQuery(params)
// 示例: SELECT id, name, email FROM users WHERE status = ? ORDER BY created_at DESC LIMIT 20 OFFSET 0

// 构建 COUNT 查询
query, args := builder.BuildCountQuery(params)
// 示例: SELECT COUNT(*) FROM users WHERE status = ?

// 构建 INSERT 查询
query := builder.BuildInsertQuery()
// 示例: INSERT INTO users (id, name, email, created_at) VALUES (?, ?, ?, ?)

// 构建 UPDATE 查询
query := builder.BuildUpdateQuery()
// 示例: UPDATE users SET name = ?, email = ?, created_at = ? WHERE id = ?

// 构建 DELETE 查询
query := builder.BuildDeleteQuery()
// 示例: DELETE FROM users WHERE id = ?
```

### QueryParams 支持的查询参数

```go
type QueryParams struct {
    // 分页
    Page     int
    PageSize int
    Offset   int
    Limit    int

    // 排序 (支持多字段)
    Sort      []SortField  // [{Field: "name", Desc: false}, {Field: "created_at", Desc: true}]

    // 过滤
    Filters map[string]string  // {"status": "active", "role": "admin"}
    Search  string              // 全文搜索

    // 字段选择
    Fields []string  // ["id", "name", "email"]

    // 关联资源
    Include []string  // ["posts", "comments"]
}
```

### ResourceHooks 生命周期钩子

在数据库操作前后执行自定义逻辑:

```go
type UserHooks struct {
    router.NoOpResourceHooks
}

func (h *UserHooks) BeforeCreate(ctx context.Context, data any) error {
    user := data.(*UserModel)

    // 设置时间戳
    now := time.Now()
    user.CreatedAt = now
    user.UpdatedAt = now

    // 生成 ID
    if user.ID == "" {
        user.ID = generateID()
    }

    // 业务规则验证
    if user.Email == "admin@example.com" && user.Role != "admin" {
        return fmt.Errorf("this email is reserved for admin users")
    }

    return nil
}

func (h *UserHooks) AfterCreate(ctx context.Context, data any) error {
    user := data.(*UserModel)
    // 发送欢迎邮件
    sendWelcomeEmail(user.Email)
    // 发布事件
    publishUserCreatedEvent(user)
    return nil
}
```

### ResourceTransformer 数据转换

转换数据库模型为 API 响应格式:

```go
type UserTransformer struct{}

func (t *UserTransformer) Transform(ctx context.Context, resource any) (any, error) {
    user := resource.(*UserModel)
    return map[string]any{
        "id":         user.ID,
        "name":       user.Name,
        "email":      maskEmail(user.Email),  // 隐藏部分邮箱
        "status":     user.Status,
        "created_at": user.Created.Format(time.RFC3339),
    }, nil
}

func (t *UserTransformer) TransformCollection(ctx context.Context, resources any) (any, error) {
    users := resources.([]UserModel)
    result := make([]map[string]any, len(users))
    for i, user := range users {
        transformed, _ := t.Transform(ctx, &user)
        result[i] = transformed.(map[string]any)
    }
    return result, nil
}
```

## 完整示例

### 用户管理 API

#### 1. 数据库表结构

```sql
CREATE TABLE users (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    email TEXT UNIQUE NOT NULL,
    status TEXT NOT NULL DEFAULT 'active',
    role TEXT NOT NULL DEFAULT 'user',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_status ON users(status);
CREATE INDEX idx_users_created_at ON users(created_at);
```

#### 2. API 请求示例

**列表查询(分页+排序+过滤)**
```bash
GET /users?page=1&page_size=20&sort=-created_at&status=active
```

**搜索用户**
```bash
GET /users?search=john&fields=id,name,email
```

**获取单个用户**
```bash
GET /users/user-123
```

**创建用户**
```bash
POST /users
Content-Type: application/json

{
  "name": "Alice Smith",
  "email": "alice@example.com",
  "status": "active",
  "role": "user"
}
```

**更新用户**
```bash
PUT /users/user-123
Content-Type: application/json

{
  "name": "Alice Johnson",
  "status": "inactive"
}
```

**删除用户**
```bash
DELETE /users/user-123
```

#### 3. 响应格式

**成功响应(列表)**
```json
{
  "data": [
    {
      "id": "user-123",
      "name": "Alice",
      "email": "a****e@example.com",
      "status": "active",
      "created_at": "2024-01-01T00:00:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "page_size": 20,
    "total_items": 100,
    "total_pages": 5,
    "has_next": true,
    "has_prev": false
  }
}
```

**错误响应**
```json
{
  "status": 404,
  "code": "NOT_FOUND",
  "message": "Record not found",
  "category": "client"
}
```

## 高级特性

### 软删除支持

```go
type ProductModel struct {
    router.BaseModel
    Name      string
    DeletedAt *time.Time `db:"deleted_at"`
}

// 软删除实现
func (r *ProductRepository) SoftDelete(ctx context.Context, id string) error {
    now := time.Now()
    query := "UPDATE products SET deleted_at = ? WHERE id = ? AND deleted_at IS NULL"
    result, err := db.ExecContext(ctx, r.db, query, now, id)
    if err != nil {
        return err
    }

    rowsAffected, _ := result.RowsAffected()
    if rowsAffected == 0 {
        return sql.ErrNoRows
    }

    return nil
}

// 只查询未删除的记录
func (r *ProductRepository) FindAllActive(ctx context.Context, params *QueryParams) ([]ProductModel, int64, error) {
    // 添加 deleted_at IS NULL 条件
    // 实际实现需要扩展 SQLBuilder
    return r.FindAll(ctx, params)
}
```

### 自定义查询方法

```go
// 根据邮箱查找用户
func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*UserModel, error) {
    query := "SELECT id, name, email, status FROM users WHERE email = ?"
    rows, err := db.QueryContext(ctx, r.db, query, email)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    if !rows.Next() {
        return nil, sql.ErrNoRows
    }

    return scanUser(rows).(*UserModel), nil
}

// 查找所有活跃用户
func (r *UserRepository) FindActiveUsers(ctx context.Context, params *QueryParams) ([]UserModel, int64, error) {
    if params.Filters == nil {
        params.Filters = make(map[string]string)
    }
    params.Filters["status"] = "active"
    return r.FindAll(ctx, params)
}
```

### 事务支持

```go
import "github.com/spcent/plumego/store/db"

func (c *UserController) TransferOwnership(ctx *contract.Ctx) {
    var req TransferRequest
    ctx.BindJSON(&req)

    err := db.WithTransaction(ctx.R.Context(), c.database, nil, func(tx *sql.Tx) error {
        // 在事务中执行多个操作
        if err := c.repo.Update(ctx.R.Context(), req.FromUserID, &fromUser); err != nil {
            return err
        }

        if err := c.repo.Update(ctx.R.Context(), req.ToUserID, &toUser); err != nil {
            return err
        }

        return nil
    })

    if err != nil {
        ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Transfer failed"})
        return
    }

    ctx.JSON(http.StatusOK, map[string]string{"status": "success"})
}
```

## 最佳实践

### 1. 分离关注点

```go
// ✅ 推荐: 清晰的分层架构
Model      →  数据库模型
Repository →  数据访问逻辑
Controller →  HTTP 请求处理
Hooks      →  业务规则
Transformer → 数据格式化

// ❌ 避免: 将所有逻辑混在控制器中
```

### 2. 使用 Context 传递请求信息

```go
func (r *UserRepository) FindAll(ctx context.Context, params *QueryParams) ([]UserModel, int64, error) {
    // 从 context 中获取请求 ID 用于日志追踪
    requestID := ctx.Value("request_id")

    // 执行查询...
}
```

### 3. 验证在控制器层

```go
func (c *DBResourceController[T]) CreateCtx(ctx *contract.Ctx) {
    var data T
    ctx.BindJSON(&data)

    // 验证在这里进行
    if c.validator != nil {
        if err := c.validator.Validate(data); err != nil {
            ctx.JSON(http.StatusUnprocessableEntity, err)
            return
        }
    }

    // Repository 只负责数据访问
    c.repository.Create(ctx.R.Context(), &data)
}
```

### 4. 使用 Hooks 处理副作用

```go
// ✅ 推荐: 在 Hooks 中处理副作用
func (h *UserHooks) AfterCreate(ctx context.Context, data any) error {
    user := data.(*UserModel)
    // 发送邮件、缓存失效、事件发布等
    sendEmail(user.Email, "Welcome!")
    cache.Delete("users:list")
    events.Publish("user.created", user)
    return nil
}

// ❌ 避免: 在 Controller 中处理副作用
```

### 5. 合理使用 Transformer

```go
// ✅ 推荐: 使用 Transformer 隐藏敏感信息
func (t *UserTransformer) Transform(ctx context.Context, resource any) (any, error) {
    user := resource.(*UserModel)
    return map[string]any{
        "id":    user.ID,
        "name":  user.Name,
        "email": maskEmail(user.Email),  // 隐藏部分邮箱
        // 不返回密码、内部标识等敏感信息
    }, nil
}
```

## API 参考

### Repository 方法

| 方法 | 描述 | 示例 |
|------|------|------|
| `FindAll(ctx, params)` | 查询列表 | `repo.FindAll(ctx, &QueryParams{Page: 1, PageSize: 20})` |
| `FindByID(ctx, id)` | 查询单条 | `repo.FindByID(ctx, "user-123")` |
| `Create(ctx, data)` | 创建 | `repo.Create(ctx, &user)` |
| `Update(ctx, id, data)` | 更新 | `repo.Update(ctx, "user-123", &user)` |
| `Delete(ctx, id)` | 删除 | `repo.Delete(ctx, "user-123")` |
| `Count(ctx, params)` | 统计 | `repo.Count(ctx, &QueryParams{Filters: filters})` |
| `Exists(ctx, id)` | 检查存在 | `repo.Exists(ctx, "user-123")` |

### QueryBuilder 配置

```go
builder := router.NewSQLBuilder("users", "id")

// 配置列
builder.WithColumns("id", "name", "email", "created_at")

// 配置扫描函数
builder.WithScanFunc(func(rows *sql.Rows) (any, error) {
    var user UserModel
    err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.CreatedAt)
    return &user, err
})

// 配置插入函数
builder.WithInsertFunc(func(data any) ([]any, error) {
    user := data.(*UserModel)
    return []any{user.ID, user.Name, user.Email, user.CreatedAt}, nil
})

// 配置更新函数
builder.WithUpdateFunc(func(data any) ([]any, error) {
    user := data.(*UserModel)
    return []any{user.Name, user.Email, user.CreatedAt}, nil
})
```

### DBResourceController 配置

```go
ctrl := router.NewDBResourceController[UserModel]("user", repo)

// 配置查询选项
ctrl.QueryBuilder.
    WithPageSize(20, 100).                        // 默认 20,最大 100
    WithAllowedSorts("name", "email", "created_at").  // 允许排序的字段
    WithAllowedFilters("status", "role")          // 允许过滤的字段

// 配置验证器
ctrl.WithValidator(myValidator)

// 配置钩子
ctrl.Hooks = &MyHooks{}

// 配置转换器
ctrl.Transformer = &MyTransformer{}
```

## 学习资源

- [resource.go](router/resource.go) - CRUD 框架核心
- [resource_db.go](router/resource_db.go) - 数据库集成
- [resource_example.go](router/resource_example.go) - 基础示例
- [resource_db_example.go](router/resource_db_example.go) - 数据库示例
- [store/db 包文档](../store/db) - 数据库抽象层

## 总结

这个 CRUD 框架提供了:

✅ **类型安全** - 使用 Go 泛型确保类型安全
✅ **高度可复用** - Repository 和 Controller 可用于任何模型
✅ **生命周期钩子** - 在关键点注入自定义逻辑
✅ **查询灵活性** - 支持分页、排序、过滤、搜索
✅ **标准化响应** - 统一的 API 响应格式
✅ **事务支持** - 集成 store/db 的事务功能
✅ **软删除** - 内置软删除模式支持
✅ **测试友好** - 所有组件易于 mock 和测试

现在你可以用几行代码就构建一个完整的、生产级别的 RESTful CRUD API!
