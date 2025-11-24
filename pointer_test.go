package ginhandlerwrapper

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestPointerTypes tests if the wrapper supports pointer types for input/output
func TestPointerTypes(t *testing.T) {
	t.Run("pointer_input_and_output", func(t *testing.T) {
		r := gin.New()

		type PtrRequest struct {
			Name  string `json:"name"`
			Email string `json:"email"`
		}

		type PtrResponse struct {
			ID    int64  `json:"id"`
			Name  string `json:"name"`
			Email string `json:"email"`
		}

		// 使用指针类型
		r.POST("/users", WrapHandler(
			func(ctx context.Context, req *PtrRequest) (*PtrResponse, error) {
				return &PtrResponse{
					ID:    1,
					Name:  req.Name,
					Email: req.Email,
				}, nil
			},
		))

		body := `{"name":"Alice","email":"alice@example.com"}`
		req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp PtrResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, int64(1), resp.ID)
		assert.Equal(t, "Alice", resp.Name)
		assert.Equal(t, "alice@example.com", resp.Email)
	})

	t.Run("pointer_output_only", func(t *testing.T) {
		r := gin.New()

		type User struct {
			ID   int64  `json:"id"`
			Name string `json:"name"`
		}

		r.GET("/user", WrapGetter(
			func(ctx context.Context) (*User, error) {
				return &User{ID: 1, Name: "Alice"}, nil
			},
		))

		req := httptest.NewRequest(http.MethodGet, "/user", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp User
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, int64(1), resp.ID)
		assert.Equal(t, "Alice", resp.Name)
	})

	t.Run("pointer_input_only", func(t *testing.T) {
		r := gin.New()

		type DeleteRequest struct {
			ID int64 `uri:"id"`
		}

		r.DELETE("/users/:id", WrapConsumer(
			func(ctx context.Context, req *DeleteRequest) error {
				assert.NotNil(t, req)
				assert.Equal(t, int64(123), req.ID)
				return nil
			},
		))

		req := httptest.NewRequest(http.MethodDelete, "/users/123", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("nil_pointer_output", func(t *testing.T) {
		r := gin.New()

		type User struct {
			ID   int64  `json:"id"`
			Name string `json:"name"`
		}

		r.GET("/user", WrapGetter(
			func(ctx context.Context) (*User, error) {
				return nil, nil
			},
		))

		req := httptest.NewRequest(http.MethodGet, "/user", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		// nil 指针会被序列化为 "null"
		assert.Equal(t, "null", w.Body.String())
	})
}
