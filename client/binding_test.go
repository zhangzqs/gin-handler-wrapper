package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
)

// TestPathParams 测试路径参数绑定
func TestPathParams(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 验证路径
		assert.Contains(t, r.URL.Path, "/users/123")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"id": 123, "name": "Alice"})
	}))
	defer server.Close()

	type GetUserRequest struct {
		ID int64 `path:"id"`
	}

	type UserResponse struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
	}

	client := resty.New()
	handler := NewClient[GetUserRequest, UserResponse](
		client,
		"GET",
		server.URL+"/users/{id}",
	)

	result, err := handler(context.Background(), GetUserRequest{ID: 123})

	assert.NoError(t, err)
	assert.Equal(t, int64(123), result.ID)
	assert.Equal(t, "Alice", result.Name)
}

// TestQueryParams 测试查询参数绑定
func TestQueryParams(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 验证查询参数
		assert.Equal(t, "2", r.URL.Query().Get("page"))
		assert.Equal(t, "10", r.URL.Query().Get("page_size"))
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"items": []int{1, 2, 3}})
	}))
	defer server.Close()

	type ListRequest struct {
		Page     int `query:"page"`
		PageSize int `query:"page_size"`
	}

	type ListResponse struct {
		Items []int `json:"items"`
	}

	client := resty.New()
	handler := NewClient[ListRequest, ListResponse](
		client,
		"GET",
		server.URL+"/items",
	)

	result, err := handler(context.Background(), ListRequest{
		Page:     2,
		PageSize: 10,
	})

	assert.NoError(t, err)
	assert.Equal(t, []int{1, 2, 3}, result.Items)
}

// TestFormParams 测试 form 标签作为查询参数
func TestFormParams(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "keyword", r.URL.Query().Get("q"))
		assert.Equal(t, "10", r.URL.Query().Get("limit"))
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"count": 5})
	}))
	defer server.Close()

	type SearchRequest struct {
		Query string `form:"q"`
		Limit int    `form:"limit"`
	}

	type SearchResponse struct {
		Count int `json:"count"`
	}

	client := resty.New()
	handler := NewClient[SearchRequest, SearchResponse](
		client,
		"GET",
		server.URL+"/search",
	)

	result, err := handler(context.Background(), SearchRequest{
		Query: "keyword",
		Limit: 10,
	})

	assert.NoError(t, err)
	assert.Equal(t, 5, result.Count)
}

// TestHeaders 测试请求头绑定
func TestHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 验证请求头
		assert.Equal(t, "Bearer token123", r.Header.Get("Authorization"))
		assert.Equal(t, "custom-value", r.Header.Get("X-Custom-Header"))
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"status": "ok"})
	}))
	defer server.Close()

	type AuthRequest struct {
		Token        string `header:"Authorization"`
		CustomHeader string `header:"X-Custom-Header"`
	}

	type Response struct {
		Status string `json:"status"`
	}

	client := resty.New()
	handler := NewClient[AuthRequest, Response](
		client,
		"GET",
		server.URL+"/protected",
	)

	result, err := handler(context.Background(), AuthRequest{
		Token:        "Bearer token123",
		CustomHeader: "custom-value",
	})

	assert.NoError(t, err)
	assert.Equal(t, "ok", result.Status)
}

// TestJSONBody 测试 JSON 请求体绑定
func TestJSONBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 验证请求体
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		assert.Equal(t, "Alice", body["name"])
		assert.Equal(t, "alice@example.com", body["email"])

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"id":    1,
			"name":  body["name"],
			"email": body["email"],
		})
	}))
	defer server.Close()

	type CreateUserRequest struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	type UserResponse struct {
		ID    int64  `json:"id"`
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	client := resty.New()
	handler := NewClient[CreateUserRequest, UserResponse](
		client,
		"POST",
		server.URL+"/users",
	)

	result, err := handler(context.Background(), CreateUserRequest{
		Name:  "Alice",
		Email: "alice@example.com",
	})

	assert.NoError(t, err)
	assert.Equal(t, int64(1), result.ID)
	assert.Equal(t, "Alice", result.Name)
}

// TestCombinedBinding 测试组合绑定
func TestCombinedBinding(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 验证路径参数
		assert.Contains(t, r.URL.Path, "/users/123")
		// 验证查询参数
		assert.Equal(t, "true", r.URL.Query().Get("verbose"))
		// 验证请求头
		assert.Equal(t, "Bearer token", r.Header.Get("Authorization"))
		// 验证请求体
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		assert.Equal(t, "Bob", body["name"])

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"id":   123,
			"name": body["name"],
		})
	}))
	defer server.Close()

	type UpdateUserRequest struct {
		ID      int64  `path:"id"`
		Verbose bool   `query:"verbose"`
		Token   string `header:"Authorization"`
		Name    string `json:"name"`
	}

	type UserResponse struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
	}

	client := resty.New()
	handler := NewClient[UpdateUserRequest, UserResponse](
		client,
		"PUT",
		server.URL+"/users/{id}",
	)

	result, err := handler(context.Background(), UpdateUserRequest{
		ID:      123,
		Verbose: true,
		Token:   "Bearer token",
		Name:    "Bob",
	})

	assert.NoError(t, err)
	assert.Equal(t, int64(123), result.ID)
	assert.Equal(t, "Bob", result.Name)
}

// TestNoTags 测试无标签结构体（整体作为 body）
func TestNoTags(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body TestRequest
		json.NewDecoder(r.Body).Decode(&body)
		assert.Equal(t, "Alice", body.Name)
		assert.Equal(t, "alice@example.com", body.Email)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(TestResponse{
			ID:    1,
			Name:  body.Name,
			Email: body.Email,
		})
	}))
	defer server.Close()

	client := resty.New()
	handler := NewClient[TestRequest, TestResponse](
		client,
		"POST",
		server.URL+"/users",
	)

	result, err := handler(context.Background(), TestRequest{
		Name:  "Alice",
		Email: "alice@example.com",
	})

	assert.NoError(t, err)
	assert.Equal(t, int64(1), result.ID)
	assert.Equal(t, "Alice", result.Name)
}

// TestPointerWithTags 测试指针类型与标签组合
func TestPointerWithTags(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.Path, "/users/456")
		assert.Equal(t, "Bearer token", r.Header.Get("Authorization"))

		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		assert.Equal(t, "Charlie", body["name"])

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"id":   456,
			"name": body["name"],
		})
	}))
	defer server.Close()

	type UpdateRequest struct {
		ID    int64  `path:"id"`
		Token string `header:"Authorization"`
		Name  string `json:"name"`
	}

	type UserResponse struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
	}

	client := resty.New()
	handler := NewClient[*UpdateRequest, *UserResponse](
		client,
		"PATCH",
		server.URL+"/users/{id}",
	)

	result, err := handler(context.Background(), &UpdateRequest{
		ID:    456,
		Token: "Bearer token",
		Name:  "Charlie",
	})

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, int64(456), result.ID)
	assert.Equal(t, "Charlie", result.Name)
}
