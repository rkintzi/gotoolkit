package queue

type chunk struct {
	elems []interface{}
	h     int
	t     int
	next  *chunk
}
type Queue struct {
	first *chunk
	last  *chunk
	size  int
}

func New(chunkSize int) *Queue {
	q := &Queue{size: chunkSize}
	q.first = q.alloc()
	return q
}

func (q *Queue) alloc() *chunk {
	ch := &chunk{elems: make([]interface{}, q.size)}
	if q.last != nil {
		q.last.next = ch
	}
	q.last = ch
	return ch
}

func (q *Queue) clear() {
	if q.first.next != nil {
		q.first = q.first.next
	} else {
		q.first.h, q.first.t = 0, 0
	}
}

func (q *Queue) Put(e interface{}) {
	if q.last.t == q.size {
		q.alloc()
	}
	q.last.elems[q.last.t] = e
	q.last.t += 1
}

func (q *Queue) Pop() (e interface{}) {
	if q.first.h == q.first.t {
		panic("Empty queue")
	}
	e, q.first.elems[q.first.h] = q.first.elems[q.first.h], nil
	q.first.h += 1
	if q.first.h == q.first.t {
		q.clear()
	}
	return
}

func (q *Queue) IsEmpty() bool {
	return q.first.h == q.first.t
}
