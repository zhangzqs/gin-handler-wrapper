# 完整示例 - Gin Handler Wrapper

这是一个完整的示例应用，展示了如何使用 `gin-handler-wrapper` 的所有功能。

## 功能展示

### 1. WrapHandler - 完整的输入输出

适用于需要请求参数和返回响应的场景：

- `POST /users` - 创建用户
- `GET /users/:id` - 获取用户详情
- `GET /users?page=1&page_size=10` - 获取用户列表（分页）
- `GET /articles/search?keyword=Go` - 搜索文章

### 2. WrapConsumer - 只有输入

适用于只需要请求参数、不需要返回数据的场景：

- `PUT /users/:id` - 更新用户
- `DELETE /users/:id` - 删除用户

### 3. WrapGetter - 只有输出

适用于不需要请求参数、只返回数据的场景：

- `GET /health` - 健康检查

### 4. WrapAction - 无输入输出

适用于触发操作、不需要参数和返回值的场景：

- `POST /cache/clear` - 清除缓存
- `POST /data/sync` - 同步数据

## 运行示例

```bash
# 进入示例目录
cd examples/

# 运行服务
go run main.go
```

服务将在 `http://localhost:8080` 启动。

## 测试 API

### 1. 创建用户

```bash
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{"name":"Alice","email":"alice@example.com"}'
```

响应：

```json
{
  "id": 3,
  "name": "Alice",
  "email": "alice@example.com",
  "created_at": "2025-01-15T10:30:00Z"
}
```

### 2. 获取用户

```bash
curl http://localhost:8080/users/1
```

响应：

```json
{
  "id": 1,
  "name": "Alice",
  "email": "alice@example.com",
  "created_at": "2025-01-15T10:00:00Z"
}
```

### 3. 获取用户列表

```bash
curl "http://localhost:8080/users?page=1&page_size=10"
```

响应：

```json
{
  "items": [
    {
      "id": 1,
      "name": "Alice",
      "email": "alice@example.com",
      "created_at": "2025-01-15T10:00:00Z"
    },
    {
      "id": 2,
      "name": "Bob",
      "email": "bob@example.com",
      "created_at": "2025-01-15T10:05:00Z"
    }
  ],
  "total": 2,
  "page": 1,
  "page_size": 10,
  "total_pages": 1
}
```

### 4. 更新用户

```bash
curl -X PUT http://localhost:8080/users/1 \
  -H "Content-Type: application/json" \
  -d '{"name":"Alice Updated"}'
```

### 5. 删除用户

```bash
curl -X DELETE http://localhost:8080/users/1
```

### 6. 健康检查

```bash
curl http://localhost:8080/health
```

响应：

```json
{
  "status": "ok",
  "time": "2025-01-15T10:30:00Z",
  "version": "1.0.0"
}
```

### 7. 搜索文章

```bash
curl "http://localhost:8080/articles/search?keyword=Go&page=1&page_size=10"
```

响应：

```json
{
  "items": [
    {
      "id": 1,
      "title": "Go 语言入门",
      "content": "这是一篇关于 Go 的文章",
      "author": "Alice"
    }
  ],
  "total": 1,
  "page": 1,
  "page_size": 10,
  "total_pages": 1
}
```

### 8. 清除缓存

```bash
curl -X POST http://localhost:8080/cache/clear
```

### 9. 同步数据

```bash
curl -X POST http://localhost:8080/data/sync
```

## 特性展示

### 自动参数绑定

示例展示了多种参数绑定方式：

- **URI 参数**: `uri:"id"` - 从路径中提取
- **Query 参数**: `form:"page"` - 从查询字符串提取
- **JSON Body**: `json:"name"` - 从请求体提取
- **组合绑定**: 可以同时使用多种绑定方式

### 参数验证

使用 Gin 的 `binding` 标签进行参数验证：

- `required` - 必填
- `email` - 邮箱格式
- `gt=0` - 大于 0
- `gte=1` - 大于等于 1
- `lte=100` - 小于等于 100

### 自定义错误处理

示例实现了 `customErrorHandler`，根据不同的错误类型返回不同的 HTTP 状态码：

- `ErrUserNotFound` → 404
- `ErrUserAlreadyExists` → 409
- `ErrInvalidID` → 400
- 其他错误 → 500

### 泛型支持

示例展示了泛型的使用：

- 指针类型: `*User`, `*HealthResponse`
- 泛型结构体: `ListResponse[User]`, `ListResponse[Article]`

## 代码结构

```
main.go
├── 数据模型 (User, Article)
├── 请求/响应类型 (各种 Request/Response)
├── 业务错误定义 (ErrUserNotFound, etc.)
├── 模拟数据库 (内存存储)
├── 业务处理器 (CreateUser, GetUser, etc.)
├── 自定义错误处理器 (customErrorHandler)
└── 路由配置 (main 函数)
```

## 最佳实践

1. **分离关注点**: 请求/响应类型、业务逻辑、错误处理分离
2. **类型安全**: 使用泛型确保类型安全
3. **参数验证**: 使用 `binding` 标签进行声明式验证
4. **错误处理**: 自定义错误处理器，返回合适的状态码
5. **代码复用**: 使用不同的包装函数处理不同场景
