package pool

import "sync"

// Resetter - интерфейс для объектов, которые могут быть сброшены к начальному состоянию.
type Resetter interface {
	Reset()
}

// Pool - пул объектов с generic-параметром, который ограничен типами с методом Reset().
type Pool[T Resetter] struct {
	pool sync.Pool
	new  func() T
}

// New - создает и возвращает указатель на структуру Pool.
func New[T Resetter](newFunc func() T) *Pool[T] {
	return &Pool[T]{
		pool: sync.Pool{
			New: func() any {
				return newFunc()
			},
		},
		new: newFunc,
	}
}

// Get - возвращает объект из пула.
func (p *Pool[T]) Get() T {
	return p.pool.Get().(T)
}

// Put - помещает объект в пул после вызова его метода Reset()
// Важно: сбрасываем состояние объекта перед возвратом в пул.
func (p *Pool[T]) Put(obj T) {
	obj.Reset()
	p.pool.Put(obj)
}
