package model

import "time"

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

// ==================== Server端请求/响应类型 ====================

// CreateUserRequest 创建用户请求
type CreateUserRequest struct {
	Name  string `json:"name" binding:"required"`
	Email string `json:"email" binding:"required,email"`
}

// GetUserRequest 获取用户请求（URI参数/路径参数）
type GetUserRequest struct {
	ID int64 `uri:"id" path:"id"` // uri for server, path for client
}

// ListUsersRequest 获取用户列表请求（Query参数）
type ListUsersRequest struct {
	Page     int `form:"page" query:"page" binding:"gte=1"`               // form for server, query for client
	PageSize int `form:"page_size" query:"page_size" binding:"gte=1,lte=100"` // form for server, query for client
}

// UpdateArticleRequest 更新文章请求（组合参数）
// Note: When combining URI + JSON params, avoid using binding validation tags
// as Gin validates the entire struct after each binding step
type UpdateArticleRequest struct {
	ID      int64  `uri:"id" path:"id"` // uri for server, path for client
	Title   string `json:"title"`       // json body field
	Content string `json:"content"`     // json body field
}

// DeleteUserRequest 删除用户请求
type DeleteUserRequest struct {
	ID int64 `uri:"id" path:"id"` // uri for server, path for client
}

// ListUsersResponse 用户列表响应
type ListUsersResponse struct {
	Total int    `json:"total"`
	Users []User `json:"users"`
}

// HealthResponse 健康检查响应
type HealthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
}

// ErrorResponse 错误响应
type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
