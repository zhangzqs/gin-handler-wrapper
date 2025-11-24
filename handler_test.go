package ginhandlerwrapper

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// Test types
type TestRequest struct {
	Name  string `json:"name" binding:"required"`
	Email string `json:"email" binding:"required,email"`
}

type TestResponse struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type TestURIRequest struct {
	ID int64 `uri:"id" binding:"required"`
}

type TestQueryRequest struct {
	Page     int `form:"page"`
	PageSize int `form:"page_size"`
}

type TestCombinedRequest struct {
	ID       int64  `uri:"id"`
	Name     string `json:"name"`
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
}

// TestWrapHandler tests the basic WrapHandler functionality
func TestWrapHandler(t *testing.T) {
	r := gin.New()

	r.POST("/users", WrapHandler(
		func(ctx context.Context, req TestRequest) (TestResponse, error) {
			return TestResponse{
				ID:    1,
				Name:  req.Name,
				Email: req.Email,
			}, nil
		},
	))

	t.Run("success", func(t *testing.T) {
		body := `{"name":"Alice","email":"alice@example.com"}`
		req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp TestResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, int64(1), resp.ID)
		assert.Equal(t, "Alice", resp.Name)
		assert.Equal(t, "alice@example.com", resp.Email)
	})

	t.Run("binding_error", func(t *testing.T) {
		body := `{"name":"Alice"}`
		req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("handler_error", func(t *testing.T) {
		r2 := gin.New()
		testErr := errors.New("test error")

		r2.POST("/users", WrapHandler(
			func(ctx context.Context, req TestRequest) (TestResponse, error) {
				return TestResponse{}, testErr
			},
		))

		body := `{"name":"Alice","email":"alice@example.com"}`
		req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r2.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "test error")
	})
}

// TestWrapGetter tests the WrapGetter functionality
func TestWrapGetter(t *testing.T) {
	r := gin.New()

	type HealthResponse struct {
		Status string `json:"status"`
	}

	r.GET("/health", WrapGetter(
		func(ctx context.Context) (HealthResponse, error) {
			return HealthResponse{Status: "ok"}, nil
		},
	))

	t.Run("success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp HealthResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, "ok", resp.Status)
	})

	t.Run("error", func(t *testing.T) {
		r2 := gin.New()
		testErr := errors.New("service unavailable")

		r2.GET("/health", WrapGetter(
			func(ctx context.Context) (HealthResponse, error) {
				return HealthResponse{}, testErr
			},
		))

		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		w := httptest.NewRecorder()

		r2.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

// TestWrapConsumer tests the WrapConsumer functionality
func TestWrapConsumer(t *testing.T) {
	r := gin.New()

	r.DELETE("/users/:id", WrapConsumer(
		func(ctx context.Context, req TestURIRequest) error {
			if req.ID == 0 {
				return errors.New("invalid id")
			}
			return nil
		},
	))

	t.Run("success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/users/123", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("handler_error", func(t *testing.T) {
		r2 := gin.New()

		type SimpleURIRequest struct {
			ID int64 `uri:"id"`
		}

		r2.DELETE("/users/:id", WrapConsumer(
			func(ctx context.Context, req SimpleURIRequest) error {
				if req.ID == 0 {
					return errors.New("invalid id")
				}
				return nil
			},
		))

		req := httptest.NewRequest(http.MethodDelete, "/users/0", nil)
		w := httptest.NewRecorder()

		r2.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "invalid id")
	})
}

// TestWrapAction tests the WrapAction functionality
func TestWrapAction(t *testing.T) {
	r := gin.New()

	executed := false
	r.POST("/tasks", WrapAction(
		func(ctx context.Context) error {
			executed = true
			return nil
		},
	))

	t.Run("success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/tasks", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.True(t, executed)
	})

	t.Run("error", func(t *testing.T) {
		r2 := gin.New()
		testErr := errors.New("task failed")

		r2.POST("/tasks", WrapAction(
			func(ctx context.Context) error {
				return testErr
			},
		))

		req := httptest.NewRequest(http.MethodPost, "/tasks", nil)
		w := httptest.NewRecorder()

		r2.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "task failed")
	})
}

