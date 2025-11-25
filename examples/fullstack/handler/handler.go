package handler

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/zhangzqs/gin-handler-wrapper/examples/fullstack/model"
	"github.com/zhangzqs/gin-handler-wrapper/examples/fullstack/service"
	ginserver "github.com/zhangzqs/gin-handler-wrapper/gin-server"
)

// ==================== Server端HTTP路由注册器 ====================

// Handler 负责将业务服务注册为HTTP路由
type Handler struct {
	svc service.Service
}

// NewHandler 创建新的HTTP路由注册器
func NewHandler(svc service.Service) *Handler {
	return &Handler{
		svc: svc,
	}
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

// RegisterRouter 设置所有路由
// 直接将业务服务的方法注册为HTTP路由，无需额外的包装函数
func (h *Handler) RegisterRouter(r gin.IRouter) {

	// 健康检查（无输入输出）
	r.GET("/health", ginserver.WrapGetter(h.svc.Health))

	// 触发任务（无输入输出）
	r.POST("/tasks", ginserver.WrapAction(h.svc.TriggerTask))

	// 用户相关路由
	users := r.Group("/users")
	{
		// 创建用户（有输入输出）
		users.POST("", ginserver.WrapHandler(h.svc.CreateUser))

		// 获取用户（URI 参数）
		users.GET("/:id", ginserver.WrapHandler(h.svc.GetUser))

		// 获取用户列表（Query 参数）
		users.GET("", ginserver.WrapHandler(h.svc.ListUsers))

		// 删除用户（只有输入，无输出，自定义错误处理）
		users.DELETE("/:id", ginserver.WrapConsumer(
			h.svc.DeleteUser,
			ginserver.WithErrorHandler(h.CustomErrorHandler),
		))
	}

	// 文章相关路由
	articles := r.Group("/articles")
	{
		// 更新文章（组合参数：URI + JSON）
		articles.PUT("/:id", ginserver.WrapHandler(h.svc.UpdateArticle))
	}
}
