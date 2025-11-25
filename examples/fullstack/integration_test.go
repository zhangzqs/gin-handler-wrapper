package main

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zhangzqs/gin-handler-wrapper/examples/fullstack/apiclient"
	"github.com/zhangzqs/gin-handler-wrapper/examples/fullstack/handler"
	"github.com/zhangzqs/gin-handler-wrapper/examples/fullstack/model"
	"github.com/zhangzqs/gin-handler-wrapper/examples/fullstack/serviceimpl"
	"github.com/zhangzqs/gin-handler-wrapper/examples/fullstack/store"
)

// TestIntegration 集成测试：测试完整的HTTP请求-响应流程
func TestIntegration(t *testing.T) {
	// 初始化测试环境
	gin.SetMode(gin.TestMode)
	dataStore := store.GetStore()
	svc := serviceimpl.NewService(dataStore)
	h := handler.NewHandler(svc)

	// 创建测试服务器
	r := gin.New()
	h.RegisterRouter(r)
	server := httptest.NewServer(r)
	defer server.Close()

	// 创建客户端
	client := apiclient.NewClient(server.URL)
	ctx := context.Background()

	// 执行集成测试（类似demo.go的逻辑）
	t.Run("HealthCheck", func(t *testing.T) {
		resp, err := client.Health(ctx)
		require.NoError(t, err)
		assert.Equal(t, "ok", resp.Status)
		assert.False(t, resp.Timestamp.IsZero())
	})

	var createdUserID int64

	t.Run("CreateUser", func(t *testing.T) {
		req := model.CreateUserRequest{
			Name:  "IntegrationTestUser",
			Email: "integration@example.com",
		}

		user, err := client.CreateUser(ctx, req)
		require.NoError(t, err)
		assert.NotZero(t, user.ID)
		assert.Equal(t, req.Name, user.Name)
		assert.Equal(t, req.Email, user.Email)
		assert.False(t, user.CreatedAt.IsZero())

		createdUserID = user.ID
	})

	t.Run("GetUser", func(t *testing.T) {
		req := model.GetUserRequest{ID: createdUserID}

		user, err := client.GetUser(ctx, req)
		require.NoError(t, err)
		assert.Equal(t, createdUserID, user.ID)
		assert.Equal(t, "IntegrationTestUser", user.Name)
		assert.Equal(t, "integration@example.com", user.Email)
	})

	t.Run("ListUsers", func(t *testing.T) {
		req := model.ListUsersRequest{
			Page:     1,
			PageSize: 10,
		}

		resp, err := client.ListUsers(ctx, req)
		require.NoError(t, err)
		assert.NotZero(t, resp.Total)
		assert.NotEmpty(t, resp.Users)

		// 验证我们创建的用户在列表中
		found := false
		for _, u := range resp.Users {
			if u.ID == createdUserID {
				found = true
				break
			}
		}
		assert.True(t, found, "Created user should be in the list")
	})

	t.Run("UpdateArticle", func(t *testing.T) {
		req := model.UpdateArticleRequest{
			ID:      1,
			Title:   "Integration Test Article",
			Content: "This article was updated in integration test",
		}

		article, err := client.UpdateArticle(ctx, req)
		require.NoError(t, err)
		assert.Equal(t, req.ID, article.ID)
		assert.Equal(t, req.Title, article.Title)
		assert.Equal(t, req.Content, article.Content)
	})

	t.Run("DeleteUser", func(t *testing.T) {
		req := model.DeleteUserRequest{ID: createdUserID}

		err := client.DeleteUser(ctx, req)
		require.NoError(t, err)

		// 验证用户已被删除
		getReq := model.GetUserRequest{ID: createdUserID}
		_, err = client.GetUser(ctx, getReq)
		assert.Error(t, err, "Deleted user should not be found")
	})

	t.Run("TriggerTask", func(t *testing.T) {
		err := client.TriggerTask(ctx)
		require.NoError(t, err)
	})
}

// TestIntegrationErrorHandling 测试错误处理
func TestIntegrationErrorHandling(t *testing.T) {
	// 初始化测试环境
	gin.SetMode(gin.TestMode)
	dataStore := store.GetStore()
	svc := serviceimpl.NewService(dataStore)
	h := handler.NewHandler(svc)

	// 创建测试服务器
	r := gin.New()
	h.RegisterRouter(r)
	server := httptest.NewServer(r)
	defer server.Close()

	// 创建客户端
	client := apiclient.NewClient(server.URL)
	ctx := context.Background()

	t.Run("GetNonExistentUser", func(t *testing.T) {
		req := model.GetUserRequest{ID: 999999}

		_, err := client.GetUser(ctx, req)
		assert.Error(t, err)
	})

	t.Run("DeleteNonExistentUser", func(t *testing.T) {
		req := model.DeleteUserRequest{ID: 999999}

		err := client.DeleteUser(ctx, req)
		assert.Error(t, err)
	})

	t.Run("CreateUserWithInvalidEmail", func(t *testing.T) {
		req := model.CreateUserRequest{
			Name:  "TestUser",
			Email: "invalid-email",
		}

		_, err := client.CreateUser(ctx, req)
		// 应该因为邮箱格式不正确而失败
		assert.Error(t, err)
	})
}
