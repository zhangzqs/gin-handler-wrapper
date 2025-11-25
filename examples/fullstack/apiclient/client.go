package apiclient

import (
	"context"
	"net/http"

	"github.com/zhangzqs/go-typed-rpc/examples/fullstack/model"
	"github.com/zhangzqs/go-typed-rpc/examples/fullstack/service"
	restyclient "github.com/zhangzqs/go-typed-rpc/resty-client"
	"resty.dev/v3"
)

// ==================== API Client 结构体（实现 service 接口）====================

// Client API客户端
type Client struct {
	cli *resty.Client
}

// 确保 Client 实现了所有服务接口
var _ service.Service = (*Client)(nil)

// NewClient 创建新的API客户端
func NewClient(cli *resty.Client) *Client {
	return &Client{cli: cli}
}

// CreateUser 创建用户
func (c *Client) CreateUser(ctx context.Context, req model.CreateUserRequest) (model.User, error) {
	return restyclient.NewClient[model.CreateUserRequest, model.User](
		c.cli, http.MethodPost, "/users",
	)(ctx, req)
}

// GetUser 获取用户
func (c *Client) GetUser(ctx context.Context, req model.GetUserRequest) (model.User, error) {
	return restyclient.NewClient[model.GetUserRequest, model.User](
		c.cli, http.MethodGet, "/users/{id}",
	)(ctx, req)
}

// ListUsers 获取用户列表
func (c *Client) ListUsers(ctx context.Context, req model.ListUsersRequest) (model.ListUsersResponse, error) {
	return restyclient.NewClient[model.ListUsersRequest, model.ListUsersResponse](
		c.cli, http.MethodGet, "/users",
	)(ctx, req)
}

// DeleteUser 删除用户
func (c *Client) DeleteUser(ctx context.Context, req model.DeleteUserRequest) error {
	return restyclient.NewConsumer[model.DeleteUserRequest](
		c.cli, http.MethodDelete, "/users/{id}",
	)(ctx, req)
}

// UpdateArticle 更新文章
func (c *Client) UpdateArticle(ctx context.Context, req model.UpdateArticleRequest) (model.Article, error) {
	return restyclient.NewClient[model.UpdateArticleRequest, model.Article](
		c.cli, http.MethodPut, "/articles/{id}",
	)(ctx, req)
}

// Health 健康检查
func (c *Client) Health(ctx context.Context) (model.HealthResponse, error) {
	return restyclient.NewGetter[model.HealthResponse](c.cli, http.MethodGet, "/health")(ctx)
}

// TriggerTask 触发任务
func (c *Client) TriggerTask(ctx context.Context) error {
	return restyclient.NewAction(c.cli, http.MethodPost, "/tasks")(ctx)
}
