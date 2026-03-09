package taskpool

import (
	"github.com/nomand-zc/lumin-client/log"
	ants "github.com/panjf2000/ants/v2"
)

const (
	// DefaultPoolSize 默认任务池大小
	DefaultPoolSize = 5
)

// 使用log.Deafult() 作为 ants 的日志输出
type Logger struct{}

// Printf 实现 ants.Logger 接口
func (l Logger) Printf(format string, v ...interface{}) {
	log.Infof(format, v...)
}

// antsPool 基于 ants 库的任务池实现
type antsPool struct {
	pool *ants.Pool
}

func (p *antsPool) Submit(task func()) error {
	return p.pool.Submit(task)
}

func (p *antsPool) Running() int {
	return p.pool.Running()
}

func (p *antsPool) Free() int {
	return p.pool.Free()
}

func (p *antsPool) Cap() int {
	return p.pool.Cap()
}

func (p *antsPool) IsClosed() bool {
	return p.pool.IsClosed()
}

func (p *antsPool) Release() {
	p.pool.Release()
}

var DefaultPool Pool

func init() {
	pool, err := NewPool(DefaultPoolSize)
	if err != nil {
		panic(err)
	}
	DefaultPool = pool
}

// NewAntsPool 创建自定义配置的任务池。
// 适用于需要独立配置（如不同并发数）的场景。
func NewPool(size int, opts ...ants.Option) (Pool, error) {
	defaultOpts := []ants.Option{
		ants.WithPreAlloc(true),
		ants.WithLogger(Logger{}),
		ants.WithPanicHandler(func(i any) {
			log.Errorf("[TaskPool] panic recovered: %v", i)
		}),
	}
	pool, err := ants.NewPool(size, append(defaultOpts, opts...)...)
	if err != nil {
		return nil, err
	}
	return &antsPool{pool: pool}, nil
}
