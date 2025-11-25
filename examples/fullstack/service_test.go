package main

import (
	"context"
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zhangzqs/gin-handler-wrapper/examples/fullstack/apiclient"
	"github.com/zhangzqs/gin-handler-wrapper/examples/fullstack/handler"
	"github.com/zhangzqs/gin-handler-wrapper/examples/fullstack/model"
	"github.com/zhangzqs/gin-handler-wrapper/examples/fullstack/service"
	"github.com/zhangzqs/gin-handler-wrapper/examples/fullstack/serviceimpl"
	"github.com/zhangzqs/gin-handler-wrapper/examples/fullstack/store"
)

// TestServiceInterface 测试Service接口的不同实现
func TestServiceInterface(t *testing.T) {
	// 准备测试数据
	dataStore := store.GetStore()

	// 创建直接调用的服务实现
	directSvc := serviceimpl.NewService(dataStore)

	// 创建HTTP服务器和客户端（RPC实现）
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := handler.NewHandler(directSvc)
	h.RegisterRouter(r)

	server := httptest.NewServer(r)
	defer server.Close()

	rpcClient := apiclient.NewClient(server.URL)

	// 测试两种实现
	testCases := []struct {
		name   string
		svc    service.Service
		isRPC  bool
	}{
		{
			name:   "DirectCall",
			svc:    directSvc,
			isRPC:  false,
		},
		{
			name:   "RPCCall",
			svc:    rpcClient,
			isRPC:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testServiceImplementation(t, tc.svc, tc.isRPC)
		})
	}
}

func testServiceImplementation(t *testing.T, svc service.Service, isRPC bool) {
	ctx := context.Background()

	// 测试健康检查
	t.Run("Health", func(t *testing.T) {
		resp, err := svc.Health(ctx)
		require.NoError(t, err)
		assert.Equal(t, "ok", resp.Status)
		assert.False(t, resp.Timestamp.IsZero())
	})

	// 测试创建用户
	t.Run("CreateUser", func(t *testing.T) {
		req := model.CreateUserRequest{
			Name:  fmt.Sprintf("TestUser_%s", map[bool]string{true: "RPC", false: "Direct"}[isRPC]),
			Email: fmt.Sprintf("test_%s@example.com", map[bool]string{true: "rpc", false: "direct"}[isRPC]),
		}

		user, err := svc.CreateUser(ctx, req)
		require.NoError(t, err)
		assert.NotZero(t, user.ID)
		assert.Equal(t, req.Name, user.Name)
		assert.Equal(t, req.Email, user.Email)
		assert.False(t, user.CreatedAt.IsZero())
	})

	// 测试获取用户
	t.Run("GetUser", func(t *testing.T) {
		// 先创建一个用户
		createReq := model.CreateUserRequest{
			Name:  "GetUserTest",
			Email: "getuser@example.com",
		}
		createdUser, err := svc.CreateUser(ctx, createReq)
		require.NoError(t, err)

		// 获取用户
		getReq := model.GetUserRequest{
			ID: createdUser.ID,
		}
		user, err := svc.GetUser(ctx, getReq)
		require.NoError(t, err)
		assert.Equal(t, createdUser.ID, user.ID)
		assert.Equal(t, createdUser.Name, user.Name)
		assert.Equal(t, createdUser.Email, user.Email)
	})

	// 测试获取用户列表
	t.Run("ListUsers", func(t *testing.T) {
		req := model.ListUsersRequest{
			Page:     1,
			PageSize: 10,
		}

		resp, err := svc.ListUsers(ctx, req)
		require.NoError(t, err)
		assert.NotZero(t, resp.Total)
		assert.NotEmpty(t, resp.Users)
	})

	// 测试更新文章
	t.Run("UpdateArticle", func(t *testing.T) {
		req := model.UpdateArticleRequest{
			ID:      1,
			Title:   "Updated Title",
			Content: "Updated Content",
		}

		article, err := svc.UpdateArticle(ctx, req)
		require.NoError(t, err)
		assert.Equal(t, req.ID, article.ID)
		assert.Equal(t, req.Title, article.Title)
		assert.Equal(t, req.Content, article.Content)
	})

	// 测试删除用户
	t.Run("DeleteUser", func(t *testing.T) {
		// 先创建一个用户
		createReq := model.CreateUserRequest{
			Name:  "DeleteUserTest",
			Email: "deleteuser@example.com",
		}
		createdUser, err := svc.CreateUser(ctx, createReq)
		require.NoError(t, err)

		// 删除用户
		deleteReq := model.DeleteUserRequest{
			ID: createdUser.ID,
		}
		err = svc.DeleteUser(ctx, deleteReq)
		require.NoError(t, err)

		// 验证用户已被删除
		getReq := model.GetUserRequest{
			ID: createdUser.ID,
		}
		_, err = svc.GetUser(ctx, getReq)
		assert.Error(t, err)
	})

	// 测试触发任务
	t.Run("TriggerTask", func(t *testing.T) {
		err := svc.TriggerTask(ctx)
		require.NoError(t, err)
	})
}
