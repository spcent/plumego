# CRUD Framework Demo

这是一个完整的、可直接运行的示例应用，展示如何使用 Plumego 的数据库集成 CRUD 框架构建 RESTful API。

## 🎯 功能展示

- ✅ **Repository 模式** - 类型安全的泛型数据访问层
- ✅ **自动 SQL 生成** - 根据 QueryParams 动态构建 SQL
- ✅ **分页支持** - 自动计算总页数、上一页、下一页
- ✅ **多字段排序** - 支持升序/降序排序
- ✅ **灵活过滤** - 按字段过滤数据
- ✅ **全文搜索** - 跨多个字段搜索
- ✅ **生命周期钩子** - Before/After 操作钩子
- ✅ **数据转换** - 自定义 API 响应格式
- ✅ **数据库集成** - 完整的 store/db 集成

## 🚀 快速开始

### 1. 运行示例

```bash
# 从项目根目录运行
cd examples/crud-demo
go run main.go
```

服务器将在 `http://localhost:8080` 启动，并自动创建内存数据库和示例数据。

### 2. 访问主页

打开浏览器访问：http://localhost:8080

你会看到一个漂亮的文档页面，列出所有可用的端点和示例。

## 📚 API 端点

### 列表查询

```bash
# 获取所有用户
curl http://localhost:8080/users

# 分页查询
curl "http://localhost:8080/users?page=1&page_size=2"

# 按创建时间降序排序
curl "http://localhost:8080/users?sort=-created_at"

# 按状态过滤
curl "http://localhost:8080/users?status=active"

# 搜索用户
curl "http://localhost:8080/users?search=alice"

# 组合查询
curl "http://localhost:8080/users?page=1&page_size=3&sort=-created_at&status=active"
```

**响应示例：**
```json
{
  "data": [
    {
      "id": "user-1",
      "name": "Alice Johnson",
      "email": "alice@example.com",
      "status": "active",
      "role": "admin",
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "page_size": 10,
    "total_items": 5,
    "total_pages": 1,
    "has_next": false,
    "has_prev": false
  }
}
```

### 获取单个用户

```bash
curl http://localhost:8080/users/user-1
```

**响应示例：**
```json
{
  "id": "user-1",
  "name": "Alice Johnson",
  "email": "alice@example.com",
  "status": "active",
  "role": "admin",
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```

### 创建用户

```bash
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{
    "name": "New User",
    "email": "newuser@example.com",
    "role": "user"
  }'
```

**响应示例：**
```json
{
  "id": "user-1234567890",
  "name": "New User",
  "email": "newuser@example.com",
  "status": "active",
  "role": "user",
  "created_at": "2024-01-01T12:00:00Z",
  "updated_at": "2024-01-01T12:00:00Z"
}
```

### 更新用户

```bash
curl -X PUT http://localhost:8080/users/user-1 \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Alice Updated",
    "status": "inactive"
  }'
```

### 删除用户

```bash
curl -X DELETE http://localhost:8080/users/user-1
```

**响应：** HTTP 204 No Content

## 🎨 代码结构

```
main.go
├── 数据模型定义 (UserModel)
├── Repository 实现 (UserRepository)
│   ├── SQLBuilder 配置
│   ├── Scan 函数
│   ├── Insert 函数
│   └── Update 函数
├── 生命周期钩子 (UserHooks)
│   ├── BeforeCreate - 设置时间戳、生成ID
│   ├── AfterCreate - 记录日志
│   ├── BeforeUpdate - 更新时间戳
│   └── AfterDelete - 清理操作
├── 数据转换器 (UserTransformer)
│   ├── Transform - 单个资源转换
│   └── TransformCollection - 集合转换
├── 控制器创建 (NewUserController)
│   ├── Repository 初始化
│   ├── QueryBuilder 配置
│   ├── Hooks 配置
│   └── Transformer 配置
└── 路由注册
```

## 📖 关键特性说明

### 1. Repository 模式

```go
type UserRepository struct {
    *router.BaseRepository[UserModel]
}
```

- 类型安全的泛型接口
- 标准 CRUD 操作
- 自动分页和计数

### 2. SQL 查询构建