// TestDefaultDecoder tests the default decoder with various binding scenarios
func TestDefaultDecoder(t *testing.T) {
	r := gin.New()

	t.Run("uri_binding", func(t *testing.T) {
		r.GET("/users/:id", WrapHandler(
			func(ctx context.Context, req TestURIRequest) (TestURIRequest, error) {
				return req, nil
			},
		))

		req := httptest.NewRequest(http.MethodGet, "/users/42", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp TestURIRequest
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, int64(42), resp.ID)
	})

	t.Run("query_binding", func(t *testing.T) {
		r2 := gin.New()
		r2.GET("/list", WrapHandler(
			func(ctx context.Context, req TestQueryRequest) (TestQueryRequest, error) {
				return req, nil
			},
		))

		req := httptest.NewRequest(http.MethodGet, "/list?page=2&page_size=10", nil)
		w := httptest.NewRecorder()

		r2.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp TestQueryRequest
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, 2, resp.Page)
		assert.Equal(t, 10, resp.PageSize)
	})

	t.Run("combined_binding", func(t *testing.T) {
		r3 := gin.New()
		r3.POST("/users/:id", WrapHandler(
			func(ctx context.Context, req TestCombinedRequest) (TestCombinedRequest, error) {
				return req, nil
			},
		))

		body := `{"name":"Alice"}`
		req := httptest.NewRequest(http.MethodPost, "/users/123?page=2&page_size=20", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r3.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp TestCombinedRequest
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, int64(123), resp.ID)
		assert.Equal(t, "Alice", resp.Name)
		assert.Equal(t, 2, resp.Page)
		assert.Equal(t, 20, resp.PageSize)
	})
}

// TestCustomDecoder tests custom decoder functionality
func TestCustomDecoder(t *testing.T) {
	r := gin.New()

	customDecoder := func(c *gin.Context) (any, error) {
		return TestRequest{
			Name:  "Custom",
			Email: "custom@example.com",
		}, nil
	}

	r.POST("/users", WrapHandler(
		func(ctx context.Context, req TestRequest) (TestRequest, error) {
			return req, nil
		},
		WithDecoder(customDecoder),
	))

	req := httptest.NewRequest(http.MethodPost, "/users", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp TestRequest
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "Custom", resp.Name)
	assert.Equal(t, "custom@example.com", resp.Email)
}

// TestCustomEncoder tests custom encoder functionality
func TestCustomEncoder(t *testing.T) {
	r := gin.New()

	type CustomResponse struct {
		Code    string      `json:"code"`
		Message string      `json:"message"`
		Data    interface{} `json:"data"`
	}

	customEncoder := func(c *gin.Context, output any) error {
		resp, ok := output.(TestResponse)
		if !ok {
			return errors.New("invalid type")
		}
		c.JSON(http.StatusOK, CustomResponse{
			Code:    "SUCCESS",
			Message: "操作成功",
			Data:    resp,
		})
		return nil
	}

	r.POST("/users", WrapHandler(
		func(ctx context.Context, req TestRequest) (TestResponse, error) {
			return TestResponse{
				ID:    1,
				Name:  req.Name,
				Email: req.Email,
			}, nil
		},
		WithEncoder(customEncoder),
	))

	body := `{"name":"Alice","email":"alice@example.com"}`
	req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp CustomResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "SUCCESS", resp.Code)
	assert.Equal(t, "操作成功", resp.Message)
}

// TestCustomErrorHandler tests custom error handler functionality
func TestCustomErrorHandler(t *testing.T) {
	r := gin.New()

	type ErrorResponse struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	}

	customErrorHandler := func(c *gin.Context, err error) {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "CUSTOM_ERROR",
			Message: err.Error(),
		})
	}

	testErr := errors.New("custom error")

	r.POST("/users", WrapHandler(
		func(ctx context.Context, req TestRequest) (TestResponse, error) {
			return TestResponse{}, testErr
		},
		WithErrorHandler(customErrorHandler),
	))

	body := `{"name":"Alice","email":"alice@example.com"}`
	req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "CUSTOM_ERROR", resp.Code)
	assert.Equal(t, "custom error", resp.Message)
}

// TestMergeOptions tests the mergeOptions functionality
func TestMergeOptions(t *testing.T) {
	customDecoder := func(c *gin.Context) (any, error) {
		return TestRequest{Name: "Custom", Email: "custom@example.com"}, nil
	}

	customEncoder := func(c *gin.Context, output any) error {
		c.JSON(http.StatusCreated, output)
		return nil
	}

	customErrorHandler := func(c *gin.Context, err error) {
		c.JSON(http.StatusBadRequest, gin.H{"custom_error": err.Error()})
	}

	opts := mergeOptions[TestRequest, TestResponse](
		WithDecoder(customDecoder),
		WithEncoder(customEncoder),
		WithErrorHandler(customErrorHandler),
	)

	assert.NotNil(t, opts.decoder)
	assert.NotNil(t, opts.encoder)
	assert.NotNil(t, opts.errorHandler)
}

// BenchmarkWrapHandler benchmarks the WrapHandler function
func BenchmarkWrapHandler(b *testing.B) {
	r := gin.New()

	r.POST("/users", WrapHandler(
		func(ctx context.Context, req TestRequest) (TestResponse, error) {
			return TestResponse{
				ID:    1,
				Name:  req.Name,
				Email: req.Email,
			}, nil
		},
	))

	body := `{"name":"Alice","email":"alice@example.com"}`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)
	}
}
