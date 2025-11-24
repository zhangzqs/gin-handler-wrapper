# Gin Handler Wrapper

一个类型安全的 Gin 处理器包装库，使用 Go 泛型提供优雅的请求/响应处理。

## 特性

- ✅ **类型安全**：使用 Go 泛型实现编译时类型检查
- ✅ **自动绑定**：支持 URI、Query、JSON、Form 等多种数据源
- ✅ **灵活定制**：可自定义解码器、编码器和错误处理器
- ✅ **便捷函数**：提供多种模板函数覆盖常见场景
- ✅ **清晰架构**：职责分离，代码易于维护

## 安装

```bash
go get github.com/your-username/gin-handler-wrapper
```

## 快速开始

### 基础用法 - WrapHandler

完整的输入输出处理器：

```go
package main

import (
    "context"
    "github.com/gin-gonic/gin"
    wrapper "github.com/your-username/gin-handler-wrapper"
)

type CreateUserReq struct {
    Name  string `json:"name" binding:"required"`
    Email string `json:"email" binding:"required,email"`
}

type CreateUserResp struct {
    ID    int64  `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
}

func main() {
    r := gin.Default()

    // 使用 WrapHandler 包装业务逻辑
    r.POST("/users", wrapper.WrapHandler(
        func(ctx context.Context, req CreateUserReq) (CreateUserResp, error) {
            // 业务逻辑：创建用户
            user := CreateUserResp{
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

### WrapGetter - 只有输出

适用于获取数据、健康检查等场景：

```go
type HealthResp struct {
    Status string `json:"status"`
    Time   string `json:"time"`
}

// 健康检查
r.GET("/health", wrapper.WrapGetter(
    func(ctx context.Context) (HealthResp, error) {
        return HealthResp{
            Status: "ok",
            Time:   time.Now().Format(time.RFC3339),
        }, nil
    },
))

// 获取用户列表
type User struct {
    ID   int64  `json:"id"`
    Name string `json:"name"`
}

r.GET("/users", wrapper.WrapGetter(
    func(ctx context.Context) ([]User, error) {
        users := []User{
            {ID: 1, Name: "Alice"},
            {ID: 2, Name: "Bob"},
        }
        return users, nil
    },
))
```

### WrapConsumer - 只有输入

适用于删除、更新等不需要返回数据的场景：

```go
type DeleteUserReq struct {
    ID int64 `uri:"id" binding:"required"`
}

// 删除用户
r.DELETE("/users/:id", wrapper.WrapConsumer(
    func(ctx context.Context, req DeleteUserReq) error {
        // 业务逻辑：删除用户
        return deleteUser(req.ID)
    },
))

type UpdatePasswordReq struct {
    UserID      int64  `uri:"id" binding:"required"`
    OldPassword string `json:"old_password" binding:"required"`
    NewPassword string `json:"new_password" binding:"required,min=6"`
}

// 更新密码
r.PUT("/users/:id/password", wrapper.WrapConsumer(
    func(ctx context.Context, req UpdatePasswordReq) error {
        // 业务逻辑：更新密码
        return updatePassword(req.UserID, req.OldPassword, req.NewPassword)
    },
))
```

### WrapAction - 无输入输出

适用于触发任务、执行操作等场景：

```go
// 触发数据同步任务
r.POST("/tasks/sync", wrapper.WrapAction(
    func(ctx context.Context) error {
        // 触发异步任务
        return triggerSyncTask()
    },
))

// 清除缓存
r.POST("/cache/clear", wrapper.WrapAction(
    func(ctx context.Context) error {
        return clearCache()
    },
))
```

## 高级用法

### 自动参数绑定

支持多种数据源的自动绑定：

```go
type GetArticleReq struct {
    ID       int64  `uri:"id"`           // 从 URI 参数绑定
    Page     int    `form:"page"`        // 从 Query 参数绑定
    PageSize int    `form:"page_size"`   // 从 Query 参数绑定
    Token    string `header:"X-Token"`   // 从 Header 绑定
}

// GET /articles/:id?page=1&page_size=10
r.GET("/articles/:id", wrapper.WrapHandler(
    func(ctx context.Context, req GetArticleReq) (Article, error) {
        // req.ID 来自 URI
        // req.Page 和 req.PageSize 来自 Query
        // req.Token 来自 Header
        return getArticle(req.ID, req.Page, req.PageSize)
    },
))
```

### 自定义错误处理

```go
// 自定义错误处理器
customErrorHandler := func(c *gin.Context, err error) {
    if errors.Is(err, ErrNotFound) {
        c.JSON(http.StatusNotFound, gin.H{
            "code":    "NOT_FOUND",
            "message": err.Error(),
        })
        return
    }

    if errors.Is(err, ErrUnauthorized) {
        c.JSON(http.StatusUnauthorized, gin.H{
            "code":    "UNAUTHORIZED",
            "message": err.Error(),
        })
        return
    }

    // 默认错误
    c.JSON(http.StatusInternalServerError, gin.H{
        "code":    "INTERNAL_ERROR",
        "message": err.Error(),
    })
}

// 使用自定义错误处理器
r.GET("/users/:id", wrapper.WrapHandler(
    func(ctx context.Context, req GetUserReq) (User, error) {
        return getUserByID(req.ID)
    },
    wrapper.WithErrorHandler[GetUserReq, User](customErrorHandler),
))
```

### 自定义编码器

```go
// 自定义响应编码器（例如：统一响应格式）
customEncoder := func() wrapper.EncoderFunc[User] {
    return func(c *gin.Context, output User) error {
        c.JSON(http.StatusOK, gin.H{
            "code":    "SUCCESS",
            "message": "操作成功",
            "data":    output,
        })
        return nil
    }
}

r.GET("/users/:id", wrapper.WrapHandler(
    func(ctx context.Context, req GetUserReq) (User, error) {
        return getUserByID(req.ID)
    },
    wrapper.WithEncoder(customEncoder()),
))
```

### 自定义解码器

```go
// 自定义解码器（例如：从 Header 中提取用户信息）
customDecoder := func() wrapper.DecoderFunc[MyRequest] {
    return func(c *gin.Context) (MyRequest, error) {
        var req MyRequest

        // 先绑定标准参数
        if err := c.ShouldBind(&req); err != nil {
            return req, err
        }

        // 从 Header 提取用户信息
        userID := c.GetHeader("X-User-ID")
        req.UserID = userID

        return req, nil
    }
}

r.POST("/api/data", wrapper.WrapHandler(
    func(ctx context.Context, req MyRequest) (MyResponse, error) {
        // req.UserID 已经从 Header 中提取
        return processData(req)
    },
    wrapper.WithDecoder(customDecoder()),
))
```

## API 参考

### 核心函数

#### WrapHandler[I, O any]

包装完整的输入输出处理器。

```go
func WrapHandler[I, O any](
    h Handler[I, O],
    options ...WrapHandlerOptionFunc[I, O],
) gin.HandlerFunc
```

#### WrapGetter[O any]

包装只有输出的处理器。

```go
func WrapGetter[O any](
    h func(ctx context.Context) (O, error),
    options ...WrapHandlerOptionFunc[struct{}, O],
) gin.HandlerFunc
```

#### WrapConsumer[I any]

包装只有输入的处理器。

```go
func WrapConsumer[I any](
    h func(ctx context.Context, args I) error,
    options ...WrapHandlerOptionFunc[I, struct{}],
) gin.HandlerFunc
```

#### WrapAction

包装无输入输出的处理器。

```go
func WrapAction(
    h func(ctx context.Context) error,
    options ...WrapHandlerOptionFunc[struct{}, struct{}],
) gin.HandlerFunc
```

### 选项函数

#### WithDecoder

自定义解码器。

```go
func WithDecoder[I, O any](decoder DecoderFunc[I]) WrapHandlerOptionFunc[I, O]
```

#### WithEncoder

自定义编码器。

```go
func WithEncoder[I, O any](encoder EncoderFunc[O]) WrapHandlerOptionFunc[I, O]
```

#### WithErrorHandler

自定义错误处理器。

```go
func WithErrorHandler[I, O any](errHandler ErrorHandlerFunc) WrapHandlerOptionFunc[I, O]
```

## 最佳实践

1. **参数验证**：使用 Gin 的 `binding` 标签进行参数验证
2. **错误处理**：定义业务错误类型，使用自定义错误处理器
3. **响应格式**：使用自定义编码器统一响应格式
4. **上下文传递**：使用 `context.Context` 传递请求级别的数据
5. **泛型使用**：合理使用泛型，避免过度抽象

## 许可证

MIT License
