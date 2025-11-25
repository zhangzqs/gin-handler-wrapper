package restyclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/zhangzqs/go-typed-rpc/handler"
	"resty.dev/v3"
)

type RequestEncoderFunc func(req *resty.Request, input any) error

type ResponseDecoderFunc func(resp *resty.Response) (any, error)

type ErrorHandlerFunc func(resp *resty.Response, err error) error

// 错误定义
var ErrEncoderReceivedWrongType = errors.New("encoder received wrong type")
var ErrDecoderReturnedWrongType = errors.New("decoder returned wrong type")

type ClientOptions struct {
	encoder      RequestEncoderFunc
	decoder      ResponseDecoderFunc
	errorHandler ErrorHandlerFunc
}

type ClientOptionFunc func(*ClientOptions)

func WithEncoder(encoder RequestEncoderFunc) ClientOptionFunc {
	return func(opts *ClientOptions) {
		opts.encoder = encoder
	}
}

func WithDecoder(decoder ResponseDecoderFunc) ClientOptionFunc {
	return func(opts *ClientOptions) {
		opts.decoder = decoder
	}
}

func WithErrorHandler(errHandler ErrorHandlerFunc) ClientOptionFunc {
	return func(opts *ClientOptions) {
		opts.errorHandler = errHandler
	}
}

// DefaultRequestEncoder 默认请求编码器
// 智能处理多种请求参数：PathParams、QueryParams、Headers、Body
// 支持标签：
// - path: 路径参数，用于 URL 路径替换
// - query/form: Query 参数
// - header: 请求头
// - json: 请求体（JSON）
func DefaultRequestEncoder[I any]() RequestEncoderFunc {
	return func(req *resty.Request, input any) error {
		if input == nil {
			return nil
		}

		v := reflect.ValueOf(input)
		// 处理指针类型
		if v.Kind() == reflect.Ptr {
			if v.IsNil() {
				return nil
			}
			v = v.Elem()
		}

		// 只处理结构体类型
		if v.Kind() != reflect.Struct {
			// 非结构体直接作为 body
			req.SetBody(input)
			return nil
		}

		t := v.Type()
		pathParams := make(map[string]string)
		queryParams := make(map[string]string)
		headers := make(map[string]string)
		bodyFields := make(map[string]any)
		hasBodyTag := false

		// 遍历所有字段
		for i := 0; i < v.NumField(); i++ {
			field := t.Field(i)
			fieldValue := v.Field(i)

			// 跳过未导出的字段
			if !field.IsExported() {
				continue
			}

			// 获取字段值的字符串表示
			var strValue string
			if fieldValue.Kind() == reflect.Ptr && fieldValue.IsNil() {
				continue // 跳过 nil 指针
			}
			strValue = fmt.Sprintf("%v", fieldValue.Interface())

			// 1. 检查 path 标签
			if pathTag := field.Tag.Get("path"); pathTag != "" {
				pathParams[pathTag] = strValue
				continue
			}

			// 2. 检查 query 或 form 标签
			if queryTag := field.Tag.Get("query"); queryTag != "" {
				queryParams[queryTag] = strValue
				continue
			}
			if formTag := field.Tag.Get("form"); formTag != "" {
				queryParams[formTag] = strValue
				continue
			}

			// 3. 检查 header 标签
			if headerTag := field.Tag.Get("header"); headerTag != "" {
				headers[headerTag] = strValue
				continue
			}

			// 4. 检查 json 标签
			if jsonTag := field.Tag.Get("json"); jsonTag != "" {
				hasBodyTag = true
				// 解析 json 标签（可能包含 omitempty 等选项）
				jsonName := strings.Split(jsonTag, ",")[0]
				if jsonName != "-" {
					bodyFields[jsonName] = fieldValue.Interface()
				}
			}
		}

		// 设置路径参数
		if len(pathParams) > 0 {
			req.SetPathParams(pathParams)
		}

		// 设置查询参数
		if len(queryParams) > 0 {
			req.SetQueryParams(queryParams)
		}

		// 设置请求头
		if len(headers) > 0 {
			req.SetHeaders(headers)
		}

		// 设置请求体
		if hasBodyTag && len(bodyFields) > 0 {
			req.SetBody(bodyFields)
		} else if !hasBodyTag && len(pathParams) == 0 && len(queryParams) == 0 && len(headers) == 0 {
			// 如果没有任何特殊标签，整个对象作为 body
			req.SetBody(input)
		}

		return nil
	}
}

