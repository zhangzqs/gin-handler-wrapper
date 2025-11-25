# Gin Handler Wrapper

一个类型安全的 Gin 框架扩展库，包含 **Server 端处理器包装**（gin-server）和 **Client 端请求构建**（resty-client）两大功能，使用 Go 泛型提供优雅的请求/响应处理。

## 项目结构

- **gin-server**: Gin 服务端请求包装功能（包名：`ginserver`）
- **resty-client**: Resty 客户端请求处理功能（包名：`restyclient`，基于 `resty.dev/v3`）
- **handler**: 通用处理函数类型定义
- **examples/fullstack**: 完整的服务端/客户端交互示例

## 特性

### Server 端
- ✅ **类型安全**：使用 Go 泛型实现编译时类型检查
- ✅ **自动绑定**：支持 URI、Query、JSON、Form 等多种数据源
- ✅ **灵活定制**：可自定义解码器、编码器和错误处理器
- ✅ **便捷函数**：提供多种模板函数覆盖常见场景
- ✅ **清晰架构**：职责分离，代码易于维护

### Client 端
- ✅ **类型安全**：完全类型安全的 HTTP 客户端
- ✅ **智能绑定**：通过标签自动处理路径参数、Query 参数、请求头和请求体
- ✅ **灵活定制**：可自定义编码器、解码器和错误处理器
- ✅ **便捷函数**：提供多种包装函数简化常见场景
- ✅ **基于 Resty**：构建在成熟的 go-resty 库之上

## 安装

```bash
go get github.com/zhangzqs/go-typed-rpc
```

## 快速开始

> 💡 **完整示例**: 查看 [examples/fullstack](./examples/fullstack) 目录获取包含 Server 和 Client 完整交互的可运行示例。

### Server 端基础用法

```go
package main

import (
    "context"
    "github.com/gin-gonic/gin"
    ginserver "github.com/zhangzqs/go-typed-rpc/gin-server"
)

type CreateUserReq struct {
    Name  string `json:"name" binding:"required"`
    Email string `json:"email" binding:"required,email"`
}

type UserResp struct {
    ID    int64  `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
}

func main() {
    r := gin.Default()

    // 使用 WrapHandler 包装业务逻辑
    r.POST("/users", ginserver.WrapHandler(
        func(ctx context.Context, req CreateUserReq) (UserResp, error) {
            user := UserResp{
                ID:    1,
                Name:  req.Name,
                Email: req.Email,
            }
            return user, nil
        },
    ))

    r.Run(":8080")
}
```

### Client 端基础用法

```go
package main

import (
    "context"
    "fmt"
    restyclient "github.com/zhangzqs/go-typed-rpc/resty-client"
    "resty.dev/v3"
)

type CreateUserReq struct {
    Name  string `json:"name"`
    Email string `json:"email"`
}

