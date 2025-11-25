package handler

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/zhangzqs/gin-handler-wrapper/examples/fullstack/model"
	"github.com/zhangzqs/gin-handler-wrapper/examples/fullstack/service"
	ginserver "github.com/zhangzqs/gin-handler-wrapper/gin-server"
)

// ==================== 自定义错误处理器 ====================

// customErrorHandler 自定义错误处理器
func customErrorHandler(c *gin.Context, err error) {
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
func RegisterRouter(r gin.IRouter, svc service.Service) {

	// 健康检查（无输入输出）
	r.GET("/health", ginserver.WrapGetter(svc.Health))

	// 触发任务（无输入输出）
	r.POST("/tasks", ginserver.WrapAction(svc.TriggerTask))

	// 用户相关路由
	users := r.Group("/users")
	{
		// 创建用户（有输入输出）
		users.POST("", ginserver.WrapHandler(svc.CreateUser))

		// 获取用户（URI 参数）
		users.GET("/:id", ginserver.WrapHandler(svc.GetUser))

		// 获取用户列表（Query 参数）
		users.GET("", ginserver.WrapHandler(svc.ListUsers))

		// 删除用户（只有输入，无输出，自定义错误处理）
		users.DELETE("/:id", ginserver.WrapConsumer(
			svc.DeleteUser,
			ginserver.WithErrorHandler(customErrorHandler),
		))
	}

	// 文章相关路由
	articles := r.Group("/articles")
	{
		// 更新文章（组合参数：URI + JSON）
		articles.PUT("/:id", ginserver.WrapHandler(svc.UpdateArticle))
	}
}
