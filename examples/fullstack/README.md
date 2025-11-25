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

项目采用清晰的分层架构和接口设计：

```
examples/fullstack/
├── main.go              # 主入口（依赖注入和启动）
├── service/             # 服务接口层
│   └── interface.go     # 业务服务接口定义
├── model/               # 数据模型层
│   └── model.go         # 数据模型、请求/响应结构定义
├── store/               # 数据存储层
│   └── store.go         # 模拟数据库操作
├── serviceimpl/         # 业务逻辑实现层
│   └── impl.go          # 纯业务逻辑实现（不依赖HTTP）
├── handler/             # HTTP适配器层（Server端）
│   └── handler.go       # 将HTTP请求转发到业务服务层
├── apiclient/           # HTTP客户端层（Client端）
│   ├── client.go        # API客户端（实现 service 接口）
│   └── demo.go          # Client端调用示例
├── *_test.go            # 测试文件
│   ├── integration_test.go    # 集成测试（完整HTTP流程）
│   └── service_test.go        # 服务接口测试（RPC vs 直接调用）
├── go.mod
└── README.md
```

### 分层说明

- **service** - 业务服务接口定义（UserService、ArticleService 等）
- **model** - 数据模型和DTO（Data Transfer Object）
- **store** - 数据持久化层（本例中为内存模拟）
- **serviceimpl** - 纯业务逻辑实现（不依赖HTTP框架）
- **handler** - HTTP适配器层，将HTTP请求转换为服务调用
- **apiclient** - HTTP客户端，实现 service 接口通过RPC调用
- **main** - 程序入口，依赖注入和启动

### 接口设计亮点

**三层实现相同的接口：**

```go
// service/interface.go - 业务接口定义
type UserService interface {
    CreateUser(ctx context.Context, req model.CreateUserRequest) (model.User, error)
    GetUser(ctx context.Context, req model.GetUserRequest) (model.User, error)
    ListUsers(ctx context.Context, req model.ListUsersRequest) (model.ListUsersResponse, error)
    DeleteUser(ctx context.Context, req model.DeleteUserRequest) error
}

// serviceimpl/impl.go - 纯业务逻辑实现
type ServiceImpl struct { store *store.Store }
var _ service.UserService = (*ServiceImpl)(nil)  // 编译时检查

// handler/handler.go - HTTP适配器（Server端）
type Handler struct { svc service.Service }
var _ service.UserService = (*Handler)(nil)  // 编译时检查

// apiclient/client.go - HTTP客户端（Client端）
type Client struct { restyClient *resty.Client }
var _ service.UserService = (*Client)(nil)   // 编译时检查
```

**依赖关系：**
```
Client (RPC) ──HTTP──> Handler (HTTP适配器) ──直接调用──> ServiceImpl (业务逻辑) ──> Store (数据)
```

**优势：**
- ✅ 业务逻辑与HTTP框架解耦（ServiceImpl纯净，不依赖Gin）
- ✅ Handler只负责HTTP协议适配（薄薄一层）
- ✅ Client和ServiceImpl可互换使用（都实现Service接口）
- ✅ 便于单元测试（可直接测试ServiceImpl）
- ✅ 便于集成测试（测试完整HTTP流程）
- ✅ 编译时类型检查和接口一致性保证

## 测试

项目包含完整的测试套件：

### 1. 集成测试 (integration_test.go)

测试完整的HTTP请求-响应流程：

```bash
go test -v -run TestIntegration
```

- 启动HTTP测试服务器
- 通过Client发送真实的HTTP请求
- 验证完整的请求处理流程（Client → HTTP → Handler → ServiceImpl → Store）
- 包含错误处理测试

### 2. 服务接口测试 (service_test.go)

测试Service接口的不同实现方式：

```bash
go test -v -run TestServiceInterface
```

**测试两种实现：**
- **DirectCall** - 直接函数调用（ServiceImpl）
- **RPCCall** - RPC远程调用（Client → HTTP → Handler → ServiceImpl）

**验证：**
- ✅ 两种实现产生相同的结果
- ✅ 接口一致性
- ✅ ServiceImpl可以独立于HTTP使用
- ✅ 通过RPC调用与直接调用行为一致

### 运行所有测试

```bash
cd examples/fullstack
go test -v
```

## 注意事项

### 组合参数的binding标签

当请求结构体包含来自多个源的参数（URI + JSON，或 URI + Query）时，**避免使用`binding`验证标签**：

```go
// ❌ 错误 - 会导致验证失败
type UpdateArticleRequest struct {
    ID      int64  `uri:"id" binding:"required"`      // URI参数
    Title   string `json:"title" binding:"required"`  // JSON参数
}

// ✅ 正确 - 组合参数不使用binding标签
type UpdateArticleRequest struct {
    ID      int64  `uri:"id" path:"id"`    // URI参数
    Title   string `json:"title"`          // JSON参数
}
```

**原因：** Gin在绑定多个源时会对整个结构体多次验证，导致在绑定JSON之前验证失败。

**对于单一源的参数，可以安全使用binding标签：**

```go
// ✅ 单一源（仅JSON）- 可以使用binding标签
type CreateUserRequest struct {
    Name  string `json:"name" binding:"required"`
    Email string `json:"email" binding:"required,email"`
}
```
