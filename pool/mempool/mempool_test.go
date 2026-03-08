package mempool

import (
	"bytes"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewPool_BytesBuffer(t *testing.T) {
	pool := NewPool(func() *bytes.Buffer {
		return new(bytes.Buffer)
	})
	assert.NotNil(t, pool)

	buf := pool.Get()
	assert.NotNil(t, buf)
	assert.Equal(t, 0, buf.Len())
}

func TestNewPool_GetPut(t *testing.T) {
	pool := NewPool(func() *bytes.Buffer {
		return new(bytes.Buffer)
	}, WithResetFunc(func(buf *bytes.Buffer) *bytes.Buffer {
		buf.Reset()
		return buf
	}))

	buf := pool.Get()
	buf.WriteString("hello")
	assert.Equal(t, "hello", buf.String())

	// 归还后重新获取，应被 resetFunc 重置
	pool.Put(buf)
	buf2 := pool.Get()
	assert.Equal(t, 0, buf2.Len())
}

func TestNewPool_WithResetFunc(t *testing.T) {
	type Obj struct {
		Value int
	}

	pool := NewPool(func() *Obj {
		return &Obj{}
	}, WithResetFunc(func(o *Obj) *Obj {
		o.Value = 0
		return o
	}))

	obj := pool.Get()
	obj.Value = 42
	pool.Put(obj)

	obj2 := pool.Get()
	assert.Equal(t, 0, obj2.Value)
}

func TestNewPool_ConcurrentAccess(t *testing.T) {
	pool := NewPool(func() *bytes.Buffer {
		return new(bytes.Buffer)
	}, WithResetFunc(func(buf *bytes.Buffer) *bytes.Buffer {
		buf.Reset()
		return buf
	}))

	var counter atomic.Int64
	var wg sync.WaitGroup

	const goroutineCount = 100
	wg.Add(goroutineCount)

	for i := 0; i < goroutineCount; i++ {
		go func() {
			defer wg.Done()
			buf := pool.Get()
			buf.WriteString("data")
			pool.Put(buf)
			counter.Add(1)
		}()
	}

	wg.Wait()
	assert.Equal(t, int64(goroutineCount), counter.Load())
}

func TestNewPool_InterfaceCompliance(t *testing.T) {
	pool := NewPool(func() []byte {
		return make([]byte, 0, 1024)
	})

	// 验证接口符合性
	var _ Pool[[]byte] = pool

	data := pool.Get()
	assert.NotNil(t, data)
	assert.Equal(t, 0, len(data))
	assert.Equal(t, 1024, cap(data))
}

func BenchmarkMemPool_GetPut(b *testing.B) {
	pool := NewPool(func() *bytes.Buffer {
		return new(bytes.Buffer)
	}, WithResetFunc(func(buf *bytes.Buffer) *bytes.Buffer {
		buf.Reset()
		return buf
	}))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf := pool.Get()
		buf.WriteString("benchmark")
		pool.Put(buf)
	}
}

func BenchmarkDirectAlloc(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf := new(bytes.Buffer)
		buf.WriteString("benchmark")
	}
}
