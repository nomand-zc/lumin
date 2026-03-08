package mempool

// Pool 内存池接口，定义了内存池的核心能力。
// 通过接口封装，方便以后低成本替换底层实现库。
type Pool[T any] interface {
	// Get 从内存池中获取一个对象
	Get() T
	// Put 将对象归还到内存池中
	Put(x T)
}
