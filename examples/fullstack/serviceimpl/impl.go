package serviceimpl

import (
	"context"
	"log"
	"time"

	"github.com/zhangzqs/gin-handler-wrapper/examples/fullstack/model"
	"github.com/zhangzqs/gin-handler-wrapper/examples/fullstack/service"
	"github.com/zhangzqs/gin-handler-wrapper/examples/fullstack/store"
)

// ==================== 业务逻辑实现（纯业务逻辑，不依赖HTTP）====================

// ServiceImpl 业务服务实现
type ServiceImpl struct {
	store *store.Store
}

// NewService 创建新的服务实例
func NewService(s *store.Store) *ServiceImpl {
	return &ServiceImpl{
		store: s,
	}
}

// 确保 ServiceImpl 实现了所有服务接口
var (
	_ service.Service = (*ServiceImpl)(nil)
)

// CreateUser 创建用户
func (s *ServiceImpl) CreateUser(ctx context.Context, req model.CreateUserRequest) (model.User, error) {
	return s.store.CreateUser(req.Name, req.Email), nil
}

// GetUser 获取用户
func (s *ServiceImpl) GetUser(ctx context.Context, req model.GetUserRequest) (model.User, error) {
	return s.store.GetUser(req.ID)
}

// ListUsers 获取用户列表
func (s *ServiceImpl) ListUsers(ctx context.Context, req model.ListUsersRequest) (model.ListUsersResponse, error) {
	// 设置默认值
	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 10
	}

	userList := s.store.ListUsers()

	return model.ListUsersResponse{
		Total: len(userList),
		Users: userList,
	}, nil
}

// UpdateArticle 更新文章
func (s *ServiceImpl) UpdateArticle(ctx context.Context, req model.UpdateArticleRequest) (model.Article, error) {
	return s.store.UpdateArticle(req.ID, req.Title, req.Content), nil
}

// DeleteUser 删除用户
func (s *ServiceImpl) DeleteUser(ctx context.Context, req model.DeleteUserRequest) error {
	return s.store.DeleteUser(req.ID)
}

// Health 健康检查
func (s *ServiceImpl) Health(ctx context.Context) (model.HealthResponse, error) {
	return model.HealthResponse{
		Status:    "ok",
		Timestamp: time.Now(),
	}, nil
}

// TriggerTask 触发任务
func (s *ServiceImpl) TriggerTask(ctx context.Context) error {
	log.Println("Task triggered successfully")
	return nil
}
