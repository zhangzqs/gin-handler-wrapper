package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"github.com/zhangzqs/gin-handler-wrapper/client"
	"github.com/zhangzqs/gin-handler-wrapper/server"
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

// ==================== Server端请求/响应类型 ====================

// CreateUserRequest 创建用户请求
type CreateUserRequest struct {
	Name  string `json:"name" binding:"required"`
	Email string `json:"email" binding:"required,email"`
}

// GetUserRequest 获取用户请求（URI参数）
type GetUserRequest struct {
	ID int64 `uri:"id" binding:"required,gt=0"`
}

// ListUsersRequest 获取用户列表请求（Query参数）
type ListUsersRequest struct {
	Page     int `form:"page" binding:"gte=1"`
	PageSize int `form:"page_size" binding:"gte=1,lte=100"`
}

// UpdateArticleRequest 更新文章请求（组合参数）
type UpdateArticleRequest struct {
	ID      int64  `uri:"id" binding:"required,gt=0"`
	Title   string `json:"title" binding:"required"`
	Content string `json:"content"`
}

// DeleteUserRequest 删除用户请求
type DeleteUserRequest struct {
	ID int64 `uri:"id" binding:"required,gt=0"`
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

// ==================== Client端请求类型（支持多种绑定）====================

// ClientGetUserRequest Client获取用户请求（路径参数）
type ClientGetUserRequest struct {
	ID int64 `path:"id"` // 使用 path 标签
}

// ClientListUsersRequest Client列表请求（Query参数）
type ClientListUsersRequest struct {
	Page     int `query:"page"`      // 使用 query 标签
	PageSize int `query:"page_size"` // 使用 query 标签
}

// ClientUpdateArticleRequest Client更新文章请求（组合参数）
type ClientUpdateArticleRequest struct {
	ID      int64  `path:"id"`              // 路径参数
	Token   string `header:"Authorization"` // 请求头
	Title   string `json:"title"`           // JSON body
	Content string `json:"content"`         // JSON body
}

// ClientCreateUserRequest Client创建用户请求（纯JSON body）
type ClientCreateUserRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// ==================== 模拟数据库 ====================

var (
	users    = make(map[int64]User)
	articles = make(map[int64]Article)
	nextID   int64 = 1
)

func init() {
	// 初始化一些测试数据
	users[1] = User{
		ID:        1,
		Name:      "Alice",
		Email:     "alice@example.com",
		CreatedAt: time.Now(),
	}
	users[2] = User{
		ID:        2,
		Name:      "Bob",
		Email:     "bob@example.com",
		CreatedAt: time.Now(),
	}
	nextID = 3
}

// ==================== Server端业务逻辑 ====================

// createUser 创建用户
func createUser(ctx context.Context, req CreateUserRequest) (User, error) {
	user := User{
		ID:        nextID,
		Name:      req.Name,
		Email:     req.Email,
		CreatedAt: time.Now(),
	}
	users[nextID] = user
	nextID++
	return user, nil
}

// getUser 获取用户
func getUser(ctx context.Context, req GetUserRequest) (User, error) {
	user, exists := users[req.ID]
	if !exists {
		return User{}, errors.New("user not found")
	}
	return user, nil
}

// listUsers 获取用户列表
func listUsers(ctx context.Context, req ListUsersRequest) (ListUsersResponse, error) {
	// 设置默认值
	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 10
	}

	// 简单实现分页
	userList := make([]User, 0)
	for _, user := range users {
		userList = append(userList, user)
	}

	return ListUsersResponse{
		Total: len(userList),
		Users: userList,
	}, nil
}

// updateArticle 更新文章
func updateArticle(ctx context.Context, req UpdateArticleRequest) (Article, error) {
	article, exists := articles[req.ID]
	if !exists {
		// 如果不存在就创建
		article = Article{
			ID:     req.ID,
			Author: "unknown",
		}
	}
	article.Title = req.Title
	article.Content = req.Content
	articles[req.ID] = article
	return article, nil
}

// deleteUser 删除用户
func deleteUser(ctx context.Context, req DeleteUserRequest) error {
	_, exists := users[req.ID]
	if !exists {
		return errors.New("user not found")
	}
	delete(users, req.ID)
	return nil
}

// health 健康检查
func health(ctx context.Context) (HealthResponse, error) {
	return HealthResponse{
		Status:    "ok",
		Timestamp: time.Now(),
	}, nil
}

// triggerTask 触发任务
func triggerTask(ctx context.Context) error {
	log.Println("Task triggered successfully")
	return nil
}

// ==================== 自定义错误处理器 ====================

type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func customErrorHandler(c *gin.Context, err error) {
	log.Printf("Error occurred: %v", err)

	statusCode := http.StatusInternalServerError
	code := "INTERNAL_ERROR"

	// 根据错误类型设置不同的状态码
	if err.Error() == "user not found" {
		statusCode = http.StatusNotFound
		code = "NOT_FOUND"
	}

	c.JSON(statusCode, ErrorResponse{
		Code:    code,
		Message: err.Error(),
	})
}