type UserResp struct {
    ID    int64  `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
}

func main() {
    // 创建 resty 客户端
    restyClient := resty.New()

    // 创建类型安全的客户端处理器
    createUser := restyclient.NewClient[CreateUserReq, UserResp](
        restyClient,
        "POST",
        "http://localhost:8080/users",
    )

    // 调用 API
    user, err := createUser(context.Background(), CreateUserReq{
        Name:  "Alice",
        Email: "alice@example.com",
    })
    if err != nil {
        panic(err)
    }

    fmt.Printf("Created user: %+v\n", user)
}
```

## Server 端详细说明

### 四种处理器类型

#### 1. WrapHandler - 完整的输入输出

```go
// 创建用户：有输入有输出
r.POST("/users", ginserver.WrapHandler(
    func(ctx context.Context, req CreateUserReq) (UserResp, error) {
        // 业务逻辑
        return user, nil
    },
))

// 获取用户：URI 参数
type GetUserReq struct {
    ID int64 `uri:"id"`
}

r.GET("/users/:id", ginserver.WrapHandler(
    func(ctx context.Context, req GetUserReq) (UserResp, error) {
        return getUserByID(req.ID)
    },
))
```

#### 2. WrapGetter - 只有输出

```go
type HealthResp struct {
    Status string `json:"status"`
}

// 健康检查：无需输入参数
r.GET("/health", ginserver.WrapGetter(
    func(ctx context.Context) (HealthResp, error) {
        return HealthResp{Status: "ok"}, nil
    },
))
```

#### 3. WrapConsumer - 只有输入

```go
type DeleteUserReq struct {
    ID int64 `uri:"id"`
}

// 删除用户：无需返回数据
r.DELETE("/users/:id", ginserver.WrapConsumer(
    func(ctx context.Context, req DeleteUserReq) error {
        return deleteUser(req.ID)
    },
))
```

#### 4. WrapAction - 无输入输出

```go
// 触发任务：无输入无输出
r.POST("/tasks/sync", ginserver.WrapAction(
    func(ctx context.Context) error {
        return triggerSyncTask()
    },
))
```

### 自动参数绑定

支持多种数据源的自动绑定：

```go
type GetArticleReq struct {
    ID       int64  `uri:"id"`         // URI 参数
    Page     int    `form:"page"`      // Query 参数
    PageSize int    `form:"page_size"` // Query 参数
}

// GET /articles/:id?page=1&page_size=10
r.GET("/articles/:id", ginserver.WrapHandler(
    func(ctx context.Context, req GetArticleReq) (Article, error) {
        return getArticle(req.ID, req.Page, req.PageSize)
    },
))
```

### 自定义选项

```go
// 自定义错误处理
customErrorHandler := func(c *gin.Context, err error) {
    c.JSON(http.StatusBadRequest, gin.H{
        "code":    "ERROR",
        "message": err.Error(),
    })
}

r.POST("/users", ginserver.WrapHandler(
    createUserHandler,
    ginserver.WithErrorHandler(customErrorHandler),
))
```

## Client 端详细说明

### 智能参数绑定

Client 端支持通过结构体标签自动处理不同类型的参数：

- `path` - 路径参数
- `query` / `form` - Query 参数
- `header` - 请求头
- `json` - JSON 请求体

#### 1. 路径参数

```go
type GetUserReq struct {
    ID int64 `path:"id"`
}

// GET /users/{id}
getUser := restyclient.NewClient[GetUserReq, UserResp](
    restyClient,
    "GET",
    "http://localhost:8080/users/{id}",
)

user, err := getUser(ctx, GetUserReq{ID: 123})
```

#### 2. Query 参数

```go
type ListUsersReq struct {
    Page     int `query:"page"`
    PageSize int `query:"page_size"`
}

// GET /users?page=1&page_size=10
listUsers := restyclient.NewClient[ListUsersReq, []UserResp](
    restyClient,
    "GET",
    "http://localhost:8080/users",
)

users, err := listUsers(ctx, ListUsersReq{
    Page:     1,
    PageSize: 10,
})
```

#### 3. 请求头

```go
type AuthReq struct {
    Token string `header:"Authorization"`
    Name  string `json:"name"`
}

// 请求头 + JSON body
createWithAuth := restyclient.NewClient[AuthReq, UserResp](
    restyClient,
    "POST",
    "http://localhost:8080/users",
)

user, err := createWithAuth(ctx, AuthReq{
    Token: "Bearer token123",
    Name:  "Alice",
})
```

#### 4. 组合使用

```go
type UpdateArticleReq struct {
    ID      int64  `path:"id"`              // 路径参数
    Token   string `header:"Authorization"` // 请求头
    Verbose bool   `query:"verbose"`        // Query 参数
    Title   string `json:"title"`           // JSON body
    Content string `json:"content"`         // JSON body
}

// PUT /articles/{id}?verbose=true
// Authorization: Bearer token
// Body: {"title": "...", "content": "..."}
updateArticle := restyclient.NewClient[UpdateArticleReq, Article](
    restyClient,
    "PUT",
    "http://localhost:8080/articles/{id}",
)

article, err := updateArticle(ctx, UpdateArticleReq{
    ID:      1,
    Token:   "Bearer token123",
    Verbose: true,
    Title:   "New Title",
    Content: "New Content",
})
```

### 便捷函数

#### NewGetter - GET 请求

```go
// GET /health
healthCheck := restyclient.NewGetter[HealthResp](
    restyClient,
    "GET",
    "http://localhost:8080/health",
)

health, err := healthCheck(ctx)
```

#### NewConsumer - 只有输入的请求

```go
// POST /users
createUser := restyclient.NewConsumer[CreateUserReq](
    restyClient,
    "POST",
    "http://localhost:8080/users",
)

err := createUser(ctx, CreateUserReq{
    Name:  "Alice",
    Email: "alice@example.com",
})
```

#### NewAction - 无输入输出的请求

```go
// DELETE /users/{id}
deleteUser := restyclient.NewAction(
    restyClient,
    "DELETE",
    "http://localhost:8080/users/{id}",
)

err := deleteUser(ctx)
```

### 自定义选项

```go
// 自定义请求编码器
customEncoder := func(req *resty.Request, input any) error {
    req.SetHeader("X-Custom", "value")
    req.SetBody(input)
    return nil
}

// 自定义响应解码器
customDecoder := func(resp *resty.Response) (any, error) {
    var result WrapperResponse
    json.Unmarshal(resp.Bytes(), &result)
    return result.Data, nil
}

// 自定义错误处理
customErrorHandler := func(resp *resty.Response, err error) error {
    if err != nil {
        return err
    }
    if resp.IsError() {
        return fmt.Errorf("API error: %s", resp.Status())
    }
    return nil
}

handler := restyclient.NewClient[Req, Resp](
    restyClient,
    "POST",
    "/api/endpoint",
    restyclient.WithEncoder(customEncoder),
    restyclient.WithDecoder(customDecoder),
    restyclient.WithErrorHandler(customErrorHandler),
)
```

## 完整示例

查看 [examples/fullstack](./examples/fullstack) 目录，展示 Server 和 Client 的完整交互：

- ✅ Server 端所有处理器类型示例
- ✅ Client 端所有绑定方式示例
- ✅ 自定义选项使用示例
- ✅ 完整的 Server/Client 交互示例

运行示例：

```bash
cd examples/fullstack
go run main.go
```

## API 参考

### gin-server 包

包名：`ginserver`

#### 核心函数

- `WrapHandler[I, O any](h handler.HandlerFunc[I, O], options...) gin.HandlerFunc`
- `WrapGetter[O any](h handler.GetterHandlerFunc[O], options...) gin.HandlerFunc`
- `WrapConsumer[I any](h handler.ConsumerHandlerFunc[I], options...) gin.HandlerFunc`
- `WrapAction(h handler.ActionHandlerFunc, options...) gin.HandlerFunc`

#### 选项函数

- `WithDecoder(decoder DecoderFunc) WrapHandlerOptionFunc`
- `WithEncoder(encoder EncoderFunc) WrapHandlerOptionFunc`
- `WithErrorHandler(errHandler ErrorHandlerFunc) WrapHandlerOptionFunc`

#### 函数签名

- `DecoderFunc`: `func(c *gin.Context) (any, error)`
- `EncoderFunc`: `func(c *gin.Context, output any) error`
- `ErrorHandlerFunc`: `func(c *gin.Context, err error)`

### resty-client 包

包名：`restyclient`

#### 核心函数

- `NewClient[I, O any](client *resty.Client, method, url string, options...) handler.HandlerFunc[I, O]`
- `NewGetter[O any](client *resty.Client, method, url string, options...) handler.GetterHandlerFunc[O]`
- `NewConsumer[I any](client *resty.Client, method, url string, options...) handler.ConsumerHandlerFunc[I]`
- `NewAction(client *resty.Client, method, url string, options...) handler.ActionHandlerFunc`

#### 选项函数

- `WithEncoder(encoder RequestEncoderFunc) ClientOptionFunc`
- `WithDecoder(decoder ResponseDecoderFunc) ClientOptionFunc`
- `WithErrorHandler(errHandler ErrorHandlerFunc) ClientOptionFunc`

#### 函数签名

- `RequestEncoderFunc`: `func(req *resty.Request, input any) error`
- `ResponseDecoderFunc`: `func(resp *resty.Response) (any, error)`
- `ErrorHandlerFunc`: `func(resp *resty.Response, err error) error`

#### 支持的标签

- `path:"paramName"` - URL 路径参数
- `query:"paramName"` - URL Query 参数
- `form:"paramName"` - URL Query 参数（别名）
- `header:"HeaderName"` - HTTP 请求头
- `json:"fieldName"` - JSON 请求体字段

### handler 包

定义通用处理函数类型：

- `HandlerFunc[I, O any]`: `func(ctx context.Context, input I) (O, error)`
- `ActionHandlerFunc`: `func(ctx context.Context) error`
- `GetterHandlerFunc[O any]`: `func(ctx context.Context) (O, error)`
- `ConsumerHandlerFunc[I any]`: `func(ctx context.Context, args I) error`

## 测试

运行测试：

```bash
# 测试所有包
go test ./...

# 测试 gin-server 包
go test -v -cover ./gin-server

# 测试 resty-client 包
go test -v -cover ./resty-client

# 测试 examples/fullstack
cd examples/fullstack && go test -v
```

## 最佳实践

### Server 端

1. **参数验证**：使用 Gin 的 `binding` 标签进行参数验证
2. **错误处理**：定义业务错误类型，使用自定义错误处理器
3. **响应格式**：使用自定义编码器统一响应格式
4. **上下文传递**：使用 `context.Context` 传递请求级别的数据

### Client 端

1. **标签使用**：合理使用 `path`、`query`、`header`、`json` 标签
2. **类型安全**：充分利用泛型确保编译时类型安全
3. **错误处理**：根据业务需求自定义错误处理逻辑
4. **客户端复用**：创建并复用 resty.Client 实例

## 许可证

MIT License
