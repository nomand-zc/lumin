package taskpool

import (
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultPool_Init(t *testing.T) {
	pool := DefaultPool()
	assert.NotNil(t, pool)
	assert.False(t, pool.IsClosed())
}

func TestDefaultPool_Singleton(t *testing.T) {
	pool1 := DefaultPool()
	pool2 := DefaultPool()
	// 接口类型比较，底层实现相同则相等
	assert.Equal(t, pool1, pool2)
}

func TestDefaultPool_SubmitTask(t *testing.T) {
	pool := DefaultPool()

	var counter atomic.Int64
	var wg sync.WaitGroup

	const taskCount = 50
	wg.Add(taskCount)

	for i := 0; i < taskCount; i++ {
		err := pool.Submit(func() {
			defer wg.Done()
			counter.Add(1)
		})
		assert.NoError(t, err)
	}

	wg.Wait()
	assert.Equal(t, int64(taskCount), counter.Load())
}

func TestNewPool_CustomSize(t *testing.T) {
	pool, err := NewPool(5)
	assert.NoError(t, err)
	assert.NotNil(t, pool)
	defer pool.Release()

	var counter atomic.Int64
	var wg sync.WaitGroup

	const taskCount = 20
	wg.Add(taskCount)

	for i := 0; i < taskCount; i++ {
		err := pool.Submit(func() {
			defer wg.Done()
			counter.Add(1)
		})
		assert.NoError(t, err)
	}

	wg.Wait()
	assert.Equal(t, int64(taskCount), counter.Load())
}

func TestPool_InterfaceMethods(t *testing.T) {
	pool, err := NewPool(10)
	assert.NoError(t, err)
	assert.NotNil(t, pool)
	defer pool.Release()

	// 测试 Cap 方法
	assert.Equal(t, 10, pool.Cap())

	// 测试 Free 方法（初始状态下空闲数等于容量）
	assert.Equal(t, 10, pool.Free())

	// 测试 Running 方法（初始无任务运行）
	assert.Equal(t, 0, pool.Running())

	// 测试 IsClosed 方法
	assert.False(t, pool.IsClosed())
}

func BenchmarkDefaultPool_Submit(b *testing.B) {
	pool := DefaultPool()
	var wg sync.WaitGroup

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wg.Add(1)
		_ = pool.Submit(func() {
			wg.Done()
		})
	}
	wg.Wait()
}

func BenchmarkGoroutine_Direct(b *testing.B) {
	var wg sync.WaitGroup

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wg.Add(1)
		go func() {
			wg.Done()
		}()
	}
	wg.Wait()
}
