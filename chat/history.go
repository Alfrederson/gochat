package chat

type History[T any] struct {
	buffer []T
	size   int
	head   int
	tail   int
	count  int
}

func NewHistory[T any](size int) *History[T] {
	return &History[T]{
		buffer: make([]T, size),
		size:   size,
		head:   0,
		tail:   0,
		count:  0,
	}
}

func (rb *History[T]) Add(item T) {
	rb.buffer[rb.head] = item
	rb.head = (rb.head + 1) % rb.size
	if rb.count < rb.size {
		rb.count++
	} else {
		rb.tail = (rb.tail + 1) % rb.size
	}
}

func (rb *History[T]) Unroll() []T {
	result := make([]T, rb.count)
	for i := 0; i < rb.count; i++ {
		index := (rb.head + rb.size - rb.count + i) % rb.size
		result[i] = rb.buffer[index]
	}
	return result
}
