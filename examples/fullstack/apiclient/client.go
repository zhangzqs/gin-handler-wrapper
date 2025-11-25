package apiclient

import (
	"context"

	"github.com/go-resty/resty/v2"
	"github.com/zhangzqs/gin-handler-wrapper/client"
	"github.com/zhangzqs/gin-handler-wrapper/examples/fullstack/model"
	"github.com/zhangzqs/gin-handler-wrapper/examples/fullstack/service"
)

// ==================== API Client 结构体（实现 service 接口）====================

// Client API客户端
type Client struct {
	baseURL     string
	restyClient *resty.Client
}

// NewClient 创建新的API客户端
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL:     baseURL,
		restyClient: resty.New().SetBaseURL(baseURL),
	}
}

// 确保 Client 实现了所有服务接口
var (
	_ service.Service = (*Client)(nil)
)

// CreateUser 创建用户
func (c *Client) CreateUser(ctx context.Context, req model.CreateUserRequest) (model.User, error) {
	handler := client.NewClient[model.CreateUserRequest, model.User](
		c.restyClient,
		"POST",
		"/users",
	)
	return handler(ctx, req)
}

// GetUser 获取用户
func (c *Client) GetUser(ctx context.Context, req model.GetUserRequest) (model.User, error) {
	handler := client.NewClient[model.GetUserRequest, model.User](
		c.restyClient,
		"GET",
		"/users/{id}",
	)
	return handler(ctx, req)
}

// ListUsers 获取用户列表
func (c *Client) ListUsers(ctx context.Context, req model.ListUsersRequest) (model.ListUsersResponse, error) {
	handler := client.NewClient[model.ListUsersRequest, model.ListUsersResponse](
		c.restyClient,
		"GET",
		"/users",
	)
	return handler(ctx, req)
}

// DeleteUser 删除用户
func (c *Client) DeleteUser(ctx context.Context, req model.DeleteUserRequest) error {
	handler := client.NewClient[model.DeleteUserRequest, struct{}](
		c.restyClient,
		"DELETE",
		"/users/{id}",
	)
	_, err := handler(ctx, req)
	return err
}

// UpdateArticle 更新文章
func (c *Client) UpdateArticle(ctx context.Context, req model.UpdateArticleRequest) (model.Article, error) {
	handler := client.NewClient[model.UpdateArticleRequest, model.Article](
		c.restyClient,
		"PUT",
		"/articles/{id}",
	)
	return handler(ctx, req)
}

// Health 健康检查
func (c *Client) Health(ctx context.Context) (model.HealthResponse, error) {
	handler := client.NewGetter[model.HealthResponse](c.restyClient, "/health")
	return handler(ctx)
}

// TriggerTask 触发任务
func (c *Client) TriggerTask(ctx context.Context) error {
	handler := client.NewAction(c.restyClient, "POST", "/tasks")
	return handler(ctx)
}
