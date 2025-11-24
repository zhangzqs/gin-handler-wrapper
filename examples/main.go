package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	wrapper "github.com/zhangzqs/gin-handler-wrapper"
)

// ==================== 数据模型 ====================

// User 用户模型
type User struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

// Article 文章模型
type Article struct {
	ID      int64  `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
	Author  string `json:"author"`
}

// ==================== 请求/响应类型 ====================

// CreateUserRequest 创建用户请求
type CreateUserRequest struct {
	Name  string `json:"name" binding:"required"`
	Email string `json:"email" binding:"required,email"`
}

// GetUserRequest 获取用户请求
type GetUserRequest struct {
	ID int64 `uri:"id" binding:"required,gt=0"`
}

// ListUsersRequest 获取用户列表请求
type ListUsersRequest struct {
	Page     int `form:"page" binding:"gte=1"`
	PageSize int `form:"page_size" binding:"gte=1,lte=100"`
}

// UpdateUserRequest 更新用户请求
type UpdateUserRequest struct {
	ID    int64  `uri:"id" binding:"required,gt=0"`
	Name  string `json:"name"`
	Email string `json:"email" binding:"omitempty,email"`
}

// DeleteUserRequest 删除用户请求
type DeleteUserRequest struct {
	ID int64 `uri:"id" binding:"required,gt=0"`
}

// SearchArticlesRequest 搜索文章请求
type SearchArticlesRequest struct {
	Keyword  string `form:"keyword" binding:"required"`
	Page     int    `form:"page" binding:"gte=1"`
	PageSize int    `form:"page_size" binding:"gte=1,lte=100"`
}

// HealthResponse 健康检查响应
type HealthResponse struct {
	Status  string    `json:"status"`
	Time    time.Time `json:"time"`
	Version string    `json:"version"`
}

// ListResponse 列表响应
type ListResponse[T any] struct {
	Items      []T   `json:"items"`
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	TotalPages int   `json:"total_pages"`
}

// ==================== 业务错误 ====================

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrInvalidID         = errors.New("invalid id")
)

// ==================== 模拟数据库 ====================

var (
	users          = make(map[int64]*User)
	nextID   int64 = 1
	articles       = []Article{
		{ID: 1, Title: "Go 语言入门", Content: "这是一篇关于 Go 的文章", Author: "Alice"},
		{ID: 2, Title: "Gin 框架使用", Content: "这是一篇关于 Gin 的文章", Author: "Bob"},
		{ID: 3, Title: "泛型编程实战", Content: "这是一篇关于泛型的文章", Author: "Charlie"},
	}
)

// ==================== 业务处理器 ====================

// CreateUser 创建用户 - 完整的输入输出
func CreateUser(ctx context.Context, req CreateUserRequest) (*User, error) {
	// 检查邮箱是否已存在
	for _, u := range users {
		if u.Email == req.Email {
			return nil, ErrUserAlreadyExists
		}
	}

	user := &User{
		ID:        nextID,
		Name:      req.Name,
		Email:     req.Email,
		CreatedAt: time.Now(),
	}
	users[nextID] = user
	nextID++

	return user, nil
}

// GetUser 获取用户 - 完整的输入输出
func GetUser(ctx context.Context, req GetUserRequest) (*User, error) {
	user, ok := users[req.ID]
	if !ok {
		return nil, ErrUserNotFound
	}
	return user, nil
}

// ListUsers 获取用户列表 - 完整的输入输出
func ListUsers(ctx context.Context, req ListUsersRequest) (*ListResponse[User], error) {
	// 设置默认值
	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 10
	}

	// 获取所有用户
	allUsers := make([]User, 0, len(users))
	for _, u := range users {
		allUsers = append(allUsers, *u)
	}

	// 分页
	start := (req.Page - 1) * req.PageSize
	end := start + req.PageSize
	if start >= len(allUsers) {
		start = len(allUsers)
	}
	if end > len(allUsers) {
		end = len(allUsers)
	}

	items := allUsers[start:end]
	total := int64(len(allUsers))
	totalPages := (int(total) + req.PageSize - 1) / req.PageSize

	return &ListResponse[User]{
		Items:      items,
		Total:      total,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: totalPages,
	}, nil
}

// UpdateUser 更新用户 - 只有输入，无输出
func UpdateUser(ctx context.Context, req UpdateUserRequest) error {
	user, ok := users[req.ID]
	if !ok {
		return ErrUserNotFound
	}

	if req.Name != "" {
		user.Name = req.Name
	}
	if req.Email != "" {
		user.Email = req.Email
	}

	return nil
}

// DeleteUser 删除用户 - 只有输入，无输出
func DeleteUser(ctx context.Context, req DeleteUserRequest) error {
	if _, ok := users[req.ID]; !ok {
		return ErrUserNotFound
	}
	delete(users, req.ID)
	return nil
}

// GetHealth 健康检查 - 只有输出，无输入
func GetHealth(ctx context.Context) (*HealthResponse, error) {
	return &HealthResponse{
		Status:  "ok",
		Time:    time.Now(),
		Version: "1.0.0",
	}, nil
}

// SearchArticles 搜索文章 - 完整的输入输出
func SearchArticles(ctx context.Context, req SearchArticlesRequest) (*ListResponse[Article], error) {
	// 设置默认值
	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 10
	}

	// 简单的关键词搜索
	var result []Article
	for _, a := range articles {
		if contains(a.Title, req.Keyword) || contains(a.Content, req.Keyword) {
			result = append(result, a)
		}
	}

	// 分页
	start := (req.Page - 1) * req.PageSize
	end := start + req.PageSize
	if start >= len(result) {
		start = len(result)
	}
	if end > len(result) {
		end = len(result)
	}

	items := result[start:end]
	total := int64(len(result))
	totalPages := (int(total) + req.PageSize - 1) / req.PageSize

	return &ListResponse[Article]{
		Items:      items,
		Total:      total,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: totalPages,
	}, nil
}

// ClearCache 清除缓存 - 无输入输出
func ClearCache(ctx context.Context) error {
	log.Println("Cache cleared")
	return nil
}

// SyncData 同步数据 - 无输入输出
func SyncData(ctx context.Context) error {
	log.Println("Data synchronization started")
	// 模拟耗时操作
	time.Sleep(100 * time.Millisecond)
	log.Println("Data synchronization completed")
	return nil
}

// ==================== 自定义错误处理器 ====================

func customErrorHandler(c *gin.Context, err error) {
	// 根据错误类型返回不同的状态码
	switch {
	case errors.Is(err, ErrUserNotFound):
		c.JSON(http.StatusNotFound, gin.H{
			"code":    "USER_NOT_FOUND",
			"message": err.Error(),
		})
	case errors.Is(err, ErrUserAlreadyExists):
		c.JSON(http.StatusConflict, gin.H{
			"code":    "USER_ALREADY_EXISTS",
			"message": err.Error(),
		})
	case errors.Is(err, ErrInvalidID):
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_ID",
			"message": err.Error(),
		})
	default:
		// 默认返回 500
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "INTERNAL_SERVER_ERROR",
			"message": err.Error(),
		})
	}
}

// ==================== 辅助函数 ====================

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && searchInString(s, substr)))
}

func searchInString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func wrapHandler[I, O any](h wrapper.Handler[I, O]) gin.HandlerFunc {
	return wrapper.WrapHandler(
		h,
		wrapper.WithErrorHandler(customErrorHandler),
	)
}

func wrapAction(h wrapper.ActionHandler) gin.HandlerFunc {
	return wrapper.WrapAction(
		h,
		wrapper.WithErrorHandler(customErrorHandler),
	)
}

func wrapGetter[O any](h wrapper.GetterHandler[O]) gin.HandlerFunc {
	return wrapper.WrapGetter(
		h,
		wrapper.WithErrorHandler(customErrorHandler),
	)
}

func wrapConsumer[I any](h wrapper.ConsumerHandler[I]) gin.HandlerFunc {
	return wrapper.WrapConsumer(
		h,
		wrapper.WithErrorHandler(customErrorHandler),
	)
}

// ==================== 主函数 ====================

func main() {
	r := gin.Default()

	// 初始化一些测试数据
	users[1] = &User{ID: 1, Name: "Alice", Email: "alice@example.com", CreatedAt: time.Now()}
	users[2] = &User{ID: 2, Name: "Bob", Email: "bob@example.com", CreatedAt: time.Now()}
	nextID = 3

	// ==================== API 路由 ====================

	// 1. WrapHandler - 完整的输入输出
	r.POST("/users", wrapHandler(CreateUser))
	r.GET("/users/:id", wrapHandler(GetUser))
	r.GET("/users", wrapHandler(ListUsers))
	r.GET("/articles/search", wrapHandler(SearchArticles))

	// 2. WrapConsumer - 只有输入，无输出
	r.PUT("/users/:id", wrapConsumer(UpdateUser))
	r.DELETE("/users/:id", wrapConsumer(DeleteUser))

	// 3. WrapGetter - 只有输出，无输入
	r.GET("/health", wrapGetter(GetHealth))

	// 4. WrapAction - 无输入输出
	r.POST("/cache/clear", wrapAction(ClearCache))
	r.POST("/data/sync", wrapAction(SyncData))

	// ==================== 启动服务 ====================

	fmt.Println("===========================================")
	fmt.Println("🚀 Gin Handler Wrapper Complete Example")
	fmt.Println("===========================================")
	fmt.Println()
	fmt.Println("API Endpoints:")
	fmt.Println()
	fmt.Println("WrapHandler (完整输入输出):")
	fmt.Println("  POST   /users                - 创建用户")
	fmt.Println("  GET    /users/:id            - 获取用户")
	fmt.Println("  GET    /users?page=1&page_size=10 - 获取用户列表")
	fmt.Println("  GET    /articles/search?keyword=Go - 搜索文章")
	fmt.Println()
	fmt.Println("WrapConsumer (只有输入):")
	fmt.Println("  PUT    /users/:id            - 更新用户")
	fmt.Println("  DELETE /users/:id            - 删除用户")
	fmt.Println()
	fmt.Println("WrapGetter (只有输出):")
	fmt.Println("  GET    /health               - 健康检查")
	fmt.Println()
	fmt.Println("WrapAction (无输入输出):")
	fmt.Println("  POST   /cache/clear          - 清除缓存")
	fmt.Println("  POST   /data/sync            - 同步数据")
	fmt.Println()
	fmt.Println("===========================================")
	fmt.Println("Server started at http://localhost:8080")
	fmt.Println("===========================================")

	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
