package mempool

import (
	"sync"
)

// Option 内存池配置选项
type Option[T any] func(*syncPool[T])

// WithResetFunc 设置对象归还时的重置函数。
// 用于在 Put 时清理对象状态，避免脏数据。
func WithResetFunc[T any](fn func(T) T) Option[T] {
	return func(p *syncPool[T]) {
		p.resetFunc = fn
	}
}

// syncPool 基于 sync.Pool 的内存池实现
type syncPool[T any] struct {
	pool      sync.Pool
	resetFunc func(T) T
}

func (p *syncPool[T]) Get() T {
	return p.pool.Get().(T)
}

func (p *syncPool[T]) Put(x T) {
	if p.resetFunc != nil {
		x = p.resetFunc(x)
	}
	p.pool.Put(x)
}

// NewPool 创建自定义配置的内存池。
// newFunc 用于指定对象的创建方式，不能为 nil。
// 适用于需要独立配置的场景。
func NewPool[T any](newFunc func() T, opts ...Option[T]) Pool[T] {
	p := &syncPool[T]{
		pool: sync.Pool{
			New: func() any {
				return newFunc()
			},
		},
	}
	for _, opt := range opts {
		opt(p)
	}
	return p
}
