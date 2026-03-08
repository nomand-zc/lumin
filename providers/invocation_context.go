package providers

import (
	"context"
)

const (
	// defaultPoolSize 默认任务池大小
	defaultPoolSize = 5
)

// Invocation 调用上下文
type Invocation struct {
	ID    string `json:"id"`
	Model string `json:"model"`
	Usage Usage  `json:"usage"`
}

// Clone 克隆调用上下文
func (inv *Invocation) Clone(opts ...InvocationOption) *Invocation {
	if inv == nil {
		return nil
	}
	newInv := &Invocation{
		ID:    inv.ID,
		Model: inv.Model,
		Usage: inv.Usage,
	}
	return newInv
}

// InvocationOption 调用上下文选项
type InvocationOption func(*Invocation)

// NewInvocation 创建调用上下文
func NewInvocation(opts ...InvocationOption) *Invocation {
	inv := &Invocation{}
	for _, opt := range opts {
		opt(inv)
	}
	return inv
}

type invocationKey struct{}

// EnsureInvocationContext 创建调用上下文
func EnsureInvocationContext(ctx context.Context) (context.Context, *Invocation) {
	if ctx == nil {
		ctx = context.Background()
	}
	inv := GetInvocation(ctx)
	if inv != nil {
		return ctx, inv
	}
	inv = &Invocation{}
	return context.WithValue(ctx, invocationKey{}, inv), inv
}

// GetInvocation 获取调用上下文
func GetInvocation(ctx context.Context) *Invocation {
	if ctx == nil {
		return nil
	}
	if inv, ok := ctx.Value(invocationKey{}).(*Invocation); ok {
		return inv
	}
	return nil
}
