package service

import (
	"context"

	"github.com/zhangzqs/gin-handler-wrapper/examples/fullstack/model"
)

// Service 综合业务服务接口
type Service interface {
	CreateUser(ctx context.Context, req model.CreateUserRequest) (model.User, error)
	GetUser(ctx context.Context, req model.GetUserRequest) (model.User, error)
	ListUsers(ctx context.Context, req model.ListUsersRequest) (model.ListUsersResponse, error)
	DeleteUser(ctx context.Context, req model.DeleteUserRequest) error
	UpdateArticle(ctx context.Context, req model.UpdateArticleRequest) (model.Article, error)
	Health(ctx context.Context) (model.HealthResponse, error)
	TriggerTask(ctx context.Context) error
}
