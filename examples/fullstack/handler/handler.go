package handler

import (
	"context"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/zhangzqs/gin-handler-wrapper/examples/fullstack/model"
	"github.com/zhangzqs/gin-handler-wrapper/examples/fullstack/service"
	"github.com/zhangzqs/gin-handler-wrapper/server"
)

// ==================== Server端业务处理器（HTTP适配器层）====================

// Handler Server端处理器，作为HTTP适配器，将HTTP请求转发到业务服务层
type Handler struct {
	svc service.Service
}

// NewHandler 创建新的处理器
func NewHandler(svc service.Service) *Handler {
	return &Handler{
		svc: svc,
	}
}

// 确保 Handler 实现了所有服务接口
var (
	_ service.UserService    = (*Handler)(nil)
	_ service.ArticleService = (*Handler)(nil)
	_ service.HealthService  = (*Handler)(nil)
	_ service.TaskService    = (*Handler)(nil)
)

// CreateUser 创建用户
func (h *Handler) CreateUser(ctx context.Context, req model.CreateUserRequest) (model.User, error) {
	return h.svc.CreateUser(ctx, req)
}

// GetUser 获取用户
func (h *Handler) GetUser(ctx context.Context, req model.GetUserRequest) (model.User, error) {
	return h.svc.GetUser(ctx, req)
}

// ListUsers 获取用户列表
func (h *Handler) ListUsers(ctx context.Context, req model.ListUsersRequest) (model.ListUsersResponse, error) {
	return h.svc.ListUsers(ctx, req)
}

// UpdateArticle 更新文章
func (h *Handler) UpdateArticle(ctx context.Context, req model.UpdateArticleRequest) (model.Article, error) {
	return h.svc.UpdateArticle(ctx, req)
}

// DeleteUser 删除用户
func (h *Handler) DeleteUser(ctx context.Context, req model.DeleteUserRequest) error {
	return h.svc.DeleteUser(ctx, req)
}

// Health 健康检查
func (h *Handler) Health(ctx context.Context) (model.HealthResponse, error) {
	return h.svc.Health(ctx)
}

// TriggerTask 触发任务
func (h *Handler) TriggerTask(ctx context.Context) error {
	return h.svc.TriggerTask(ctx)
}

// ==================== 自定义错误处理器 ====================

// CustomErrorHandler 自定义错误处理器
func (h *Handler) CustomErrorHandler(c *gin.Context, err error) {
	log.Printf("Error occurred: %v", err)

	statusCode := http.StatusInternalServerError
	code := "INTERNAL_ERROR"

	// 根据错误类型设置不同的状态码
	if err.Error() == "user not found" {
		statusCode = http.StatusNotFound
		code = "NOT_FOUND"
	}

	c.JSON(statusCode, model.ErrorResponse{
		Code:    code,
		Message: err.Error(),
	})
}

// ==================== 路由设置 ====================

// SetupRouter 设置所有路由
func (h *Handler) RegisterRouter(r gin.IRouter) {

	// 健康检查（无输入输出）
	r.GET("/health", server.WrapGetter(h.Health))

	// 触发任务（无输入输出）
	r.POST("/tasks", server.WrapAction(h.TriggerTask))

	// 用户相关路由
	users := r.Group("/users")
	{
		// 创建用户（有输入输出）
		users.POST("", server.WrapHandler(h.CreateUser))

		// 获取用户（URI 参数）
		users.GET("/:id", server.WrapHandler(h.GetUser))

		// 获取用户列表（Query 参数）
		users.GET("", server.WrapHandler(h.ListUsers))

		// 删除用户（只有输入，无输出，自定义错误处理）
		users.DELETE("/:id", server.WrapConsumer(
			h.DeleteUser,
			server.WithErrorHandler(h.CustomErrorHandler),
		))
	}

	// 文章相关路由
	articles := r.Group("/articles")
	{
		// 更新文章（组合参数：URI + JSON）
		articles.PUT("/:id", server.WrapHandler(h.UpdateArticle))
	}
}
