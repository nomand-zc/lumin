package taskpool

// Pool 任务池接口，定义了任务池的核心能力。
// 通过接口封装，方便以后低成本替换底层实现库。
type Pool interface {
	// Submit 提交一个任务到任务池中执行
	Submit(task func()) error
	// Running 返回当前正在运行的任务数
	Running() int
	// Free 返回当前空闲的 worker 数量
	Free() int
	// Cap 返回任务池容量
	Cap() int
	// IsClosed 返回任务池是否已关闭
	IsClosed() bool
	// Release 释放任务池资源
	Release()
}