// ==================== Server端路由设置 ====================

func setupServer() *gin.Engine {
	r := gin.Default()

	// 健康检查（无输入输出）
	r.GET("/health", server.WrapGetter(health))

	// 触发任务（无输入输出）
	r.POST("/tasks", server.WrapAction(triggerTask))

	// 用户相关路由
	users := r.Group("/users")
	{
		// 创建用户（有输入输出）
		users.POST("", server.WrapHandler(createUser))

		// 获取用户（URI 参数）
		users.GET("/:id", server.WrapHandler(getUser))

		// 获取用户列表（Query 参数）
		users.GET("", server.WrapHandler(listUsers))

		// 删除用户（只有输入，无输出）
		users.DELETE("/:id", server.WrapConsumer(deleteUser,
			server.WithErrorHandler(customErrorHandler),
		))
	}

	// 文章相关路由
	articles := r.Group("/articles")
	{
		// 更新文章（组合参数：URI + JSON）
		articles.PUT("/:id", server.WrapHandler(updateArticle))
	}

	return r
}

// ==================== Client端示例 ====================

func runClientExamples(baseURL string) {
	fmt.Println("\n========== Client端调用示例 ==========")

	// 创建 resty 客户端
	restyClient := resty.New().SetBaseURL(baseURL)

	// 1. 健康检查（GET，无参数）
	fmt.Println("\n1. 健康检查")
	healthCheck := client.NewGetter[HealthResponse](restyClient, "/health")
	healthResp, err := healthCheck(context.Background())
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Health: %+v\n", healthResp)
	}

	// 2. 创建用户（POST，JSON body）
	fmt.Println("\n2. 创建用户")
	createUserClient := client.NewClient[ClientCreateUserRequest, User](
		restyClient,
		"POST",
		"/users",
	)
	newUser, err := createUserClient(context.Background(), ClientCreateUserRequest{
		Name:  "Charlie",
		Email: "charlie@example.com",
	})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Created user: %+v\n", newUser)
	}

	// 3. 获取用户（GET，路径参数）
	fmt.Println("\n3. 获取用户")
	getUserClient := client.NewClient[ClientGetUserRequest, User](
		restyClient,
		"GET",
		"/users/{id}", // 使用 {id} 作为路径参数占位符
	)
	user, err := getUserClient(context.Background(), ClientGetUserRequest{ID: 1})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Got user: %+v\n", user)
	}

	// 4. 获取用户列表（GET，Query参数）
	fmt.Println("\n4. 获取用户列表")
	listUsersClient := client.NewClient[ClientListUsersRequest, ListUsersResponse](
		restyClient,
		"GET",
		"/users",
	)
	userList, err := listUsersClient(context.Background(), ClientListUsersRequest{
		Page:     1,
		PageSize: 10,
	})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("User list: %+v\n", userList)
	}

	// 5. 更新文章（PUT，组合参数：路径+请求头+Body）
	fmt.Println("\n5. 更新文章")
	updateArticleClient := client.NewClient[ClientUpdateArticleRequest, Article](
		restyClient,
		"PUT",
		"/articles/{id}",
	)
	article, err := updateArticleClient(context.Background(), ClientUpdateArticleRequest{
		ID:      1,
		Token:   "Bearer my-token",
		Title:   "New Article",
		Content: "This is the content",
	})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Updated article: %+v\n", article)
	}

	// 6. 删除用户（DELETE，路径参数）
	fmt.Println("\n6. 删除用户")
	deleteUserClient := client.NewClient[ClientGetUserRequest, struct{}](
		restyClient,
		"DELETE",
		"/users/{id}",
	)
	_, err = deleteUserClient(context.Background(), ClientGetUserRequest{ID: 2})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Println("User deleted successfully")
	}

	// 7. 触发任务（POST，无输入输出）
	fmt.Println("\n7. 触发任务")
	triggerTaskClient := client.NewAction(restyClient, "POST", "/tasks")
	err = triggerTaskClient(context.Background())
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Println("Task triggered successfully")
	}
}

// ==================== 主函数 ====================

func main() {
	// 设置服务器
	r := setupServer()

	// 启动服务器
	port := "8080"
	baseURL := fmt.Sprintf("http://localhost:%s", port)

	go func() {
		log.Printf("Server starting on %s", baseURL)
		if err := r.Run(":" + port); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// 等待服务器启动
	time.Sleep(500 * time.Millisecond)

	// 运行客户端示例
	runClientExamples(baseURL)

	// 保持程序运行一段时间
	fmt.Println("\n按 Ctrl+C 退出...")
	time.Sleep(2 * time.Second)
}
