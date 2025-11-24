package server

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type DecoderFunc func(c *gin.Context) (any, error)

type EncoderFunc func(c *gin.Context, output any) error

type ErrorHandlerFunc func(c *gin.Context, err error)

// 错误定义
var ErrDecoderReturnedWrongType = errors.New("decoder returned wrong type")

type WrapHandlerOptions struct {
	decoder      DecoderFunc
	encoder      EncoderFunc
	errorHandler ErrorHandlerFunc
}

type WrapHandlerOptionFunc func(*WrapHandlerOptions)

func WithDecoder(decoder DecoderFunc) WrapHandlerOptionFunc {
	return func(opts *WrapHandlerOptions) {
		opts.decoder = decoder
	}
}

func WithEncoder(encoder EncoderFunc) WrapHandlerOptionFunc {
	return func(opts *WrapHandlerOptions) {
		opts.encoder = encoder
	}
}

func WithErrorHandler(errHandler ErrorHandlerFunc) WrapHandlerOptionFunc {
	return func(opts *WrapHandlerOptions) {
		opts.errorHandler = errHandler
	}
}

// DefaultDecoder 默认解码器
// 支持多种绑定方式：URI、Query、JSON、Form 等
func DefaultDecoder[I any]() DecoderFunc {
	return func(c *gin.Context) (any, error) {
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
func DefaultEncoder[O any]() EncoderFunc {
	return func(c *gin.Context, output any) error {
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

func mergeOptions[I, O any](
	options ...WrapHandlerOptionFunc,
) *WrapHandlerOptions {
	opts := WrapHandlerOptions{
		decoder:      DefaultDecoder[I](),
		encoder:      DefaultEncoder[O](),
		errorHandler: DefaultErrorHandler(),
	}
	for _, opt := range options {
		opt(&opts)
	}
	return &opts
}

type Handler[I, O any] func(ctx context.Context, args I) (O, error)

func WrapHandler[I, O any](
	h Handler[I, O],
	options ...WrapHandlerOptionFunc,
) gin.HandlerFunc {
	opts := mergeOptions[I, O](options...)
	decoder := opts.decoder
	encoder := opts.encoder
	errHandler := opts.errorHandler

	return func(c *gin.Context) {
		argAny, err := decoder(c)
		if err != nil {
			errHandler(c, err)
			return
		}

		// 类型断言
		args, ok := argAny.(I)
		if !ok {
			errHandler(c, ErrDecoderReturnedWrongType)
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

type ActionHandler func(ctx context.Context) error

// WrapAction 包装无输入输出的处理器
// 适用场景：触发任务、执行操作等不需要请求参数和响应数据的场景
func WrapAction(
	h ActionHandler,
	options ...WrapHandlerOptionFunc,
) gin.HandlerFunc {
	return WrapHandler(func(ctx context.Context, _ struct{}) (struct{}, error) {
		return struct{}{}, h(ctx)
	}, options...)
}

type GetterHandler[O any] func(ctx context.Context) (O, error)

// WrapGetter 包装只有输出的处理器
// 适用场景：获取数据、健康检查等不需要请求参数的查询场景
func WrapGetter[O any](
	h GetterHandler[O],
	options ...WrapHandlerOptionFunc,
) gin.HandlerFunc {
	return WrapHandler(func(ctx context.Context, _ struct{}) (O, error) {
		return h(ctx)
	}, options...)
}

type ConsumerHandler[I any] func(ctx context.Context, args I) error

// WrapConsumer 包装只有输入的处理器
// 适用场景：删除操作、更新操作等不需要返回数据的场景
func WrapConsumer[I any](
	h ConsumerHandler[I],
	options ...WrapHandlerOptionFunc,
) gin.HandlerFunc {
	return WrapHandler(func(ctx context.Context, args I) (struct{}, error) {
		return struct{}{}, h(ctx, args)
	}, options...)
}
