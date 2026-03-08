package queue

import "context"

type Consumer[T any] interface {
	Closed() bool
	// Pop 从队列中弹出一个元素，如果队列为空则阻塞等待
	Pop(ctx context.Context) (T, error)
	// Each 阻塞式遍历队列中的所有元素，直到队列关闭或上下文取消。
	// 对每个元素调用 fn，如果 fn 返回 error 则终止遍历并返回该 error。
	Each(ctx context.Context, fn func(T) error) error
	Len() int
}

type Producer[T any] interface {
	Closed() bool
	// Push 向队列中推入一个元素，如果队列已满则阻塞等待
	Push(ctx context.Context, item T) error
}

type Queue[T any] interface {
	Consumer[T]
	Producer[T]
	Closed() bool
	Close() error
}