// DefaultResponseDecoder 默认响应解码器
// 自动将响应体反序列化为目标类型
func DefaultResponseDecoder[O any]() ResponseDecoderFunc {
	return func(resp *resty.Response) (any, error) {
		var result O
		// resty v3: resp.Bytes() 替代了 v2 的 resp.Body()
		bodyBytes := resp.Bytes()
		if len(bodyBytes) == 0 {
			// 空响应体，返回零值
			return result, nil
		}
		if err := json.Unmarshal(bodyBytes, &result); err != nil {
			return nil, err
		}
		return result, nil
	}
}

// DefaultErrorHandler 默认错误处理器
// 检查 HTTP 状态码和错误
func DefaultErrorHandler() ErrorHandlerFunc {
	return func(resp *resty.Response, err error) error {
		if err != nil {
			return err
		}
		if resp.IsError() {
			return errors.New(resp.Status())
		}
		return nil
	}
}

func mergeOptions[I, O any](
	options ...ClientOptionFunc,
) *ClientOptions {
	opts := ClientOptions{
		encoder:      DefaultRequestEncoder[I](),
		decoder:      DefaultResponseDecoder[O](),
		errorHandler: DefaultErrorHandler(),
	}
	for _, opt := range options {
		opt(&opts)
	}
	return &opts
}

type Client[I, O any] struct {
	restyClient *resty.Client
	method      string
	url         string
	options     *ClientOptions
}

// NewClient 创建一个通用的 HTTP 客户端
// 支持完全自定义的输入输出类型
func NewClient[I, O any](
	restyClient *resty.Client,
	method string,
	url string,
	options ...ClientOptionFunc,
) handler.HandlerFunc[I, O] {
	opts := mergeOptions[I, O](options...)

	return func(ctx context.Context, input I) (O, error) {
		var zero O

		req := restyClient.R().SetContext(ctx)

		// 编码请求
		if err := opts.encoder(req, input); err != nil {
			return zero, err
		}

		// 发送请求
		resp, err := req.Execute(method, url)

		// 错误处理
		if err := opts.errorHandler(resp, err); err != nil {
			return zero, err
		}

		// 解码响应
		resultAny, err := opts.decoder(resp)
		if err != nil {
			return zero, err
		}

		// 类型断言
		result, ok := resultAny.(O)
		if !ok {
			return zero, ErrDecoderReturnedWrongType
		}

		return result, nil
	}
}

// NewAction 创建无输入输出的客户端处理器
// 适用场景：触发任务、执行操作等不需要请求参数和响应数据的场景
func NewAction(
	restyClient *resty.Client,
	method string,
	url string,
	options ...ClientOptionFunc,
) handler.ActionHandlerFunc {
	handler := NewClient[struct{}, struct{}](restyClient, method, url, options...)
	return func(ctx context.Context) error {
		_, err := handler(ctx, struct{}{})
		return err
	}
}

// NewGetter 创建只有输出的客户端处理器
// 适用场景：获取数据、健康检查等不需要请求参数的查询场景
func NewGetter[O any](
	restyClient *resty.Client,
	method string,
	url string,
	options ...ClientOptionFunc,
) handler.GetterHandlerFunc[O] {
	handler := NewClient[struct{}, O](restyClient, method, url, options...)
	return func(ctx context.Context) (O, error) {
		return handler(ctx, struct{}{})
	}
}

// NewConsumer 创建只有输入的客户端处理器
// 适用场景：删除操作、更新操作等不需要返回数据的场景
func NewConsumer[I any](
	restyClient *resty.Client,
	method string,
	url string,
	options ...ClientOptionFunc,
) handler.ConsumerHandlerFunc[I] {
	handler := NewClient[I, struct{}](restyClient, method, url, options...)
	return func(ctx context.Context, args I) error {
		_, err := handler(ctx, args)
		return err
	}
}
