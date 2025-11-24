package ginhandlerwrapper

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

type DecoderFunc[I any] func(c *gin.Context) (I, error)

type EncoderFunc[O any] func(c *gin.Context, output O) error

type ErrorHandlerFunc func(c *gin.Context, err error)

type Handler[I, O any] func(ctx context.Context, args I) (O, error)

type WrapHandlerOptions[I, O any] struct {
	decoder      DecoderFunc[I]
	encoder      EncoderFunc[O]
	errorHandler ErrorHandlerFunc
}

type WrapHandlerOptionFunc[I, O any] func(*WrapHandlerOptions[I, O])

func WithDecoder[I, O any](decoder DecoderFunc[I]) WrapHandlerOptionFunc[I, O] {
	return func(opts *WrapHandlerOptions[I, O]) {
		opts.decoder = decoder
	}
}

func WithEncoder[I, O any](encoder EncoderFunc[O]) WrapHandlerOptionFunc[I, O] {
	return func(opts *WrapHandlerOptions[I, O]) {
		opts.encoder = encoder
	}
}

func WithErrorHandler[I, O any](errHandler ErrorHandlerFunc) WrapHandlerOptionFunc[I, O] {
	return func(opts *WrapHandlerOptions[I, O]) {
		opts.errorHandler = errHandler
	}
}

// DefaultDecoder 默认解码器
// 支持多种绑定方式：URI、Query、JSON、Form 等
func DefaultDecoder[I any]() DecoderFunc[I] {
	return func(c *gin.Context) (I, error) {
		var args I

		// 1. 绑定 URI 参数（仅当有 URI 参数时）
		if len(c.Params) > 0 {
			if err := c.ShouldBindUri(&args); err != nil {
				return args, err
			}
		}

		// 2. 根据 Content-Type 绑定请求体
		if c.Request.ContentLength > 0 {
			// 使用 ShouldBind 自动根据 Content-Type 选择绑定方式
			if err := c.ShouldBind(&args); err != nil {
				return args, err
			}
		}

		// 3. 绑定 Query 参数（仅当有 Query 时）
		if len(c.Request.URL.Query()) > 0 {
			if err := c.ShouldBindQuery(&args); err != nil {
				return args, err
			}
		}

		return args, nil
	}
}

// DefaultEncoder 默认编码器
// 自动将响应序列化为 JSON，使用 200 状态码
func DefaultEncoder[O any]() EncoderFunc[O] {
	return func(c *gin.Context, output O) error {
		c.JSON(http.StatusOK, output)
		return nil
	}
}

// DefaultErrorHandler 默认错误处理器
// 所有错误统一返回 500 状态码
func DefaultErrorHandler() ErrorHandlerFunc {
	return func(c *gin.Context, err error) {
		if err == nil {
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

func wrapHandler[I, O any](
	h Handler[I, O],
	decoder DecoderFunc[I],
	encoder EncoderFunc[O],
	errHandler ErrorHandlerFunc,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		args, err := decoder(c)
		if err != nil {
			errHandler(c, err)
			return
		}

		output, err := h(c.Request.Context(), args)
		if err != nil {
			errHandler(c, err)
			return
		}

		if err := encoder(c, output); err != nil {
			errHandler(c, err)
			return
		}
	}
}

func mergeOptions[I, O any](
	options ...WrapHandlerOptionFunc[I, O],
) *WrapHandlerOptions[I, O] {
	opts := WrapHandlerOptions[I, O]{
		decoder:      DefaultDecoder[I](),
		encoder:      DefaultEncoder[O](),
		errorHandler: DefaultErrorHandler(),
	}
	for _, opt := range options {
		opt(&opts)
	}
	return &opts
}

func WrapHandler[I, O any](
	h Handler[I, O],
	options ...WrapHandlerOptionFunc[I, O],
) gin.HandlerFunc {
	opts := mergeOptions(options...)
	return wrapHandler(h, opts.decoder, opts.encoder, opts.errorHandler)
}

// WrapAction 包装无输入输出的处理器
// 适用场景：触发任务、执行操作等不需要请求参数和响应数据的场景
func WrapAction(
	h func(ctx context.Context) error,
	options ...WrapHandlerOptionFunc[struct{}, struct{}],
) gin.HandlerFunc {
	return WrapHandler(func(ctx context.Context, _ struct{}) (struct{}, error) {
		return struct{}{}, h(ctx)
	}, options...)
}

// WrapGetter 包装只有输出的处理器
// 适用场景：获取数据、健康检查等不需要请求参数的查询场景
func WrapGetter[O any](
	h func(ctx context.Context) (O, error),
	options ...WrapHandlerOptionFunc[struct{}, O],
) gin.HandlerFunc {
	return WrapHandler(func(ctx context.Context, _ struct{}) (O, error) {
		return h(ctx)
	}, options...)
}

// WrapConsumer 包装只有输入的处理器
// 适用场景：删除操作、更新操作等不需要返回数据的场景
func WrapConsumer[I any](
	h func(ctx context.Context, args I) error,
	options ...WrapHandlerOptionFunc[I, struct{}],
) gin.HandlerFunc {
	return WrapHandler(func(ctx context.Context, args I) (struct{}, error) {
		return struct{}{}, h(ctx, args)
	}, options...)
}
