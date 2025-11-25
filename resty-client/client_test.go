package restyclient

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"resty.dev/v3"
)

// Test types
type TestRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type TestResponse struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type HealthResponse struct {
	Status string `json:"status"`
}

// TestNewClient tests the basic NewClient functionality
func TestNewClient(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var req TestRequest
			json.NewDecoder(r.Body).Decode(&req)

			resp := TestResponse{
				ID:    1,
				Name:  req.Name,
				Email: req.Email,
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := resty.New()
		handler := NewClient[TestRequest, TestResponse](client, "POST", server.URL+"/users")

		result, err := handler(context.Background(), TestRequest{
			Name:  "Alice",
			Email: "alice@example.com",
		})

		assert.NoError(t, err)
		assert.Equal(t, int64(1), result.ID)
		assert.Equal(t, "Alice", result.Name)
		assert.Equal(t, "alice@example.com", result.Email)
	})

	t.Run("http_error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error":"bad request"}`))
		}))
		defer server.Close()

		client := resty.New()
		handler := NewClient[TestRequest, TestResponse](client, "POST", server.URL+"/users")

		_, err := handler(context.Background(), TestRequest{
			Name:  "Alice",
			Email: "alice@example.com",
		})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "400")
	})

	t.Run("json_decode_error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`invalid json`))
		}))
		defer server.Close()

		client := resty.New()
		handler := NewClient[TestRequest, TestResponse](client, "POST", server.URL+"/users")

		_, err := handler(context.Background(), TestRequest{
			Name:  "Alice",
			Email: "alice@example.com",
		})

		assert.Error(t, err)
	})
}

// TestNewGetter tests the NewGetter functionality
func TestNewGetter(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			resp := HealthResponse{Status: "ok"}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := resty.New()
		handler := NewGetter[HealthResponse](client, http.MethodGet, server.URL+"/health")

		result, err := handler(context.Background())

		assert.NoError(t, err)
		assert.Equal(t, "ok", result.Status)
	})

	t.Run("error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusServiceUnavailable)
		}))
		defer server.Close()

		client := resty.New()
		handler := NewGetter[HealthResponse](client, http.MethodGet, server.URL+"/health")

		_, err := handler(context.Background())

		assert.Error(t, err)
	})
}

// TestNewPoster tests the NewPoster functionality
func TestNewPoster(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "POST", r.Method)
			var req TestRequest
			json.NewDecoder(r.Body).Decode(&req)
			assert.Equal(t, "Alice", req.Name)
			assert.Equal(t, "alice@example.com", req.Email)
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client := resty.New()
		handler := NewConsumer[TestRequest](client, http.MethodPost, server.URL+"/users")

		err := handler(context.Background(), TestRequest{
			Name:  "Alice",
			Email: "alice@example.com",
		})

		assert.NoError(t, err)
	})

	t.Run("error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
		}))
		defer server.Close()

		client := resty.New()
		handler := NewConsumer[TestRequest](client, http.MethodPost, server.URL+"/users")

		err := handler(context.Background(), TestRequest{
			Name:  "Alice",
			Email: "alice@example.com",
		})

		assert.Error(t, err)
	})
}

// TestNewPutter tests the NewPutter functionality
func TestNewPutter(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "PUT", r.Method)
			var req TestRequest
			json.NewDecoder(r.Body).Decode(&req)
			assert.Equal(t, "Alice", req.Name)
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client := resty.New()
		handler := NewConsumer[TestRequest](client, http.MethodPut, server.URL+"/users/1")

		err := handler(context.Background(), TestRequest{
			Name:  "Alice",
			Email: "alice@example.com",
		})

		assert.NoError(t, err)
	})
}

// TestNewDeleter tests the NewDeleter functionality
func TestNewDeleter(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "DELETE", r.Method)
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client := resty.New()
		handler := NewAction(client, http.MethodDelete, server.URL+"/users/1")

		err := handler(context.Background())

		assert.NoError(t, err)
	})

	t.Run("error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		client := resty.New()
		handler := NewAction(client, http.MethodDelete, server.URL+"/users/999")

		err := handler(context.Background())

		assert.Error(t, err)
	})
}

// TestNewAction tests the NewAction functionality
func TestNewAction(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		executed := false
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			executed = true
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client := resty.New()
		handler := NewAction(client, "POST", server.URL+"/tasks")

		err := handler(context.Background())

		assert.NoError(t, err)
		assert.True(t, executed)
	})

	t.Run("error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		client := resty.New()
		handler := NewAction(client, "POST", server.URL+"/tasks")

		err := handler(context.Background())

		assert.Error(t, err)
	})
}

// TestCustomEncoder tests custom encoder functionality
func TestCustomEncoder(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 检查自定义请求头
		assert.Equal(t, "custom-value", r.Header.Get("X-Custom-Header"))
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(TestResponse{ID: 1, Name: "Alice", Email: "alice@example.com"})
	}))
	defer server.Close()

	customEncoder := func(req *resty.Request, input any) error {
		req.SetHeader("X-Custom-Header", "custom-value")
		req.SetBody(input)
		return nil
	}

	client := resty.New()
	handler := NewClient[TestRequest, TestResponse](
		client,
		"POST",
		server.URL+"/users",
		WithEncoder(customEncoder),
	)

	result, err := handler(context.Background(), TestRequest{
		Name:  "Alice",
		Email: "alice@example.com",
	})

	assert.NoError(t, err)
	assert.Equal(t, int64(1), result.ID)
}

// TestCustomDecoder tests custom decoder functionality
func TestCustomDecoder(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 返回非标准格式的响应
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"code":"SUCCESS","data":{"id":1,"name":"Custom","email":"custom@example.com"}}`))
	}))
	defer server.Close()

	type WrapperResponse struct {
		Code string       `json:"code"`
		Data TestResponse `json:"data"`
	}

	customDecoder := func(resp *resty.Response) (any, error) {
		var wrapper WrapperResponse
		if err := json.Unmarshal(resp.Bytes(), &wrapper); err != nil {
			return TestResponse{}, err
		}
		return wrapper.Data, nil
	}

	client := resty.New()
	handler := NewClient[TestRequest, TestResponse](
		client,
		"POST",
		server.URL+"/users",
		WithDecoder(customDecoder),
	)

	result, err := handler(context.Background(), TestRequest{
		Name:  "Alice",
		Email: "alice@example.com",
	})

	assert.NoError(t, err)
	assert.Equal(t, int64(1), result.ID)
	assert.Equal(t, "Custom", result.Name)
}

