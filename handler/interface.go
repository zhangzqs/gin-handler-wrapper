package handler

import "context"

type HandlerFunc[I, O any] func(ctx context.Context, input I) (O, error)

type ActionHandlerFunc func(ctx context.Context) error

type GetterHandlerFunc[O any] func(ctx context.Context) (O, error)

type ConsumerHandlerFunc[I any] func(ctx context.Context, args I) error
