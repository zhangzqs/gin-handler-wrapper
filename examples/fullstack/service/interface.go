package service

import (
	"context"

	"github.com/zhangzqs/gin-handler-wrapper/examples/fullstack/model"
)

// ==================== 业务服务接口定义 ====================

// UserService 用户服务接口
type UserService interface {
	CreateUser(ctx context.Context, req model.CreateUserRequest) (model.User, error)
	GetUser(ctx context.Context, req model.GetUserRequest) (model.User, error)
	ListUsers(ctx context.Context, req model.ListUsersRequest) (model.ListUsersResponse, error)
	DeleteUser(ctx context.Context, req model.DeleteUserRequest) error
}

// ArticleService 文章服务接口
type ArticleService interface {
	UpdateArticle(ctx context.Context, req model.UpdateArticleRequest) (model.Article, error)
}

// HealthService 健康检查服务接口
type HealthService interface {
	Health(ctx context.Context) (model.HealthResponse, error)
}

// TaskService 任务服务接口
type TaskService interface {
	TriggerTask(ctx context.Context) error
}

// Service 综合业务服务接口
type Service interface {
	UserService
	ArticleService
	HealthService
	TaskService
}