```go
builder := router.NewSQLBuilder("users", "id").
    WithColumns("id", "name", "email", "status", "role", "created_at", "updated_at").
    WithScanFunc(scanUser).
    WithInsertFunc(userInsertValues).
    WithUpdateFunc(userUpdateValues)
```

自动生成的 SQL：
- `SELECT id, name, email FROM users WHERE status = ? ORDER BY created_at DESC LIMIT 10 OFFSET 0`
- `INSERT INTO users (id, name, email) VALUES (?, ?, ?)`
- `UPDATE users SET name = ?, email = ? WHERE id = ?`

### 3. 生命周期钩子

```go
func (h *UserHooks) BeforeCreate(ctx context.Context, data any) error {
    user := data.(*UserModel)
    // 设置时间戳、生成ID、验证规则
    user.CreatedAt = time.Now()
    user.ID = generateID()
    return nil
}
```

### 4. 查询参数配置

```go
ctrl.QueryBuilder.
    WithPageSize(10, 50).                                    // 默认10, 最大50
    WithAllowedSorts("name", "email", "created_at").        // 允许排序的字段
    WithAllowedFilters("status", "role")                    // 允许过滤的字段
```

## 🧪 测试查询

### 分页测试

```bash
# 第一页，每页2条
curl "http://localhost:8080/users?page=1&page_size=2"

# 第二页
curl "http://localhost:8080/users?page=2&page_size=2"
```

### 排序测试

```bash
# 按名称升序
curl "http://localhost:8080/users?sort=name"

# 按创建时间降序
curl "http://localhost:8080/users?sort=-created_at"

# 多字段排序
curl "http://localhost:8080/users?sort=status,-created_at"
```

### 过滤测试

```bash
# 只显示活跃用户
curl "http://localhost:8080/users?status=active"

# 只显示管理员
curl "http://localhost:8080/users?role=admin"

# 组合过滤
curl "http://localhost:8080/users?status=active&role=user"
```

### 搜索测试

```bash
# 搜索包含 "alice" 的用户
curl "http://localhost:8080/users?search=alice"

# 搜索邮箱
curl "http://localhost:8080/users?search=example.com"
```

## 📊 示例数据

应用启动时会自动创建5个示例用户：

| ID | Name | Email | Role | Status |
|----|------|-------|------|--------|
| user-1 | Alice Johnson | alice@example.com | admin | active |
| user-2 | Bob Smith | bob@example.com | user | active |
| user-3 | Charlie Brown | charlie@example.com | user | active |
| user-4 | Diana Prince | diana@example.com | user | active |
| user-5 | Eve Davis | eve@example.com | user | active |

## 🎓 学习资源

- [CRUD_FRAMEWORK.md](../../router/CRUD_FRAMEWORK.md) - 完整文档
- [resource.go](../../router/resource.go) - 基础 CRUD 框架
- [resource_db.go](../../router/resource_db.go) - 数据库集成
- [resource_example.go](../../router/resource_example.go) - 基础示例
- [resource_db_example.go](../../router/resource_db_example.go) - 数据库示例

## 💡 扩展建议

### 添加更多端点

```go
// 批量创建
r.Post("/users/batch", contract.AdaptCtxHandler(ctrl.BatchCreateCtx, r.Logger()))

// 批量删除
r.Delete("/users/batch", contract.AdaptCtxHandler(ctrl.BatchDeleteCtx, r.Logger()))

// 自定义端点
r.Get("/users/active", contract.AdaptCtxHandler(func(ctx *contract.Ctx) {
    // 使用 repository 的自定义方法
}, r.Logger()))
```

### 添加验证

```go
import "github.com/spcent/plumego/validator"

// 在控制器中添加验证器
ctrl.WithValidator(validator.NewValidator(nil))
```

### 添加更多 Hooks

```go
func (h *UserHooks) BeforeUpdate(ctx context.Context, id string, data any) error {
    // 检查权限
    // 验证业务规则
    // 等等
    return nil
}
```

## 🎉 总结

这个示例展示了如何用不到 300 行代码构建一个完整的、生产级别的 RESTful CRUD API！

**核心优势：**
- ✅ 类型安全
- ✅ 高度可复用
- ✅ 最小样板代码
- ✅ 强大的查询能力
- ✅ 易于扩展
- ✅ 生产就绪

开始构建你自己的 API 吧！ 🚀
