package queue

import "testing"

const (
	benchIterations = 5000000
	chunkSize       = 500
)

func TestNewIsEmpty(t *testing.T) {
	q := New(chunkSize)
	if !q.IsEmpty() {
		t.Error("New queue should be empty")
	}
	if q.first.next != nil || q.first != q.last {
		t.Error("New queue should have one chunk only")
	}
}

func TestPutChunkSizeElems(t *testing.T) {
	q := New(chunkSize)
	for i := 0; i < chunkSize; i++ {
		q.Put(i)
	}
	if q.first.next != nil || q.first != q.last {
		t.Error("Should have one chunk only")
	}
	var i int
	for i = 0; !q.IsEmpty(); i++ {
		el := q.Pop().(int)
		if el != i {
			t.Error("*Queue is not a fifo")
		}
	}
	if i < chunkSize {
		t.Error("Too few element in queue")
	} else if i > chunkSize {
		t.Error("Too many element in queue")
	}
}

func TestPutTwoTimesChunkSizeElems(t *testing.T) {
	q := New(chunkSize)
	for i := 0; i < 2*chunkSize; i++ {
		q.Put(i)
	}
	if q.first.next != q.last || q.first == q.last || q.first.next.next != nil {
		t.Error("Should have two chunk only")
	}
	var i int
	for i = 0; i < chunkSize; i++ {
		el := q.Pop().(int)
		if el != i {
			t.Error("*Queue is not a fifo")
		}
	}
	if q.first.next != nil || q.first != q.last {
		t.Error("Should have one chunk only")
	}
	for ; !q.IsEmpty(); i++ {
		el := q.Pop().(int)
		if el != i {
			t.Error("Queue is not a fifo")
		}
	}
	if i < 2*chunkSize {
		t.Error("Too few element in queue")
	} else if i > 2*chunkSize {
		t.Error("Too many element in queue")
	}
}

func TestPutOneAndAHalfTimesChunkSizeElems(t *testing.T) {
	N := chunkSize + chunkSize/2
	q := New(chunkSize)
	for i := 0; i < N; i++ {
		q.Put(i)
	}
	if q.first.next != q.last || q.first == q.last || q.first.next.next != nil {
		t.Error("Should have two chunk only")
	}
	var i int
	for i = 0; i < chunkSize; i++ {
		el := q.Pop().(int)
		if el != i {
			t.Error("*Queue is not a fifo")
		}
	}
	if q.first.next != nil || q.first != q.last {
		t.Error("Should have one chunk only")
	}
	for ; !q.IsEmpty(); i++ {
		el := q.Pop().(int)
		if el != i {
			t.Error("*Queue is not a fifo")
		}
	}
	if i < N {
		t.Error("Too few element in queue")
	} else if i > N {
		t.Error("Too many element in queue")
	}
	if q.first.next != nil || q.first != q.last || q.first.h != 0 || q.first.t != 0 {
		t.Error("Should have one empty chunk")
	}
}

func TestPanicOnPopFromEmptyQueue(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Should panic on pop from empty queue")
		}
	}()
	q := New(10)
	q.Pop()
}

func Benchmark_ANRNx1(b *testing.B) {
	b.N = benchIterations
	q := New(chunkSize)
	for i := 0; i < b.N; i += 1 {
		q.Put(i)
	}
	for i := 0; i < b.N; i += 1 {
		q.Pop()
	}
}

func Benchmark_A1R1(b *testing.B) {
	b.N = benchIterations
	q := New(chunkSize)
	for i := 0; i < b.N; i += 1 {
		q.Put(i)
		q.Pop()
	}
}

func Benchmark_A1R2(b *testing.B) {
	b.N = benchIterations
	q := New(chunkSize)
	N2 := b.N / 2
	for i := 0; i < N2; i += 1 {
		q.Put(i)
	}
	for i := 0; i < N2; i += 1 {
		q.Put(i)
		q.Pop()
		q.Pop()
	}
}

func Benchmark_A2R1(b *testing.B) {
	b.N = benchIterations
	q := New(chunkSize)
	N2 := b.N / 2
	for i := 0; i < N2; i += 1 {
		q.Put(i)
		q.Put(i)
		q.Pop()
	}
	for i := 0; i < N2; i += 1 {
		q.Pop()
	}
}
