# Fullstack 示例

这个示例展示了如何构建一个完整的全栈应用，同时使用 server 和 client 包实现 HTTP 服务端和客户端的无缝交互。

## Server 端功能

展示了 server 包的所有功能：

1. **WrapHandler** - 完整的请求/响应处理
   - 创建用户（JSON body）
   - 获取用户（URI 参数）
   - 获取用户列表（Query 参数）
   - 更新文章（组合参数：URI + JSON）

2. **WrapGetter** - 只有响应
   - 健康检查

3. **WrapConsumer** - 只有请求
   - 删除用户

4. **WrapAction** - 无输入输出
   - 触发任务

5. **自定义错误处理**
   - 演示自定义 ErrorHandler

## Client 端功能

展示了 client 包的多种绑定方式：

1. **路径参数绑定** (`path` 标签)
   ```go
   type GetUserRequest struct {
       ID int64 `path:"id"`
   }
   ```

2. **Query 参数绑定** (`query` 标签)
   ```go
   type ListUsersRequest struct {
       Page     int `query:"page"`
       PageSize int `query:"page_size"`
   }
   ```

3. **请求头绑定** (`header` 标签)
   ```go
   type AuthRequest struct {
       Token string `header:"Authorization"`
   }
   ```

4. **JSON Body 绑定** (`json` 标签)
   ```go
   type CreateUserRequest struct {
       Name  string `json:"name"`
       Email string `json:"email"`
   }
   ```

5. **组合绑定** - 同时使用多种标签
   ```go
   type UpdateArticleRequest struct {
       ID      int64  `path:"id"`              // 路径参数
       Token   string `header:"Authorization"` // 请求头
       Title   string `json:"title"`           // JSON body
       Content string `json:"content"`         // JSON body
   }
   ```

## 运行示例

```bash
cd examples/fullstack
go mod tidy
go run main.go
```

程序会：
1. 启动 HTTP 服务器（端口 8080）
2. 自动运行客户端示例，演示所有功能
3. 输出每个操作的结果

## 输出示例

```
Server starting on http://localhost:8080

========== Client端调用示例 ==========

1. 健康检查
Health: {Status:ok Timestamp:2024-11-25 00:00:00}

2. 创建用户
Created user: {ID:3 Name:Charlie Email:charlie@example.com CreatedAt:2024-11-25 00:00:00}

3. 获取用户
Got user: {ID:1 Name:Alice Email:alice@example.com CreatedAt:2024-11-25 00:00:00}

4. 获取用户列表
User list: {Total:3 Users:[...]}

5. 更新文章
Updated article: {ID:1 Title:New Article Content:This is the content Author:unknown}

6. 删除用户
User deleted successfully

7. 触发任务
Task triggered successfully
```

## 学习要点

### Server 端

- 使用 `server.WrapHandler` 及其变体简化处理器编写
- 支持 URI、Query、JSON、Form 等多种参数绑定
- 可以自定义 Decoder、Encoder 和 ErrorHandler
- 完全支持泛型，类型安全

### Client 端

- 使用 `client.NewClient` 及其变体创建 HTTP 客户端
- 通过结构体标签灵活控制请求参数位置
- 支持路径参数、Query 参数、请求头、请求体的任意组合
- 自动序列化请求和反序列化响应
- 完全支持泛型，类型安全

## 代码结构

```
main.go
├── 数据模型
│   ├── User
│   └── Article
├── Server 端
│   ├── 请求/响应类型
│   ├── 业务逻辑函数
│   ├── 路由设置
│   └── 自定义错误处理
└── Client 端
    ├── 请求类型（带标签）
    └── 调用示例
```