// TestCustomErrorHandler tests custom error handler functionality
func TestCustomErrorHandler(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"code":"VALIDATION_ERROR","message":"invalid input"}`))
	}))
	defer server.Close()

	type ErrorResponse struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	}

	customErrorHandler := func(resp *resty.Response, err error) error {
		if err != nil {
			return err
		}
		if resp.IsError() {
			var errResp ErrorResponse
			json.Unmarshal(resp.Bytes(), &errResp)
			return errors.New(errResp.Code + ": " + errResp.Message)
		}
		return nil
	}

	client := resty.New()
	handler := NewClient[TestRequest, TestResponse](
		client,
		"POST",
		server.URL+"/users",
		WithErrorHandler(customErrorHandler),
	)

	_, err := handler(context.Background(), TestRequest{
		Name:  "Alice",
		Email: "alice@example.com",
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "VALIDATION_ERROR")
	assert.Contains(t, err.Error(), "invalid input")
}

// TestPointerTypes tests if the wrapper supports pointer types
func TestPointerTypes(t *testing.T) {
	t.Run("pointer_input_and_output", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var req TestRequest
			json.NewDecoder(r.Body).Decode(&req)
			resp := TestResponse{
				ID:    1,
				Name:  req.Name,
				Email: req.Email,
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := resty.New()
		handler := NewClient[*TestRequest, *TestResponse](client, "POST", server.URL+"/users")

		result, err := handler(context.Background(), &TestRequest{
			Name:  "Alice",
			Email: "alice@example.com",
		})

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, int64(1), result.ID)
		assert.Equal(t, "Alice", result.Name)
	})

	t.Run("pointer_output_only", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := HealthResponse{Status: "ok"}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := resty.New()
		handler := NewGetter[*HealthResponse](client, http.MethodGet, server.URL+"/health")

		result, err := handler(context.Background())

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "ok", result.Status)
	})

	t.Run("pointer_input_only", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var req TestRequest
			json.NewDecoder(r.Body).Decode(&req)
			assert.Equal(t, "Alice", req.Name)
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client := resty.New()
		handler := NewConsumer[*TestRequest](client, http.MethodPost, server.URL+"/users")

		err := handler(context.Background(), &TestRequest{
			Name:  "Alice",
			Email: "alice@example.com",
		})

		assert.NoError(t, err)
	})
}

// TestMergeOptions tests the mergeOptions functionality
func TestMergeOptions(t *testing.T) {
	customEncoder := func(req *resty.Request, input any) error {
		req.SetBody(input)
		return nil
	}

	customDecoder := func(resp *resty.Response) (any, error) {
		return TestResponse{ID: 1, Name: "Custom", Email: "custom@example.com"}, nil
	}

	customErrorHandler := func(resp *resty.Response, err error) error {
		return errors.New("custom error")
	}

	opts := mergeOptions[TestRequest, TestResponse](
		WithEncoder(customEncoder),
		WithDecoder(customDecoder),
		WithErrorHandler(customErrorHandler),
	)

	assert.NotNil(t, opts.encoder)
	assert.NotNil(t, opts.decoder)
	assert.NotNil(t, opts.errorHandler)
}

// BenchmarkNewClient benchmarks the NewClient function
func BenchmarkNewClient(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req TestRequest
		json.NewDecoder(r.Body).Decode(&req)
		resp := TestResponse{
			ID:    1,
			Name:  req.Name,
			Email: req.Email,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := resty.New()
	handler := NewClient[TestRequest, TestResponse](client, "POST", server.URL+"/users")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		handler(context.Background(), TestRequest{
			Name:  "Alice",
			Email: "alice@example.com",
		})
	}
}
